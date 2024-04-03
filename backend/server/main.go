package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fraudmarc/fraudmarc-ce/backend/lib"
	"github.com/gorilla/mux"
)

type query struct {
	Params struct {
		Path struct {
			Domain    string `json:"domain"`
			StartDate string `json:"start"`
			EndDate   string `json:"end"`
		} `json:"path"`
		QueryString struct {
			StartDate  string `json:"start"`
			EndDate    string `json:"end"`
			Source     string `json:"source"`
			SourceType string `json:"source_type"`
		} `json:"querystring"`
	} `json:"params"`
}

type getDetailReturn struct {
	DetailRows []lib.DmarcReportingForwarded `json:"detail_rows"`
}
type result []interface{}

func main() {
	var reports []lib.AggregateReport
	err := lib.DBreporting.SQL(`
		select * from "AggregateReport"`).QueryStructs(&reports)
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.HandleFunc("/api/domains/", handleDomainList)
	r.HandleFunc("/api/domains/{domain}/report", handleDomainSummary)
	r.HandleFunc("/api/domains/{domain}/report/detail", handleDmarcDetail)
	r.HandleFunc("/api/domains/{domain}/chart/dmarc", handleDmarcChart)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("/dist"))))
	if err = http.ListenAndServe(":7489", r); err != nil {
		log.Fatal(err)
	}
}

func handleDomainList(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")

	domainList := lib.GetDomainList()
	enc := json.NewEncoder(w)
	enc.Encode(&domainList)
}

func handleDomainSummary(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	vals := r.URL.Query()
	vars := mux.Vars(r)
	domain := vars["domain"]
	defer r.Body.Close()
	gr := lib.GetDmarcReportGeneral(vals["start"][0], vals["end"][0], domain)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(&gr)
}

func handleDmarcChart(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	defer r.Body.Close()
	vals := r.URL.Query()
	vars := mux.Vars(r)
	domain := vars["domain"]

	log.Println(vals["end"][0])

	chartRet, err := lib.GetDmarcChartData(vals["start"][0], vals["end"][0], domain)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(&chartRet)

}

func handleDmarcDetail(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	vals := r.URL.Query()
	vars := mux.Vars(r)
	domain := vars["domain"]

	gr := getDetailReturn{}

	// query all entries for specific sender:
	results := lib.GetDmarcReportDetail(vals["start"][0], vals["end"][0], domain, vals["source"][0], vals["source_type"][0])
	gr.DetailRows = results

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(&gr)

}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
