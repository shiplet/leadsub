package types

import (
	"encoding/csv"
	"net/http"
)

type GridLeadRequest struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}

type GridLeadResponse struct {
	Total int            `json:"total"`
	Data  []GridLeadData `json:"data"`
}

type GridLeadData struct {
	ID     string `json:"id"`
	LeadID string `json:"leadID"`
}

type TrueLeadTrackingData struct {
	Success bool         `json:"success"`
	Data    TrackingData `json:"data"`
	Msg     string       `json:"msg"`
}

type TrackingData struct {
	Response       string `json:"response"`
	CampaignName   string `json:"campaignName"`
	SuccessMsg     string
	RequestURL     string `json:"requestURL"`
	LandingPageURL string `json:"landingPageURL"`
	ReferralURL    string `json:"referralURL"`
}

type TrackingResponse struct {
	Errors []TrackingError `json:"errors"`
	LeadID string          `json:"lead_id"`
	Msg    string          `json:"msg"`
	Price  string          `json:"price"`
	Result string          `json:"result"`
}

type TrackingError struct {
	Error string `json:"error"`
	Field string `json:"field"`
}

type LeadsCSV struct {
	Title   string
	Writer  *csv.Writer
	Records [][]string
}

type CsvRecordBuilder struct {
	LeadID               string
	LeadspediaInternalID string
	Campaign             string
	LeadspediaResponse   string
	RequestURL           string
	LandingPageURL       string
	ReferralURL          string
}

type TrueLeadIDChannel struct {
	Res   *http.Response
	Id    string
	Index int
	Csv   *LeadsCSV
}

type TrackingLeadDataChannel struct {
	Id     string
	LeadID string
	Csv    *LeadsCSV
}

type GetCallsRequest struct {
	From  string
	Start int
	Limit int
}

type AllCallsResponse struct {
	Success  bool                  `json:"success"`
	Response AllCallsInnerResponse `json:"response"`
}

type AllCallsInnerResponse struct {
	Start int        `json:"start"`
	Limit int        `json:"limit"`
	Total int        `json:"total"`
	Data  []CallData `json:"data"`
}

type CallData struct {
	InboundCallID string `json:"inboundCallID"`
	CallUUID      string `json:"callUUID"`
}
