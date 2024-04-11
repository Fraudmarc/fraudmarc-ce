// 2016/11/04 - speed - insert in transaction batches
// 2016/11/06 - speed - http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_BestPractices.html
//                      sync commit off on 2016/11/06

package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/fraudmarc/fraudmarc-ce/backend/lib"
	"github.com/lib/pq"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding/charmap"
)

const (
	chanSize    = 5000
	dbChunkRows = 5000
	dbBatchMs   = 250
)

var (
	arnLambdaDmarcARResolveBulk = os.Getenv("ArnLambdaDmarcARResolveBulk")
)

type revIP4 struct {
	Byte   [4]byte
	String string
}

type AWSSettings struct {
	Region       *string
	Session      *session.Session
	SvcSNS       *sns.SNS
	SvcLambda    *lambda.Lambda
	SessionReady bool
}

type ARBulkInput struct {
	Params struct {
		RecordStart int    `json:"first_record"`
		MessageID   string `json:"message_id"`
	} `json:"params"`
}

type AddrNames struct {
	Addr  string
	Names []string
}

// DmarcReportPrepareAttachment unzip the dmarc report data, and return the decompressed xml
func DmarcReportPrepareAttachment(f io.Reader) (io.Reader, error) {

	m, err := mail.ReadMessage(f)
	if err != nil {
		return nil, err
	}

	header := m.Header

	mediaType, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("PrepareAttachment: error parsing media type")
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(m.Body, params["boundary"])

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return nil, fmt.Errorf("PrepareAttachment: EOF before valid attachment")
			}
			if err != nil {
				return nil, err
			}

			// need to add checks to ensure base64
			partType, _, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
			if err != nil {
				return nil, fmt.Errorf("PrepareAttachment: error parsing media type of part")
			}

			// if gzip
			if strings.HasPrefix(partType, "application/gzip") ||
				strings.HasPrefix(partType, "application/x-gzip") ||
				strings.HasPrefix(partType, "application/gzip-compressed") ||
				strings.HasPrefix(partType, "application/gzipped") ||
				strings.HasPrefix(partType, "application/x-gunzip") ||
				strings.HasPrefix(partType, "application/x-gzip-compressed") ||
				strings.HasPrefix(partType, "gzip/document") {

				decodedBase64 := base64.NewDecoder(base64.StdEncoding, p)
				decompressed, err := gzip.NewReader(decodedBase64)
				if err != nil {
					return nil, err
				}

				return decompressed, nil
			}

			// if zip
			if strings.HasPrefix(partType, "application/zip") || // google style
				strings.HasPrefix(partType, "application/x-zip-compressed") { // yahoo style

				decodedBase64 := base64.NewDecoder(base64.StdEncoding, p)
				decompressed, err := ExtractZipFile(decodedBase64)
				if err != nil {
					return nil, err
				}

				return decompressed, nil
			}

			// if xml
			if strings.HasPrefix(partType, "text/xml") {
				return p, nil
			}

			// if application/octetstream, check filename
			if strings.HasPrefix(partType, "application/octet-stream") {

				if strings.HasSuffix(p.FileName(), ".zip") {
					decodedBase64 := base64.NewDecoder(base64.StdEncoding, p)
					decompressed, err := ExtractZipFile(decodedBase64)
					if err != nil {
						return nil, err
					}

					return decompressed, nil
				}

				if strings.HasSuffix(p.FileName(), ".gz") {
					decodedBase64 := base64.NewDecoder(base64.StdEncoding, p)
					decompressed, _ := gzip.NewReader(decodedBase64)

					return decompressed, nil
				}
			}
		}

	}

	// if gzip
	if strings.HasPrefix(mediaType, "application/gzip") || // proper :)
		strings.HasPrefix(mediaType, "application/x-gzip") || // gmail attachment
		strings.HasPrefix(mediaType, "application/gzip-compressed") ||
		strings.HasPrefix(mediaType, "application/gzipped") ||
		strings.HasPrefix(mediaType, "application/x-gunzip") ||
		strings.HasPrefix(mediaType, "application/x-gzip-compressed") ||
		strings.HasPrefix(mediaType, "gzip/document") {

		decodedBase64 := base64.NewDecoder(base64.StdEncoding, m.Body)
		decompressed, _ := gzip.NewReader(decodedBase64)

		return decompressed, nil

	}

	// if zip
	if strings.HasPrefix(mediaType, "application/zip") || // google style
		strings.HasPrefix(mediaType, "application/x-zip-compressed") { // yahoo style
		decodedBase64 := base64.NewDecoder(base64.StdEncoding, m.Body)
		decompressed, err := ExtractZipFile(decodedBase64)
		if err != nil {
			return nil, err
		}

		return decompressed, nil
	}

	// if xml
	if strings.HasPrefix(mediaType, "text/xml") {
		return m.Body, nil
	}

	return nil, fmt.Errorf("PrepareAttachment: reached the end, no attachment found.")
}

// parse: decode the dmarc report xml file and update five tables in five go routines
func parse(r io.Reader, messageId string) {

	var wg sync.WaitGroup

	var chanAggregateReport = make(chan lib.AggregateReport, chanSize)
	var chanAggregateReportRecord = make(chan lib.AggregateReportRecord, chanSize)
	var chanDkimAuthResult = make(chan lib.DKIMAuthResult, chanSize)
	var chanSpfAuthResult = make(chan lib.SPFAuthResult, chanSize)
	var chanPoReason = make(chan lib.POReason, chanSize)

	fb := &lib.AggregateReport{}

	// functional in some cases
	// http://stackoverflow.com/questions/34712015/unmarshal-multiple-xml-items

	// https://groups.google.com/forum/#!topic/golang-nuts/FHmzYmM5r5Y
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(fb); err != nil {
		panic(err)
	}

	timestamp1, _ := strconv.Atoi(strings.TrimSpace(fb.RawDateRangeBegin))
	fb.DateRangeBegin = int64(timestamp1)
	timestamp2, _ := strconv.Atoi(strings.TrimSpace(fb.RawDateRangeEnd))
	fb.DateRangeEnd = int64(timestamp2)

	fb.MessageId = messageId

	txnAR, _ := lib.DB.Begin()
	txnARR, _ := lib.DB.Begin()
	txnDKIM, _ := lib.DB.Begin()
	txnSPF, _ := lib.DB.Begin()
	txnPO, _ := lib.DB.Begin()

	wg.Add(5)
	go writeToAggregateReport(chanAggregateReport, &wg, txnAR)
	go writeToAggregateReportRecord(chanAggregateReportRecord, &wg, txnARR)
	go writeToDkimAuthResult(chanDkimAuthResult, &wg, txnDKIM)
	go writeToSpfAuthResult(chanSpfAuthResult, &wg, txnSPF)
	go writeToPoReason(chanPoReason, &wg, txnPO)

	chanAggregateReport <- *fb

	//-------------------------------------------------------------------------------------
	// harvest-dmarc-ips:
	//-------------------------------------------------------------------------------------
	// loop through each record to aggregate all IPs from this report:
	ipList := []string{}
	for _, rec := range fb.Records {
		ipList = append(ipList, rec.SourceIP)
	}

	//-------------------------------------------------------------------------------------
	// process each record:
	//-------------------------------------------------------------------------------------
	for num, rec := range fb.Records {

		rec.AggregateReport_id = messageId
		rec.RecordNumber = int64(num)

		// Push the data into this COPY batch
		chanAggregateReportRecord <- rec

		for _, dkim := range rec.AuthDKIM {
			dkim.AggregateReport_id = rec.AggregateReport_id
			dkim.RecordNumber = rec.RecordNumber
			chanDkimAuthResult <- dkim
		}

		for _, spf := range rec.AuthSPF {
			spf.AggregateReport_id = rec.AggregateReport_id
			spf.RecordNumber = rec.RecordNumber
			chanSpfAuthResult <- spf
		}

		for _, reason := range rec.POReason {
			reason.AggregateReport_id = rec.AggregateReport_id
			reason.RecordNumber = rec.RecordNumber
			chanPoReason <- reason
		}

	}

	close(chanAggregateReport)
	close(chanAggregateReportRecord)
	close(chanDkimAuthResult)
	close(chanSpfAuthResult)
	close(chanPoReason)

	wg.Wait()

	err := txnAR.Commit()
	if err != nil {
		log.Fatal(err)
	}
	err2 := txnARR.Commit()
	if err2 != nil {
		log.Fatal(err2)
	}
	err3 := txnDKIM.Commit()
	if err3 != nil {
		log.Fatal(err3)
	}
	err4 := txnSPF.Commit()
	if err4 != nil {
		log.Fatal(err4)
	}
	err5 := txnPO.Commit()
	if err5 != nil {
		log.Fatal(err5)
	}

	ParseDmarcReportBulk(messageId, 0)

	return

} // parse

// ExtractFile extracts first file from zip archive
func ExtractZipFile(r io.Reader) (io.ReadCloser, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	br := bytes.NewReader(buf.Bytes())
	zip, err := zip.NewReader(br, br.Size())
	if err != nil {
		return nil, err
	}

	if len(zip.File) == 0 {
		return nil, fmt.Errorf("zip: archive is empty: %s", "unreachable")
	}

	return zip.File[0].Open()
}

// http://stackoverflow.com/questions/34712015/unmarshal-multiple-xml-items
func makeCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" {
		// Windows-1252 is a superset of ISO-8859-1, so should do here
		return charmap.Windows1252.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("Unknown charset: %s", charset)
}

func writeToAggregateReport(rx chan lib.AggregateReport, waitgroup *sync.WaitGroup, txn *sql.Tx) {
	defer waitgroup.Done()

	stmt, err := txn.Prepare(pq.CopyInSchema("public", lib.ARTable,
		"MessageId",
		"Organization",
		"Email",
		"ExtraContact",
		"ReportID",
		"DateRangeBegin",
		"DateRangeEnd",
		"Domain",
		"AlignDKIM",
		"AlignSPF",
		"Policy",
		"SubdomainPolicy",
		"Percentage",
		"FailureReport"))
	if err != nil {
		log.Fatal(err)
	}

	for item := range rx {
		_, err = stmt.Exec(item.MessageId,
			item.Organization,
			item.Email,
			item.ExtraContact,
			item.ReportID,
			item.DateRangeBegin,
			item.DateRangeEnd,
			item.Domain,
			item.AlignDKIM,
			item.AlignSPF,
			item.Policy,
			item.SubdomainPolicy,
			item.Percentage,
			item.FailureReport)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

}

func writeToAggregateReportRecord(rx chan lib.AggregateReportRecord, waitgroup *sync.WaitGroup, txn *sql.Tx) {
	defer waitgroup.Done()

	stmt, err := txn.Prepare(pq.CopyInSchema("public", lib.ARRTable,
		"AggregateReport_id",
		"RecordNumber",
		"SourceIP",
		"Count",
		"Disposition",
		"EvalDKIM",
		"EvalSPF",
		"HeaderFrom",
		"EnvelopeFrom",
		"EnvelopeTo"))
	if err != nil {
		log.Fatal(err)
	}

	for item := range rx {
		_, err = stmt.Exec(item.AggregateReport_id,
			item.RecordNumber,
			item.SourceIP,
			item.Count,
			item.Disposition,
			item.EvalDKIM,
			item.EvalSPF,
			item.HeaderFrom,
			item.EnvelopeFrom,
			item.EnvelopeTo)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}

}

func writeToDkimAuthResult(rx chan lib.DKIMAuthResult, waitgroup *sync.WaitGroup, txn *sql.Tx) {
	defer waitgroup.Done()

	stmt, err := txn.Prepare(pq.CopyInSchema("public", "DkimAuthResult",
		"AggregateReport_id",
		"RecordNumber",
		"Domain",
		"Selector",
		"Result",
		"HumanResult"))
	if err != nil {
		log.Fatal(err)
	}

	for item := range rx {
		_, err = stmt.Exec(item.AggregateReport_id,
			item.RecordNumber,
			item.Domain,
			item.Selector,
			item.Result,
			item.HumanResult)
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func writeToSpfAuthResult(rx chan lib.SPFAuthResult, waitgroup *sync.WaitGroup, txn *sql.Tx) {
	defer waitgroup.Done()

	stmt, err := txn.Prepare(pq.CopyInSchema("public", "SpfAuthResult",
		"AggregateReport_id",
		"RecordNumber",
		"Domain",
		"Scope",
		"Result"))
	if err != nil {
		log.Fatal(err)
	}

	for item := range rx {
		_, err = stmt.Exec(item.AggregateReport_id,
			item.RecordNumber,
			item.Domain,
			item.Scope,
			item.Result)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func writeToPoReason(rx chan lib.POReason, waitgroup *sync.WaitGroup, txn *sql.Tx) {
	defer waitgroup.Done()

	stmt, err := txn.Prepare(pq.CopyInSchema("public", "PoReason",
		"AggregateReport_id",
		"RecordNumber",
		"Reason",
		"Comment"))
	if err != nil {
		log.Fatal(err)
	}

	for item := range rx {
		_, err = stmt.Exec(item.AggregateReport_id,
			item.RecordNumber,
			item.Reason,
			item.Comment)
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// ParseDmarcReportBulk invokes the lambda function used to update dmarc_reporting_entries table
func ParseDmarcReportBulk(messageID string, firstRecord int) {

	prepLambdaSession()

	input := ARBulkInput{}
	input.Params.MessageID = messageID
	input.Params.RecordStart = firstRecord

	buf, _ := json.Marshal(input)

	// spin this off as separate lambda:
	invokeInput := lambda.InvokeInput{
		FunctionName:   aws.String(arnLambdaDmarcARResolveBulk),
		InvocationType: aws.String("Event"),
		Payload:        buf,
	}

	_, errInvoke := lib.SvcLambda.Invoke(&invokeInput)
	if errInvoke != nil {
		log.Println("Failed to invoke DRE update bulk lambda: ", errInvoke)
	}

}

func prepLambdaSession() AWSSettings {

	// new AWS session for S3 operations:
	lib.AwsConfig.Region = aws.String(lib.LambdaRegion)
	lib.Sess = session.New(&lib.AwsConfig)
	lib.SvcLambda = lambda.New(lib.Sess)
	lib.SessionReady = true

	return AWSSettings{
		Region:       lib.AwsConfig.Region,
		Session:      lib.Sess,
		SvcLambda:    lib.SvcLambda,
		SessionReady: lib.SessionReady,
	}

}
