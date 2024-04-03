package lib

import (
	"fmt"
	"log"
	"os"
	"time"
)

type daily struct {
	Name   string   `json:"name"`
	Series []Volume `json:"series"`
}

// Volume Name is the timestamp, Value is the pass/fail quantity on that day
type Volume struct {
	Name  int64 `json:"name"`
	Value int64 `json:"value"`
}

type chartReturn struct {
	ChartData []daily `json:"chartdata"`
	Domain    string  `json:"domain"`
}

// ChartContainer structure of result
type ChartContainer struct {
	Full []result `json:"full"`
	Pass []result `json:"pass"`
	Fail []result `json:"fail"`
}

type result []interface{}

func GetDmarcChartData(start, end, domain string) (chartReturn, error) {

	t := time.Now()

	tStart, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		log.Println("ERROR reading start time", err)
		tStart = time.Now().AddDate(0, 0, -30)
	}
	tEnd, err := time.Parse(time.RFC3339Nano, end)
	if err != nil {
		log.Println("ERROR reading end time", err)
		tEnd = time.Now()
	}

	if tEnd.After(t) {
		tEnd = t
	}

	tstart := tStart.Unix()
	tend := tEnd.Unix()

	rawData, err := GetDmarcDatedWeeklyChart(domain, tstart, tend)

	if err != nil {
		log.Printf("ERROR on GetDmarcDatedWeeklyChart, %s", err)
	}

	chartData := []daily{}
	chartData = append(chartData, daily{Name: "pass", Series: []Volume{}})
	chartData = append(chartData, daily{Name: "fail", Series: []Volume{}})

	for i, val := range rawData.Full {
		timestamp, _ := val[0].(int64)
		pass, _ := rawData.Pass[i][1].(int64)
		fail, _ := rawData.Fail[i][1].(int64)

		chartData[0].Series = append(chartData[0].Series, Volume{Name: timestamp, Value: pass})
		chartData[1].Series = append(chartData[1].Series, Volume{Name: timestamp, Value: fail})

	}

	retVal := chartReturn{
		ChartData: chartData,
		Domain:    domain,
	}

	return retVal, err
}

// GetDmarcDatedWeeklyChart returns the weekly dmarc data
func GetDmarcDatedWeeklyChart(domain string, start, end int64) (ChartContainer, error) {
	var chart ChartContainer

	if start == 0 || end == 0 {
		now := time.Now()
		end = now.UTC().Unix()
		start = now.UTC().AddDate(0, 0, -30).Unix()
	}

	dailyResults, err := getDmarcDailyAll(domain, start, end)
	if err != nil {
		return chart, err
	}

	fmt.Fprintf(os.Stderr, "number of days returned : %d\n %v", len(dailyResults), dailyResults)

	var currentTime int64
	var lastDay int64
	for _, day := range dailyResults {

		//Pad the graph with 0 days to keep entries for every day.
		for lastDay++; lastDay < day.Day; lastDay++ {
			currentTime = ((lastDay * 86000) + start) * 1000
			chart.Full = append(chart.Full, result{currentTime, 0})
			chart.Pass = append(chart.Pass, result{currentTime, 0})
			chart.Fail = append(chart.Fail, result{currentTime, 0})
		}

		currentTime = ((day.Day * 86000) + start) * 1000
		chart.Full = append(chart.Full, result{currentTime, day.Passing + day.Failing})
		chart.Pass = append(chart.Pass, result{currentTime, day.Passing})
		chart.Fail = append(chart.Fail, result{currentTime, day.Failing})
	}

	//Pad the end of the graph with 0 days to keep entries for every day.
	//NOTE: this padding and the loop above could probably be solved by using a generated series in the query
	if currentTime > 0 {
		if currentTime+(86000*1000) < end*1000 {
			log.Println("Padding end of the chart. Current", currentTime, " and end ", end)
		} else {
			log.Println("No chart padding required. Current", currentTime, " and end ", end)
		}
		for currentTime += 86000 * 1000; currentTime < end*1000; currentTime += 86000 * 1000 {
			chart.Full = append(chart.Full, result{currentTime, 0})
			chart.Pass = append(chart.Pass, result{currentTime, 0})
			chart.Fail = append(chart.Fail, result{currentTime, 0})
		}
	} else {
		log.Println("No data returned")
	}

	return chart, nil
}

func getDmarcDailyAll(domain string, timeBegin, timeEnd int64) ([]*DmarcDailyBuckets, error) {

	var results []*DmarcDailyBuckets

	//Normalize on day boundaries, using the start time as the boundary.
	days := (timeEnd - timeBegin) / 86400
	timeEnd = timeBegin + (days * 86400)

	fmt.Fprintf(os.Stderr, "days calculated: %d", days)

	tx, err := DBreporting.Begin()
	defer tx.AutoCommit()
	if err != nil {
		return results, err
	}

	_, err = tx.SQL(`SET LOCAL enable_hashjoin=off`).Exec()
	if err != nil {
		return results, err
	}

	err = tx.SQL(`
	select width_bucket(ar."DateRangeEnd", $1, $2, $3) as day,
	SUM(case when arr."EvalDKIM" = 'pass' or arr."EvalSPF" = 'pass' then arr."Count" else 0 end) as passing,
	SUM(case when arr."EvalDKIM" != 'pass' and arr."EvalSPF" != 'pass' then arr."Count" else 0 end) as failing
	FROM "AggregateReportRecord" arr JOIN "AggregateReport" ar ON arr."AggregateReport_id"=ar."MessageId"
	WHERE ar."DateRangeEnd" >= $1 AND ar."DateRangeEnd" < $2 AND ar."Domain" = $4
	group by day
	order by day`, timeBegin, timeEnd, days, domain).QueryStructs(&results)

	if err != nil {
		return results, err
	}

	return results, err

}
