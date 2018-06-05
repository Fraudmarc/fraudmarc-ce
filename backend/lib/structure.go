package lib

import (
	"strings"
)

//DmarcReportingSummary structure used on summary table
type DmarcReportingSummary struct {
	Source               string `json:"source"` // could be name, domain_name, or IP
	TotalCount           int64  `json:"total_count"`
	DispositionPassCount int64  `json:"pass_count"`
	SPFAlignedCount      int64  `json:"spf_aligned_count"`
	DKIMAlignedCount     int64  `json:"dkim_aligned_count"`
	FullyAlignedCount    int64  `json:"fully_aligned_count"`
	SourceType           string `json:"source_type"`
}

//DomainSummaryCounts structure used on calculating the whole volume and passing rate in the time range of one domain
type DomainSummaryCounts struct {
	ReportCount       int64 `json:"report_count"`
	MessageCount      int64 `json:"message_count"`
	DKIMAlignedCount  int64 `json:"dkim_aligned_count"`
	SPFAlignedCount   int64 `json:"spf_aligned_count"`
	FullyAlignedCount int64 `json:"fully_aligned_count"`
}

//DmarcReportingDefault structure used to
type DmarcReportingDefault struct {
	SourceIP      string   `json:"source_ip" db:"source_ip"`
	ESP           string   `json:"esp" db:"esp"`
	HostName      string   `json:"host_name" db:"host_name"`
	DomainName    string   `json:"domain_name" db:"domain_name"`
	Country       string   `json:"country" db:"country"`
	MessageCount  int64    `json:"message_count" db:"message_count"`
	Disposition   string   `json:"disposition" db:"disposition"`
	EvalDKIM      string   `json:"eval_dkim" db:"eval_dkim"`
	EvalSPF       string   `json:"eval_spf" db:"eval_spf"`
	ReverseLookup []string `json:"reverse_lookup" db:"reverse_lookup"`
}

type DMARCStats struct {
	MessageCount         int64
	DispositionPassCount int64
	SPFAlignedCount      int64
	DKIMAlignedCount     int64
	FullyAlignedCount    int64
	SourceType           string
}

//DmarcReportingForwarded structure feeds the data used to generate detail table
type DmarcReportingForwarded struct {
	SourceIP         string   `json:"source_ip" db:"source_ip"`
	ESP              string   `json:"esp" db:"esp"`
	DomainName       string   `json:"domain_name" db:"domain_name"`
	HostName         string   `json:"host_name" db:"host_name"`
	ReverseLookup    []string `json:"reverse_lookup" db:"reverse_lookup"`
	Country          string   `json:"country" db:"country"`
	MessageCount     int64    `json:"message_count" db:"message_count"`
	Disposition      string   `json:"disposition" db:"disposition"`
	EvalDKIM         string   `json:"eval_dkim" db:"eval_dkim"`
	EvalSPF          string   `json:"eval_spf" db:"eval_spf"`
	HeaderFrom       string   `json:"header_from" db:"header_from"`
	EnvelopeFrom     string   `json:"envelope_from" db:"envelope_from"`
	EnvelopeTo       string   `json:"envelope_to" db:"envelope_to"`
	AuthDKIMDomain   []string `json:"auth_dkim_domain" db:"auth_dkim_domain"`
	AuthDKIMSelector []string `json:"auth_dkim_selector" db:"auth_dkim_selector"`
	AuthDKIMResult   []string `json:"auth_dkim_result" db:"auth_dkim_result"`
	AuthSPFDomain    []string `json:"auth_spf_domain" db:"auth_spf_domain"`
	AuthSPFScope     []string `json:"auth_spf_scope" db:"auth_spf_scope"`
	AuthSPFResult    []string `json:"auth_spf_result" db:"auth_spf_result"`
	POReason         []string `json:"po_reason" db:"po_reason"`
	POComment        []string `json:"po_comment" db:"po_comment"`
	Count            int64    `json:"count,omitempty" db:"count"` // used with queries involving SUM(message_count) AS count
}

type Values struct {
	Value *int64 `db:"Value" json:"value"`
}

type DmarcDailyBuckets struct {
	Day     int64
	Passing int64
	Failing int64
}

type DmarcReportingSummaryList []DmarcReportingSummary

func (d DmarcReportingSummaryList) Len() int      { return len(d) }
func (d DmarcReportingSummaryList) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d DmarcReportingSummaryList) Less(i, j int) bool {
	if d[i].TotalCount < d[j].TotalCount {
		return true
	}
	if d[i].TotalCount > d[j].TotalCount {
		return false
	}
	return d[i].TotalCount < d[j].TotalCount
}

type getSummaryReturn struct {
	Summary             []DmarcReportingSummary `json:"summary"`
	MaxVolume           int64                   `json:"max_volume"`
	DomainSummaryCounts DomainSummaryCounts     `json:"domain_summary_counts"`
	StartDate           string                  `json:"start_date"`
	EndDate             string                  `json:"end_date"`
	Domain              string                  `json:"domain"`
}

func (d DmarcReportingDefault) Label() (source string, sourceType string) {

	// last resort is to show IP:
	source = d.SourceIP
	sourceType = "IP"

	// these are ordered from most-preferred to next-to-least preferred:
	if len(d.ESP) > 0 {
		source = d.ESP
		sourceType = "ESP"
	} else if len(d.DomainName) > 0 {
		source = d.DomainName
		sourceType = "DomainName"
	} else if len(d.HostName) > 0 {
		source = d.HostName
		sourceType = "HostName"
	} else if len(d.ReverseLookup[0]) > 0 {
		revsource, _ := GetOrgDomain(d.ReverseLookup[0])
		if len(revsource) > 0 {
			revsource = strings.TrimRight(revsource, ".")
			if len(revsource) > 0 {
				source = revsource
				sourceType = "ReverseLookup"
			}
		}
	}

	return

} // Label
