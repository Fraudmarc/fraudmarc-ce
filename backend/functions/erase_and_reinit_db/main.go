package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fraudmarc/fraudmarc-ce/backend/lib"
)

func main() {
	lambda.Start(func(ctx context.Context, event json.RawMessage) (interface{}, error) {

		queryList := []string{
			`DROP TABLE IF EXISTS public."AggregateReport"`,
			`CREATE TABLE public."AggregateReport" (
				"MessageId" text,
				"Organization" text,
				"Email" text,
				"ExtraContact" text,
				"ReportID" text,
				"DateRangeBegin" bigint,
				"DateRangeEnd" bigint,
				"Domain" text,
				"AlignDKIM" text,
				"AlignSPF" text,
				"Policy" text,
				"SubdomainPolicy" text,
				"Percentage" integer,
				"FailureReport" text
			)`,
			`DROP TABLE IF EXISTS public."AggregateReportRecord"`,
			`CREATE TABLE public."AggregateReportRecord" (
				"AggregateReport_id" text,
				"RecordNumber" bigint,
				"SourceIP" text,
				"Count" bigint,
				"Disposition" text,
				"EvalDKIM" text,
				"EvalSPF" text,
				"HeaderFrom" text,
				"EnvelopeFrom" text,
				"EnvelopeTo" text
			)`,
			`DROP TABLE IF EXISTS public."DkimAuthResult"`,
			`CREATE TABLE public."DkimAuthResult" (
				"AggregateReport_id" text,
				"RecordNumber" bigint,
				"Domain" text,
				"Selector" text,
				"Result" text,
				"HumanResult" text
			)`,
			`DROP TABLE IF EXISTS public."PoReason"`,
			`CREATE TABLE public."PoReason" (
				"AggregateReport_id" text,
				"RecordNumber" bigint,
				"Reason" text,
				"Comment" text
			)`,
			`DROP TABLE IF EXISTS public."SpfAuthResult"`,
			`CREATE TABLE public."SpfAuthResult" (
				"AggregateReport_id" text,
				"RecordNumber" bigint,
				"Domain" text,
				"Scope" text,
				"Result" text
			)`,
			`DROP TABLE IF EXISTS public.dmarc_reporting_entries`,
			`CREATE TABLE public.dmarc_reporting_entries (
				message_id text,
				record_number integer,
				domain text,
				policy text,
				subdomain_policy text,
				align_dkim text,
				align_spf text,
				pct integer,
				source_ip inet,
				esp text,
				org_name text,
				org_id text,
				host_name text,
				host_name_matches_ip text,
				city text,
				state text,
				country text,
				longitude text,
				latitude text,
				reverse_lookup text[],
				message_count integer,
				disposition text,
				eval_dkim text,
				eval_spf text,
				header_from text,
				envelope_from text,
				envelope_to text,
				auth_dkim_domain text[],
				auth_dkim_selector text[],
				auth_dkim_result text[],
				auth_spf_domain text[],
				auth_spf_scope text[],
				auth_spf_result text[],
				po_reason text[],
				po_comment text[],
				start_date bigint,
				end_date bigint,
				last_update timestamp without time zone,
				domain_name text
			)`,
		}

		for i := range queryList {
			_, err := lib.GetReportingDB().SQL(queryList[i]).Exec()
			if err != nil {
				log.Printf("dbquery error: %s\n", err.Error())
				return nil, err
			}
		}

		return "done", nil
	})
}
