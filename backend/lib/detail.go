package lib

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"
)

// GetDmarcReportDetail returns the dmarc report details used to be shown on detail panel
func GetDmarcReportDetail(startdate, endate, domain, source, sourcetype string) []DmarcReportingForwarded {

	t := time.Now()

	tStart, err := time.Parse(time.RFC3339Nano, startdate)
	if err != nil {
		log.Println("ERROR reading start time", err)
		tStart = time.Now().AddDate(0, 0, -30)
	}
	tEnd, err := time.Parse(time.RFC3339Nano, endate)
	if err != nil {
		log.Println("ERROR reading end time", err)
		tEnd = time.Now()
	}

	if tEnd.After(t) {
		tEnd = t
	}

	start := tStart.Unix()
	end := tEnd.Unix()

	results := []DmarcReportingForwarded{}

	selectTerm := `
	SELECT 	
	  SUM(message_count) AS count,
	  source_ip,
	  esp,
	  domain_name,
	  host_name,
	  revlookup.i,
	  country,
	  disposition,
	  eval_dkim,
	  eval_spf,
	  header_from,
	  envelope_from,
	  envelope_to,
	  auth_dkim_domain,
	  auth_dkim_selector,
	  auth_dkim_result,
	  auth_spf_domain,
	  auth_spf_scope,
	  auth_spf_result,
	  po_reason,
	  po_comment`

	groupTerm := `
	GROUP BY
	  source_ip,
	  esp,
	  domain_name,
	  host_name,
	  revlookup.i,
	  country,
	  disposition,
	  eval_dkim,
	  eval_spf,
	  header_from,
	  envelope_from,
	  envelope_to,
	  auth_dkim_domain,
	  auth_dkim_selector,
	  auth_dkim_result,
	  auth_spf_domain,
	  auth_spf_scope,
	  auth_spf_result,
	  po_reason,
	  po_comment
	`
	from := `FROM dmarc_reporting_entries dre cross join lateral unnest(coalesce(nullif(reverse_lookup,'{}'),array[null::text])) as revlookup(i)`

	source2 := fmt.Sprintf("%s%s%s", "%", source, ".")

	qargs := []interface{}{domain, start, end, source}

	var qterm string

	//Dependence here on empty strings instead of nulls, and on the Label() function algorithm.
	if net.ParseIP(source) != nil {
		qterm = `AND source_ip = $4::inet`
	} else {
		qterm = `AND (esp = $4 OR
	( esp = '' AND ( domain_name = $4 OR ( domain_name = '' AND (revlookup.i LIKE $5)))))`
		qargs = append(qargs, source2)
	}

	qsql := fmt.Sprintf(`%s
%s
WHERE dre.domain = $1
AND dre.end_date >= $2
AND dre.end_date <= $3
%s
`, selectTerm, from, qterm)

	qsql = fmt.Sprintf("%s%s ORDER BY count DESC", qsql, groupTerm)

	rows, errQuery := DB.Query(qsql, qargs...)
	defer rows.Close()

	if errQuery != nil {
		log.Println("errQuery:  ", errQuery)
		return results
	}

	k := 0
	for rows.Next() {

		if err := rows.Err(); err != nil {
			log.Println("errRows:  ", err)
			return results
		}

		k++
		dr := DmarcReportingForwarded{}

		// these hold bytes to be converted to string arrays in scanned data:
		var revlookup sql.RawBytes
		var adkimdomain, adkimselector, adkimresult sql.RawBytes
		var aspfdomain, aspfscope, aspfresult sql.RawBytes
		var aporeason, apocomment sql.RawBytes

		errScan := rows.Scan(
			&dr.MessageCount,
			&dr.SourceIP,
			&dr.ESP,
			&dr.DomainName,
			&dr.HostName,
			&revlookup,
			&dr.Country,
			&dr.Disposition,
			&dr.EvalDKIM,
			&dr.EvalSPF,
			&dr.HeaderFrom,
			&dr.EnvelopeFrom,
			&dr.EnvelopeTo,
			&adkimdomain,
			&adkimselector,
			&adkimresult,
			&aspfdomain,
			&aspfscope,
			&aspfresult,
			&aporeason,
			&apocomment,
		)

		dr.ReverseLookup = rawBytesToStringArray(revlookup)
		dr.AuthDKIMDomain = rawBytesToStringArray(adkimdomain)
		dr.AuthDKIMSelector = rawBytesToStringArray(adkimselector)
		dr.AuthDKIMResult = rawBytesToStringArray(adkimresult)
		dr.AuthSPFDomain = rawBytesToStringArray(aspfdomain)
		dr.AuthSPFScope = rawBytesToStringArray(aspfscope)
		dr.AuthSPFResult = rawBytesToStringArray(aspfresult)
		dr.POReason = rawBytesToStringArray(aporeason)
		dr.POComment = rawBytesToStringArray(apocomment)

		if errScan != nil {
		}

		results = append(results, dr)

	} // rows.Next

	return results

}
