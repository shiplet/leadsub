package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"leadsub/calls"
	"leadsub/data"
	"leadsub/env"
	"leadsub/types"
	"leadsub/utils"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var emptyLeads []string
var skippedLeads []string

/*
leadsub combines public and non-public Leadspedia API's to get tracking data information
for leads, given each lead's public ID and/or UUID. It then exports the relevant data
into a CSV.
*/
func main() {
	/*
		Leadspedia tracks calls differently than traditional leads.
		Each call has a 36-character UUID, and the only way to determine
		which call we're dealing with from an API perspective, is to query every single call
		from a given time-period, and store its non-uuid identifier locally.
		Once we have this non-uuid, we can hit the private API using this value.
		This gives us the relevant lead information.
	*/
	start := time.Now()
	calls.GetAllCalls()
	fmt.Print("\n\n")

	total := buildCSV("full_missing_january.csv", data.Ids)

	fmt.Println("\nempty leads", len(emptyLeads))
	fmt.Println("skipped leads", len(skippedLeads))
	fmt.Println("total leads processed: ", total)
	fmt.Printf("total runtime: %.2f minutes\n", time.Since(start).Minutes())
}

/*
buildCSV queries the relevant data and builds the CSV containing the relevant data
*/
func buildCSV(title string, ids []string) int {
	var wg sync.WaitGroup
	var origID string
	file, err := os.Create(title)
	totalIds := len(ids)
	if err != nil {
		log.Fatalf("failed to create csv: %s", err)
	}

	lcsv := initCSV(file, title)
	callMsg := ""
	for index, id := range ids {
		bypass := false
		origID = id
		if len(id) > 15 {
			call := calls.FindCall(id)
			if call == nil {
				skippedLeads = append(skippedLeads, id)
				continue
			}

			id = call.InboundCallID
			origID = call.CallUUID
			bypass = true
			callMsg = fmt.Sprintf(" | adding data for call: %s", call.CallUUID)
		}
		utils.Printing.Workers = fmt.Sprintf("\rprepping workers for %s: %d/%d %s", title, index+1, len(ids), callMsg)
		utils.HandlePrinting()
		wg.Add(1)
		go getGridLeads(id, origID, totalIds, lcsv, bypass, &wg)
		time.Sleep(100 * time.Millisecond)
	}
	wg.Wait()
	lcsv.Writer.WriteAll(lcsv.Records)
	fmt.Println()
	return len(lcsv.Records)
}

/*
initCSV creates the in-memory CSV data, and populates the header row
*/
func initCSV(writer io.Writer, title string) *types.LeadsCSV {
	return &types.LeadsCSV{
		Title:  title,
		Writer: csv.NewWriter(writer),
		Records: [][]string{
			{"Lead ID", "Leadspedia Internal ID", "Campaign", "Leadspedia Response", "Landing Page URL", "Referral URL", "Request URL"},
		},
	}
}

/*
getGridLeads queries Leadspedia's non-public API using their notion of a LeadID
in order to get the relevant information for each lead
*/
func getGridLeads(id string, origID string, totalIds int, csv *types.LeadsCSV, bypassIdCheck bool, wg *sync.WaitGroup) {
	defer wg.Done()
	client := &http.Client{}
	body, err := json.Marshal(types.GridLeadRequest{
		Property: "leadID",
		Value:    id,
	})

	if err != nil {
		log.Fatalf("Failed to parse JSON: %s", err)
	}

	uri := fmt.Sprintf("https://vivint.leadspedia.net/app/data/leads/Grid-Leads.php?filter=[%s]", body)
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("Cookie", env.PHP_SESSID)
	res, e := client.Do(req)
	if e != nil {
		log.Fatalln("error: ", e)
	}

	var leadID string
	if bypassIdCheck {
		leadID = id
	} else {
		leadID = getTrueLeadID(res, id)
	}

	var record types.CsvRecordBuilder

	if leadID == "" {
		record.LeadID = origID
		record.LeadspediaInternalID = ""
		record.LandingPageURL = ""
		record.ReferralURL = ""
		record.RequestURL = ""
	} else {
		leadTrackingData := getLeadTrackingData(leadID)
		if leadTrackingData != nil {
			record.LeadID = origID
			record.LeadspediaInternalID = leadID
			record.Campaign = leadTrackingData.Data.CampaignName
			record.LeadspediaResponse = leadTrackingData.Data.Response
			record.LandingPageURL = leadTrackingData.Data.LandingPageURL
			record.ReferralURL = leadTrackingData.Data.ReferralURL
			record.RequestURL = leadTrackingData.Data.RequestURL
		} else {
			record.LeadID = origID
			record.LeadspediaInternalID = id
			record.LandingPageURL = ""
			record.ReferralURL = ""
			record.RequestURL = ""
		}
	}

	csv.Records = append(csv.Records, []string{
		record.LeadID,
		record.LeadspediaInternalID,
		record.Campaign,
		record.LeadspediaResponse,
		record.LandingPageURL,
		record.ReferralURL,
		record.RequestURL,
	})

	progress := len(csv.Records)
	current := int(math.Round(float64(progress-1) / float64(totalIds) * 100))
	utils.Printing.Progress = fmt.Sprintf("\r%s [%s%s] %d/%d", csv.Title, strings.Repeat("=", current), strings.Repeat(" ", 100-current), progress-1, totalIds)
	utils.HandlePrinting()
	return
}

/*
getTrueLeadID parses Leadspedia responses to get each lead's internal, or "true", ID
in order to query Leadspedia's non-public API for relevant information
*/
func getTrueLeadID(resp *http.Response, id string) string {
	var parsed types.GridLeadResponse

	err := json.NewDecoder(resp.Body).Decode(&parsed)
	if err != nil {
		emptyLeads = append(emptyLeads, id)
		return ""
	}

	return parsed.Data[0].ID
}

/*
getLeadTrackingData queries Leadspedia's non-public API for tracking- and campaign-related data
*/
func getLeadTrackingData(trueLeadID string) *types.TrueLeadTrackingData {
	client := &http.Client{}
	data := url.Values{}
	data.Set("id", trueLeadID)

	req, _ := http.NewRequest("POST", "https://vivint.leadspedia.net/app/data/leads/Form-Tracking.php", strings.NewReader(data.Encode()))
	req.Header.Add("Cookie", env.PHP_SESSID)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to request tracking data: %s", err)
	}

	var parsed types.TrueLeadTrackingData

	err = json.NewDecoder(res.Body).Decode(&parsed)
	if err != nil {
		log.Fatalf("failed to decode trueLeadTrackingData: %s", err)
	}

	if parsed.Success {
		return &parsed
	}

	fmt.Printf("\n\n\nrequest failed: %s\n\n\n", parsed.Msg)
	return nil
}

/*
getLeadTrackingResponse parses Leadspedia's response and returns the Golang object
*/
func getLeadTrackingResponse(data *types.TrueLeadTrackingData) *types.TrackingData {
	var parsed types.TrackingData

	err := json.NewDecoder(strings.NewReader(data.Data.Response)).Decode(&parsed)
	if err != nil {
		return &types.TrackingData{
			SuccessMsg: data.Data.Response,
		}
	}
	return &parsed
}
