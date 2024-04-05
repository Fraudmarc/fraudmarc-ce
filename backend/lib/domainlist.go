package lib

import (
	"github.com/lib/pq"
)

// GetDomainList returns the array of domains users have set up the dmarc policy for
func GetDomainList() pq.StringArray {
	qsql := `SELECT DISTINCT "domain" from dmarc_reporting_entries`
	var result pq.StringArray
	DBreporting.SQL(qsql).QueryStructs(&result)
	return result
}
