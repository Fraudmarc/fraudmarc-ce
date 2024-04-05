package lib

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

func GetDmarcReportGeneral(startdate, endate, domain string) getSummaryReturn {

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

	gr := getSummaryReturn{
		Domain: domain,
	}

	summary := []DmarcReportingSummary{}
	counts := DomainSummaryCounts{}

	// harvest all reports that have been received in the last 'dayCount' days:

	summary, counts = QSummary(domain, start, end)

	gr.StartDate = tStart.Format(time.RFC3339Nano)
	gr.EndDate = tEnd.Format(time.RFC3339Nano)

	gr.Summary = summary
	gr.DomainSummaryCounts = counts
	if len(summary) > 0 {
		gr.MaxVolume = summary[0].TotalCount
	}

	return gr
}

// QSummary is a summary view of dmarc evaluations per sender
// Possible cases to consider:
// 1.  Sender company name is resolved        -> use name
// 2.  Sender domain but not name is resolved -> use domain
// 3.  Sender domain and name both unresolved -> use IP
func QSummary(domain string, start int64, end int64) ([]DmarcReportingSummary, DomainSummaryCounts) {

	summaryRowMax := 1000
	results := []DmarcReportingSummary{}
	thecounts := DomainSummaryCounts{}

	//----------------------------------------------------------------
	// query the full set of company records:
	//----------------------------------------------------------------
	qsql := fmt.Sprintf(`SELECT SUM(message_count) AS count,
	source_ip, esp, domain_name, reverse_lookup,
	country,
	disposition,
	eval_dkim,
	eval_spf
FROM dmarc_reporting_entries dre
WHERE dre.domain = $1
AND dre.end_date >= $2
AND dre.end_date <= $3
GROUP BY source_ip, esp, domain_name, reverse_lookup,
  country,
  disposition,
  eval_dkim,
  eval_spf
`)
	// one could ORDER BY here, but the way counts are counted will break that

	qargs := []interface{}{domain, start, end}

	rows, errQuery := DB.Query(qsql, qargs...)
	defer rows.Close()

	if errQuery != nil {
		log.Println("errQuery:  ", errQuery)
		return results, thecounts
	}

	//----------------------------------------------------------------
	// for each row, keep counts of the respective alignments:
	//----------------------------------------------------------------
	countMap := make(map[string]*DMARCStats)
	for rows.Next() {

		if err := rows.Err(); err != nil {
			log.Println("errRows:  ", err)
			return results, thecounts
		}

		dr := DmarcReportingDefault{}
		var revlookup sql.RawBytes

		errScan := rows.Scan(
			&dr.MessageCount,
			&dr.SourceIP,
			&dr.ESP,
			&dr.DomainName,
			&revlookup,
			&dr.Country,
			&dr.Disposition,
			&dr.EvalDKIM,
			&dr.EvalSPF,
		)
		if errScan != nil {
			log.Println("errScan:  ", errScan)
		}
		dr.ReverseLookup = rawBytesToStringArray(revlookup)

		source, sourceType := dr.Label()

		// initialize with new source if needed:
		if countMap[source] == nil {
			countMap[source] = &DMARCStats{
				SourceType: sourceType,
			}
		}

		// keep track of counts by source:
		countMap[source].MessageCount += dr.MessageCount
		if strings.Compare(dr.EvalDKIM, "pass") == 0 && strings.Compare(dr.EvalSPF, "pass") != 0 {
			countMap[source].DKIMAlignedCount += dr.MessageCount
			countMap[source].DispositionPassCount += dr.MessageCount
		} else if strings.Compare(dr.EvalDKIM, "pass") != 0 && strings.Compare(dr.EvalSPF, "pass") == 0 {
			countMap[source].SPFAlignedCount += dr.MessageCount
			countMap[source].DispositionPassCount += dr.MessageCount
		} else if strings.Compare(dr.EvalDKIM, "pass") == 0 && strings.Compare(dr.EvalSPF, "pass") == 0 {
			countMap[source].DKIMAlignedCount += dr.MessageCount
			countMap[source].SPFAlignedCount += dr.MessageCount
			countMap[source].FullyAlignedCount += dr.MessageCount
			countMap[source].DispositionPassCount += dr.MessageCount
		}

	} // rows.Next

	// build up the (now very unsorted) reporting summary data)
	for key, _ := range countMap {
		summary := DmarcReportingSummary{
			Source:               key,
			TotalCount:           countMap[key].MessageCount,
			DispositionPassCount: countMap[key].DispositionPassCount,
			SPFAlignedCount:      countMap[key].SPFAlignedCount,
			DKIMAlignedCount:     countMap[key].DKIMAlignedCount,
			FullyAlignedCount:    countMap[key].FullyAlignedCount,
			SourceType:           countMap[key].SourceType,
		}

		results = append(results, summary)

		thecounts.MessageCount += countMap[key].MessageCount
		thecounts.SPFAlignedCount += countMap[key].SPFAlignedCount
		thecounts.DKIMAlignedCount += countMap[key].DKIMAlignedCount
		thecounts.FullyAlignedCount += countMap[key].FullyAlignedCount
	}

	// sort in descending order by volume:
	sort.Sort(sort.Reverse(DmarcReportingSummaryList(results))) // don't fuck with this

	// build json:

	if len(results) > summaryRowMax {
		return results[0 : summaryRowMax-1], thecounts
	}
	return results, thecounts

} // QSummary

func rawBytesToStringArray(rb sql.RawBytes) []string {
	rbStr := string(rb)
	rbStr = strings.TrimLeft(rbStr, "{")
	rbStr = strings.TrimRight(rbStr, "}")
	return strings.Split(rbStr, ",")
} // rawBytesToStringArray
