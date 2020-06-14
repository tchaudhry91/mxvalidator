package mxvalidate

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"
)

var badRecords = []string{
	"localhost",
	"127.0.0.1",
	"0.0.0.0",
	"",
	".",
}

const statusUnresolvable = "Unresolvable"
const statusInvalidMX = "InvalidMX"
const statusValidMX = "ValidMX"

// ValidateRequest is the request containing domains to validate
type ValidateRequest struct {
	Domains []string `json:"domains,omitempty"`
}

// ValidateResponse is the response containing the results of the validation
type ValidateResponse struct {
	Results []DomainResponse `json:"results,omitempty"`
}

func respond(w http.ResponseWriter, req *http.Request, data interface{}, status int, err error) {
	defer log.Printf("Status:%d", status)
	if err != nil {
		defer log.Printf("Err: %v", err)
	}
	w.WriteHeader(status)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			log.Printf("Err: %v", err)
		}
	}
}

// DomainResponse holds the response data for a single domain
type DomainResponse struct {
	Domain string `json:"domain,omitempty"`
	Valid  bool   `json:"valid"`
	Status string `json:"status,omitempty"`
	AnyMX  string `json:"any_mx,omitempty"`
}

// ValidateMX is a cloud function to validate mx records for domains
func ValidateMX(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "mxvalidator.tux-sudo.com")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "mxvalidator.tux-sudo.com")

	// Decode request
	var req ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respond(w, r, nil, http.StatusBadRequest, err)
	}
	resp := ValidateResponse{
		Results: []DomainResponse{},
	}

	var wg sync.WaitGroup

	wg.Add(len(req.Domains))
	resChan := make(chan DomainResponse)

	for _, domain := range req.Domains {
		go func(dom string) {
			resChan <- validate(dom)
			wg.Done()
		}(domain)
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	for domResp := range resChan {
		resp.Results = append(resp.Results, domResp)
	}

	respond(w, r, resp, 200, nil)
}

func validate(domain string) DomainResponse {
	mxRes, err := net.LookupMX(domain)
	if err != nil {
		return DomainResponse{domain, false, statusUnresolvable, ""}
	}

	anyMX := ""
	for i := range mxRes {
		anyMX = mxRes[i].Host
		if !stringInSlice(anyMX, badRecords) {
			return DomainResponse{domain, true, statusValidMX, anyMX}
		}
	}
	return DomainResponse{domain, false, statusInvalidMX, anyMX}
}

func stringInSlice(s string, ss []string) bool {
	for _, str := range ss {
		if s == str {
			return true
		}
	}
	return false
}
