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
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/fraudmarc/fraudmarc-ce/backend/lib"
	"github.com/lib/pq"
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

// DmarcARRContext structure of updating ARR table in txn
type DmarcARRContext struct {
	Txn  *sql.Tx
	Stmt *sql.Stmt
	Lock sync.Mutex
}

// AddrNames includes address and its names array
type AddrNames struct {
	Addr  string
	Names []string
}

// AWSSettings sets up the aws comfiguration
type AWSSettings struct {
	Region       *string
	Session      *session.Session
	SvcSNS       *sns.SNS
	SvcLambda    *lambda.Lambda
	SessionReady bool
}

// PrepareAttachment unzip the dmarc report data, and return the decompressed xml
func PrepareAttachment(f io.Reader) (io.Reader, error) {
	//read f and redurns the message, m as type Message
	m, err := mail.ReadMessage(f)
	if err != nil {
		return nil, err
	}

	header := m.Header

	//find media type
	mediaType, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("PrepareAttachment: error parsing media type")
	}

	//if file is multipart
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

	return nil, fmt.Errorf("prepareAttachment: reached the end, no attachment found")
}

// ExtractZipFile extracts first file from zip archive.
func ExtractZipFile(r io.Reader) (io.ReadCloser, error) {

	//create a buffer and copy r to it
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, r)
	if err != nil {
		return nil, err
	}

	//read bytes
	br := bytes.NewReader(buf.Bytes())
	zip, err := zip.NewReader(br, br.Size())
	if err != nil {
		return nil, err
	}

	//check that the file is not empty
	if len(zip.File) == 0 {
		return nil, fmt.Errorf("zip: archive is empty: %s", "unreachable")
	}

	//access the file's contnts
	return zip.File[0].Open()
}

// Open begins a database transaction
func (d *DmarcARRContext) Open() (err error) {
	d.Txn, err = lib.DB.Begin()
	return
}

// Close ends a database transaction
func (d *DmarcARRContext) Close() (err error) {
	if d.Stmt != nil {
		d.Stmt.Close()
	}

	if d.Txn != nil {
		err = d.Txn.Commit()
	}
	return
}

// GetStmtLockIfEmpty returns statement of a DmarcARRContext and locks if statement is empty
func (d *DmarcARRContext) GetStmtLockIfEmpty() *sql.Stmt {

	if d.Stmt != nil {
		return d.Stmt
	}

	d.Lock.Lock()
	if d.Stmt != nil {
		d.Lock.Unlock()
		return d.Stmt
	}

	return nil

}

// ResolveAddrNames returns a struct containging an address and a list of names mapping to it
func ResolveAddrNames(addr string) (addrNames AddrNames, err error) {

	addrNames = AddrNames{
		Addr: addr,
	}

	names, errLookupAddr := net.LookupAddr(addr)
	if len(names) > 0 {
		addrNames.Names = names
	} else if errLookupAddr != nil {
		err = errLookupAddr
	}

	return
}

// ParseDmarcReportBulk invokes the lambda function updating DMR table
func ParseDmarcReportBulk(messageID string, firstRecord int) {
	//prepare aws lambda session for S3 operations
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

// PrepLambdaSession configures a lambda session
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

// ParseDmarcARRParallel uses the parse result of decoding xml and aggregate the information from senderdomain
// and reverse ip look up result to update the dmarc_reporting_entries table
func ParseDmarcARRParallel(lookupPoolSize, dbReadPoolSize int, fb lib.AggregateReport) (chan *lib.AggregateReportRecord, *sync.WaitGroup) {

	var i int

	var inputChan = make(chan *lib.AggregateReportRecord, 2048)
	var initialChan = make(chan *lib.DmarcReportingFull, 2048)
	var lookupOutChan = make(chan *lib.DmarcReportingFull, 2048)
	var dbReadOutChan = make(chan *lib.DmarcReportingFull, 2048)

	//begin transaction on DmarcARRContext
	drCtx := DmarcARRContext{}
	drCtx.Open()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {

		defer wg.Done()

		//parse input
		for arr := range inputChan {
			dr := lib.DmarcReportingFull{
				MessageId:       fb.MessageId,
				Domain:          fb.Domain,
				Policy:          fb.Domain,
				SubdomainPolicy: fb.SubdomainPolicy,
				AlignDKIM:       fb.AlignDKIM,
				AlignSPF:        fb.AlignSPF,
				Pct:             fb.Percentage,
				StartDate:       fb.DateRangeBegin,
				EndDate:         fb.DateRangeEnd,
				RecordNumber:    arr.RecordNumber,
				SourceIP:        arr.SourceIP,
				MessageCount:    arr.Count,
				Disposition:     arr.Disposition,
				EvalDKIM:        arr.EvalDKIM,
				EvalSPF:         arr.EvalSPF,
				HeaderFrom:      arr.HeaderFrom,
				EnvelopeFrom:    arr.EnvelopeFrom,
				EnvelopeTo:      arr.EnvelopeTo,
			}

			for _, dkim := range arr.AuthDKIM {
				dr.AuthDKIMDomain = append(dr.AuthDKIMDomain, dkim.Domain)
				dr.AuthDKIMSelector = append(dr.AuthDKIMSelector, dkim.Selector)
				dr.AuthDKIMResult = append(dr.AuthDKIMResult, dkim.Result)
			}

			for _, spf := range arr.AuthSPF {
				dr.AuthSPFDomain = append(dr.AuthSPFDomain, spf.Domain)
				dr.AuthSPFScope = append(dr.AuthSPFScope, spf.Scope)
				dr.AuthSPFResult = append(dr.AuthSPFResult, spf.Result)
			}

			for _, po := range arr.POReason {
				dr.POReason = append(dr.POReason, po.Reason)
				dr.POComment = append(dr.POComment, po.Comment)
			}

			initialChan <- &dr
		}
		close(initialChan)
	}()

	wg.Add(lookupPoolSize)
	var wgLookup sync.WaitGroup
	wgLookup.Add(lookupPoolSize)
	for i = 0; i < lookupPoolSize; i++ {
		go func() {
			defer wg.Done()
			defer wgLookup.Done()
			//reverse loookup
			for dr := range initialChan {
				addrNames, _ := ResolveAddrNames(dr.SourceIP)
				dr.ReverseLookup = addrNames.Names
				lookupOutChan <- dr
			}

		}()
	}

	wg.Add(dbReadPoolSize)
	var wgdbRead sync.WaitGroup
	wgdbRead.Add(dbReadPoolSize)
	for i = 0; i < dbReadPoolSize; i++ {
		go func() {
			defer wg.Done()
			defer wgdbRead.Done()

			var ipsbStmt *sql.Stmt
			ipMap := make(map[string][]*lib.DmarcReportingFull)
			bufferedRecords := []*lib.DmarcReportingFull{}
			var item *lib.DmarcReportingFull
			var output *lib.DmarcReportingFull
			for item = range lookupOutChan {
				ipMap[item.SourceIP] = append(ipMap[item.SourceIP], item)
				bufferedRecords = append(bufferedRecords, item)
				if len(ipMap) == 10 {
					ipIntelLookup10FillDRF(ipMap, &ipsbStmt)

					for _, output = range bufferedRecords {
						dbReadOutChan <- output
					}
					bufferedRecords = []*lib.DmarcReportingFull{}
					ipMap = make(map[string][]*lib.DmarcReportingFull)
				}

			}

			if len(ipMap) > 0 {
				ipIntelLookup10FillDRF(ipMap, &ipsbStmt)

				for _, output = range bufferedRecords {
					dbReadOutChan <- output
				}
			}

			if ipsbStmt != nil {
				ipsbStmt.Close()
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		for dr := range dbReadOutChan {
			dr.LastUpdate = time.Now().Format("2006-01-02 15:04:05.000000000")

			writeToDmarcReportingFull(*dr, drCtx.Txn, &drCtx)
		}
		drCtx.Close()
	}()

	go func() {
		wgLookup.Wait()
		close(lookupOutChan)
		wgdbRead.Wait()
		close(dbReadOutChan)
	}()

	return inputChan, &wg

}

func ipIntelLookup10FillDRF(ipMap map[string][]*lib.DmarcReportingFull, ipsbStmt **sql.Stmt) {
	var ip string
	ipList := []string{}
	var ipsbList []lib.SBGeo
	for ip = range ipMap {
		ipList = append(ipList, ip)
		sb, err := SenderbaseIPData(ip)
		if err != nil {
			return
		}
		ipsbList = append(ipsbList, sb)
	}

	var curDrf *lib.DmarcReportingFull

	for i, curIntel := range ipsbList {
		for _, curDrf = range ipMap[ipList[i]] {
			curDrf.OrgName = curIntel.OrgName
			curDrf.OrgId = curIntel.OrgID
			curDrf.HostName = curIntel.Hostname
			curDrf.DomainName = curIntel.DomainName
			curDrf.HostnameMatchesIP = curIntel.HostnameMatchesIP
			curDrf.City = curIntel.City
			curDrf.State = curIntel.State
			curDrf.Country = curIntel.Country
			curDrf.Longitude = curIntel.Longitude
			curDrf.Latitude = curIntel.Latitude
		}
	}

	return
}

// SenderbaseIPData query the senderbase to find out the org name of ip
func SenderbaseIPData(sip string) (sbGeo lib.SBGeo, err error) {

	// convert from string input to net.IP:
	ip := net.ParseIP(sip).To4()
	if ip == nil {
		log.Println("ip6 address")
		return
	}

	// reverse the byte-order of IP:
	srevip := byteReverseIP4(ip)

	// senderbase ip-specific domain to query:
	domain := fmt.Sprintf("%s.query.senderbase.org", srevip.String)

	// perform the lookup:
	log.Println("SB:  lookupTXT  ", sip)
	txtRecords, errLookupTXT := net.LookupTXT(domain)
	log.Println("SB:  lookupTXT2  ", sip)
	if errLookupTXT != nil {
		err = fmt.Errorf("SBIPD - errLookupTXT:  %s\n%s", domain, errLookupTXT)
		log.Println(err)
		return
	}
	if len(txtRecords) < 1 {
		err = fmt.Errorf("no TXT records found for IP %s\n%s", domain, sip)
		log.Println(err)
		return
	}

	rr := txtRecords[0]
	log.Println("SB:  TXT proc  ", rr)

	// handle multiple TXT records:
	sbStr := rr
	if len(txtRecords) > 1 {
		sort.Slice(txtRecords, func(i, j int) bool { return txtRecords[i][0] < txtRecords[j][0] })
		sbStr = ""
		for j := range txtRecords {
			// each TXT leads with '[0-9]-'...strip away this 2-char prefix
			txtRecords[j] = txtRecords[j][2:len(txtRecords[j])] // this could use regex improvement
			sbStr = fmt.Sprintf("%s%s", sbStr, txtRecords[j])
		}
	}
	sbFields := strings.Split(sbStr, "|")
	sbMap := map[string]string{}
	for j := range sbFields {
		sbm := strings.Split(sbFields[j], "=")
		sbMap[sbm[0]] = sbm[1]
	}

	log.Println("SB:  struct  ", sip)
	sbGeo.OrgName = sbMap["1"]
	sbGeo.OrgID = sbMap["4"]
	sbGeo.OrgCategory = sbMap["5"]
	sbGeo.Hostname = strings.ToLower(sbMap["20"])
	sbGeo.DomainName = strings.ToLower(sbMap["21"])
	sbGeo.HostnameMatchesIP = sbMap["22"]
	sbGeo.City = sbMap["50"]
	sbGeo.State = sbMap["51"]
	sbGeo.Country = sbMap["53"]
	sbGeo.Longitude = sbMap["54"]
	sbGeo.Latitude = sbMap["55"]

	return
}

// ByteReverseIP4 translates an IP4 into reverse byte order
func byteReverseIP4(ip net.IP) (revip revIP4) {

	for j := 0; j < len(ip); j++ {
		revip.Byte[len(ip)-j-1] = ip[j]
		revip.String = fmt.Sprintf("%d.%s", ip[j], revip.String)
	}

	revip.String = strings.TrimRight(revip.String, ".")

	return
}

func writeToDmarcReportingFull(dr lib.DmarcReportingFull, txn *sql.Tx, ctx *DmarcARRContext) {

	log.Println("writeToDmarcReportingFull - start", dr.SourceIP)
	var stmt *sql.Stmt
	var errPrepare error

	if ctx == nil || ctx.GetStmtLockIfEmpty() == nil { //Prepare a statement if we have no repeat context, or it's empty
		stmt, errPrepare = txn.Prepare(pq.CopyInSchema("public", lib.DreTable,
			"message_id",
			"policy",
			"subdomain_policy",
			"align_dkim",
			"align_spf",
			"pct",
			"record_number",
			"domain",
			"source_ip",
			"esp",
			"org_name",
			"org_id",
			"host_name",
			"domain_name",
			"host_name_matches_ip",
			"city",
			"state",
			"country",
			"longitude",
			"latitude",
			"reverse_lookup",
			"message_count",
			"disposition",
			"eval_dkim",
			"eval_spf",
			"header_from",
			"envelope_from",
			"envelope_to",
			"auth_dkim_domain",
			"auth_dkim_selector",
			"auth_dkim_result",
			"auth_spf_domain",
			"auth_spf_scope",
			"auth_spf_result",
			"po_reason",
			"po_comment",
			"start_date",
			"end_date",
			"last_update",
		))
		if errPrepare != nil {
			log.Println("dmarcReport.writeToDmarcReportingFull - errPrepare:  ", errPrepare)
			return
		}
		if ctx != nil { //Store for next time, if we can
			ctx.Stmt = stmt
			ctx.Lock.Unlock()
		}
	} else {
		stmt = ctx.Stmt //Use the existing prepared statement
	}

	var err error
	item := dr
	_, err = stmt.Exec(
		item.MessageId,
		item.Policy,
		item.SubdomainPolicy,
		item.AlignDKIM,
		item.AlignSPF,
		item.Pct,
		item.RecordNumber,
		item.Domain,
		item.SourceIP,
		item.ESP,
		item.OrgName,
		item.OrgId,
		item.HostName,
		item.DomainName,
		item.HostnameMatchesIP,
		item.City,
		item.State,
		item.Country,
		item.Longitude,
		item.Latitude,
		pq.Array(item.ReverseLookup),
		item.MessageCount,
		item.Disposition,
		item.EvalDKIM,
		item.EvalSPF,
		item.HeaderFrom,
		item.EnvelopeFrom,
		item.EnvelopeTo,
		pq.Array(item.AuthDKIMDomain),
		pq.Array(item.AuthDKIMSelector),
		pq.Array(item.AuthDKIMResult),
		pq.Array(item.AuthSPFDomain),
		pq.Array(item.AuthSPFScope),
		pq.Array(item.AuthSPFResult),
		pq.Array(item.POReason),
		pq.Array(item.POComment),
		item.StartDate,
		item.EndDate,
		item.LastUpdate,
	) // stmt.Exec
	if err != nil {
		log.Println("dmarcReport.writeToDmarcReportingFull - err:  ", err)
	}

	if ctx == nil { // Close the statement if it will not be used again
		errClose := stmt.Close()
		if errClose != nil {
			log.Println("dmarcReport.writeToDmarcReportingFull - errClose:  ", errClose)
			return
		}
	}
	log.Println("writeToDmarcReportingFull - end", dr.SourceIP)

} // writeToDmarcReportingFull
