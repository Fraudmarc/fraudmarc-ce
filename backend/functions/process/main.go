package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fraudmarc/fraudmarc-ce/backend/lib"

	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/net/html/charset"
)

// ARBulkInput structure of request parameters
type ARBulkInput struct {
	Params struct {
		RecordStart int    `json:"first_record"`
		MessageID   string `json:"message_id"`
	} `json:"params"`
}

func main() {

	lambda.Start(func(ctx context.Context, event json.RawMessage) (interface{}, error) {
		var q ARBulkInput
		if err := json.Unmarshal(event, &q); err != nil {
			return nil, err
		}

		bucket_name := os.Getenv("BUCKET_NAME")

		svc := s3.New(session.New())

		params := &s3.GetObjectInput{
			Bucket: aws.String(bucket_name),        // Required
			Key:    aws.String(q.Params.MessageID), // Required
		}
		resp, err := svc.GetObject(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Fprintln(os.Stderr, err.Error())
			return nil, err
		}

		// Pretty-print the response data.
		m, err := PrepareAttachment(resp.Body)
		if err != nil {
			panic(err)
		}

		decoder := xml.NewDecoder(m)
		decoder.CharsetReader = charset.NewReaderLabel

		fb := &lib.AggregateReport{}

		if err := decoder.Decode(fb); err != nil {
			panic(err)
		}

		timestamp1, _ := strconv.Atoi(strings.TrimSpace(fb.RawDateRangeBegin))
		fb.DateRangeBegin = int64(timestamp1)
		timestamp2, _ := strconv.Atoi(strings.TrimSpace(fb.RawDateRangeEnd))
		fb.DateRangeEnd = int64(timestamp2)

		fb.MessageId = q.Params.MessageID
		//TODO: Error checking
		if lib.DB == nil || lib.DB.Ping() != nil {
			lib.DBreporting = lib.GetTheRunner("REPORTING")
			lib.DB = lib.DBreporting.DB.DB
		}

		recordStop := q.Params.RecordStart + lib.RecordChunk
		if recordStop > len(fb.Records) {
			recordStop = len(fb.Records)
		}

		chanARR, wg := ParseDmarcARRParallel(50, 4, *fb)

		for i := q.Params.RecordStart; i < recordStop; i++ {
			//fb.Records[i].AggregateReport_id = fb.MessageId
			fb.Records[i].RecordNumber = int64(i)
			chanARR <- &fb.Records[i]
		}

		close(chanARR)
		wg.Wait()

		if recordStop != len(fb.Records) {
			//Recursively call ourselves if there is more work to be done!
			fmt.Println("Recursively calling again for ", q.Params.MessageID)
			ParseDmarcReportBulk(q.Params.MessageID, recordStop)
		}

		return nil, nil
	})

} // main
