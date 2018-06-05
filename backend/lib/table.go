package lib

import (
	"fmt"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type AggregateReport struct {
	MessageId         string                  `db:"MessageId"`
	Organization      string                  `xml:"report_metadata>org_name" db:"Organization"`
	Email             string                  `xml:"report_metadata>email" db:"Email"`
	ExtraContact      string                  `xml:"report_metadata>extra_contact_info" db:"ExtraContact"` // minOccurs="0"
	ReportID          string                  `xml:"report_metadata>report_id" db:"ReportID"`
	RawDateRangeBegin string                  `xml:"report_metadata>date_range>begin" db:"RawDateRangeBegin"`
	RawDateRangeEnd   string                  `xml:"report_metadata>date_range>end" db:"RawDateRangeEnd"`
	DateRangeBegin    int64                   `db:"DateRangeBegin"`
	DateRangeEnd      int64                   `db:"DateRangeEnd"`
	Errors            []string                `xml:"report_metadata>error" db:"Errors"`
	Domain            string                  `xml:"policy_published>domain" db:"Domain"`
	AlignDKIM         string                  `xml:"policy_published>adkim" db:"AlignDKIM"` // minOccurs="0"
	AlignSPF          string                  `xml:"policy_published>aspf" db:"AlignSPF"`   // minOccurs="0"
	Policy            string                  `xml:"policy_published>p" db:"Policy"`
	SubdomainPolicy   string                  `xml:"policy_published>sp" db:"SubdomainPolicy"`
	Percentage        int                     `xml:"policy_published>pct" db:"Percentage"`
	FailureReport     string                  `xml:"policy_published>fo" db:"FailureReport"`
	Records           []AggregateReportRecord `xml:"record" db:"Records"`
}

type AggregateReportRecord struct {
	SourceIP           string           `xml:"row>source_ip" db:"SourceIP"`
	Count              int64            `xml:"row>count" db:"Count"`
	Disposition        string           `xml:"row>policy_evaluated>disposition" db:"Disposition"` // ignore, quarantine, reject
	EvalDKIM           string           `xml:"row>policy_evaluated>dkim" db:"EvalDKIM"`           // pass, fail
	EvalSPF            string           `xml:"row>policy_evaluated>spf" db:"EvalSPF"`             // pass, fail
	POReason           []POReason       `xml:"row>policy_evaluated>reason" db:"POReason"`
	HeaderFrom         string           `xml:"identifiers>header_from" db:"HeaderFrom"`
	EnvelopeFrom       string           `xml:"identifiers>envelope_from" db:"EnvelopeFrom"`
	EnvelopeTo         string           `xml:"identifiers>envelope_to" db:"EnvelopeTo"` // min 0
	AuthDKIM           []DKIMAuthResult `xml:"auth_results>dkim" db:"AuthDKIM"`         // min 0
	AuthSPF            []SPFAuthResult  `xml:"auth_results>spf" db:"AuthSPF"`
	AggregateReport_id string           `db:"AggregateReport_id"`
	RecordNumber       int64            `db:"RecordNumber"`
}

type POReason struct {
	Reason             string `xml:"type" db:"Reason"`
	Comment            string `xml:"comment" db:"Comment"`
	AggregateReport_id string `db:"AggregateReport_id"`
	RecordNumber       int64  `db:"RecordNumber"`
}

type DKIMAuthResult struct {
	Domain             string `xml:"domain" db:"Domain"`
	Selector           string `xml:"selector" db:"Selector"`
	Result             string `xml:"result" db:"Result"`
	HumanResult        string `xml:"human_result" db:"HumanResult"`
	AggregateReport_id string `db:"AggregateReport_id"`
	RecordNumber       int64  `db:"RecordNumber"`
}

type SPFAuthResult struct {
	Domain             string `xml:"domain" db:"Domain"`
	Scope              string `xml:"scope" db:"Scope"`
	Result             string `xml:"result" db:"Result"`
	AggregateReport_id string `db:"AggregateReport_id"`
	RecordNumber       int64  `db:"RecordNumber"`
}

// DmarcReportingFull ...
type DmarcReportingFull struct {
	MessageId         string   `json:"message_id" db:"message_id"`
	RecordNumber      int64    `json:"record_number" db:"record_number"`
	Domain            string   `json:"domain" db:"domain"`
	Policy            string   `json:"policy" db:"policy"`
	SubdomainPolicy   string   `json:"subdomain_policy" db:"subdomain_policy"`
	AlignDKIM         string   `json:"align_dkim" db:"align_dkim"`
	AlignSPF          string   `json:"align_spf" db:"align_spf"`
	Pct               int      `json:"pct" db:"pct"`
	SourceIP          string   `json:"source_ip" db:"source_ip"`
	ESP               string   `json:"esp" db:"esp"`
	OrgName           string   `json:"org_name" db:"org_name"`
	OrgId             string   `json:"org_id" db:"org_id"`
	HostName          string   `json:"host_name" db:"host_name"`
	DomainName        string   `json:"domain_name" db:"domain_name"`
	HostnameMatchesIP string   `json:"host_name_matches_ip" db:"host_name_matches_ip"`
	City              string   `json:"city" db:"city"`
	State             string   `json:"state" db:"state"`
	Country           string   `json:"country" db:"country"`
	Longitude         string   `json:"longitude" db:"longitude"`
	Latitude          string   `json:"latitude" db:"latitude"`
	ReverseLookup     []string `json:"reverse_lookup" db:"reverse_lookup"`
	MessageCount      int64    `json:"message_count" db:"message_count"`
	Disposition       string   `json:"disposition" db:"disposition"`
	EvalDKIM          string   `json:"eval_dkim" db:"eval_dkim"`
	EvalSPF           string   `json:"eval_spf" db:"eval_spf"`
	HeaderFrom        string   `json:"header_from" db:"header_from"`
	EnvelopeFrom      string   `json:"envelope_from" db:"envelope_from"`
	EnvelopeTo        string   `json:"envelope_to" db:"envelope_to"`
	AuthDKIMDomain    []string `json:"auth_dkim_domain" db:"auth_dkim_domain"`
	AuthDKIMSelector  []string `json:"auth_dkim_selector" db:"auth_dkim_selector"`
	AuthDKIMResult    []string `json:"auth_dkim_result" db:"auth_dkim_result"`
	AuthSPFDomain     []string `json:"auth_spf_domain" db:"auth_spf_domain"`
	AuthSPFScope      []string `json:"auth_spf_scope" db:"auth_spf_scope"`
	AuthSPFResult     []string `json:"auth_spf_result" db:"auth_spf_result"`
	POReason          []string `json:"po_reason" db:"po_reason"`
	POComment         []string `json:"po_comment" db:"po_comment"`
	StartDate         int64    `json:"start_date" db:"start_date"`
	EndDate           int64    `json:"end_date" db:"end_date"`
	LastUpdate        string   `json:"last_update" db:"last_update"`
	Id                int64    `json:"id" db:"id"`
} // DmarcReportingFull

type SBGeo struct {
	OrgName           string `json:"org_name" db:"org_name"`
	OrgID             string `json:"org_id" db:"org_id"`
	OrgCategory       string `json:"org_category" db:"org_category"`
	Hostname          string `json:"hostname" db:"hostname"`
	DomainName        string `json:"domain_name" db:"domain_name"`
	HostnameMatchesIP string `json:"hostname_matches_ip" db:"hostname_matches_ip"`
	City              string `json:"city" db:"city"`
	State             string `json:"state" db:"state"`
	Country           string `json:"country" db:"country"`
	Longitude         string `json:"longitude" db:"longitude"`
	Latitude          string `json:"latitude" db:"latitude"`
}

// GetOrgDomain returns the org domain for the input domain according to the
// mechanisms of code in the publicsuffix package
func GetOrgDomain(domain string) (orgDomain string, err error) {

	// trailing periods should be stripped before splitting:
	domain = strings.TrimRight(domain, ".")

	// resolve organizational domain:
	icannDomain, isIcann := publicsuffix.PublicSuffix(domain)

	if !isIcann {
		// bad domain spec
		//  -- not sure how we will ever be here if HostDomain is compliant
		err = fmt.Errorf("bad organizational domain: %s", domain)
		return
	}

	domainLabels := strings.Split(domain, ".")
	icannLabels := strings.Split(icannDomain, ".")
	icl := len(icannLabels)
	dcl := len(domainLabels)
	orgDomain = ""

	if dcl-icl <= 1 {
		// domain is org domain:
		orgDomain = domain
		return
	}
	for j := dcl - icl - 1; j < dcl; j++ {
		orgDomain = orgDomain + domainLabels[j] + "."
	}

	return
}
