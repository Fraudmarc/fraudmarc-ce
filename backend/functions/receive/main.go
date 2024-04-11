package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

// from http://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email-notifications-contents.html
// and http://docs.aws.amazon.com/ses/latest/DeveloperGuide/receiving-email-action-lambda-example-functions.html
type records struct {
	SES struct {
		Mail struct {
			MessageID   string   `json:"messageId"`
			Destination []string `json:"destination"`
		} `json:"mail"`
	} `json:"ses"`
}

type query struct {
	Records []records `json:"Records"`
}

type SpfLookup struct {

	// Returns true if a valid SPF record is found
	Valid bool `json:"valid"`

	// Raw SPF record
	Record string `json:"record,omitempty"`

	// Verbose English explanation of SPF policy
	Description string `json:"description,omitempty"`

	Records []string `json:"records"`
}

func publishNotification(snsTopicArn string, snsSubject string, snsMessage string, wg *sync.WaitGroup) {
	defer wg.Done()

	sess, err := session.NewSession()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create session: %s", err.Error())
		return
	}

	svc := sns.New(sess)

	params := &sns.PublishInput{
		Message: aws.String(snsMessage), // Required

		Subject:  aws.String(snsSubject),
		TopicArn: aws.String(snsTopicArn),
	}
	resp, err := svc.Publish(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Fprintln(os.Stderr, resp)
}

func main() {
	lambda.Start(func(ctx context.Context, event json.RawMessage) (interface{}, error) {
		var spf SpfLookup
		var q query

		if err := json.Unmarshal(event, &q); err != nil {
			return nil, err
		}

		svc := s3.New(session.New())

		bucket_name := os.Getenv("BUCKET_NAME")
		bucket_prefix := os.Getenv("BUCKET_PREFIX")
		object_key := bucket_prefix + q.Records[0].SES.Mail.MessageID

		params := &s3.GetObjectInput{
			Bucket: aws.String(bucket_name), // Required
			Key:    aws.String(object_key),  // Required
		}
		resp, err := svc.GetObject(params)

		if err != nil {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Fprintln(os.Stderr, err.Error())
			return nil, err
		}

		log.Println("message id:  ", object_key)

		// Pretty-print the response data.
		m, err := DmarcReportPrepareAttachment(resp.Body)
		if err != nil {
			panic(err)
		}

		parse(m, object_key)

		spf.Description = "dummy struct is used"

		return spf, nil
	})
}
