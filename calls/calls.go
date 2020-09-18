package calls

import (
	"encoding/json"
	"fmt"
	"leadsub/env"
	"leadsub/types"
	"leadsub/utils"
	"log"
	"net/http"
)

var allCalls []types.CallData
var startDates []string = []string{
	"2020-01-01",
	"2020-02-01",
	"2020-03-01",
	"2020-04-01",
	"2020-05-01",
	"2020-06-01",
	"2020-07-01",
	"2020-08-01",
	"2020-09-01",
}

/*
GetAllCalls gathers all leadspedia calls in a series of max-1000-item batches,
based on the total available in the specified timeframe
*/
func GetAllCalls() {
	for _, date := range startDates {
		getCallsForMonth(date)
	}
	fmt.Println("total calls found: ", len(allCalls))
}

/*
getCallsForMonth captures and batches calls to Leadspedia's public Call API.
Requests get sent in 1000-item chunks, and results get appended to allCalls.
*/
func getCallsForMonth(dateRange string) {
	res, e := getCalls(types.GetCallsRequest{
		From:  dateRange,
		Start: 0,
		Limit: 1,
	})
	if e != nil {
		log.Fatalln("failed to get calls: ", e)
	}

	if res.Response.Total > 0 {
		var reqs []types.GetCallsRequest

		limit := 1000
		batch := 0
		for i := 0; i < res.Response.Total; i += limit {
			batch = utils.Min(limit, res.Response.Total-(len(reqs)*limit))
			if batch != 0 {
				reqs = append(reqs, types.GetCallsRequest{
					From:  dateRange,
					Start: i,
					Limit: batch,
				})
			}
		}

		for i, req := range reqs {
			fmt.Printf("\rgathering call data for %s, batch %d of %d", dateRange, i+1, len(reqs))
			if res, e := getCalls(req); e != nil {
				log.Fatalln("failed to fetch call data: ", e)
			} else {
				allCalls = append(allCalls, res.Response.Data[:]...)
			}
		}

		fmt.Println()
	}
}

/*
FindCall finds the call with the specified UUID and returns nil if not found
*/
func FindCall(uuid string) *types.CallData {
	for _, call := range allCalls {
		if call.CallUUID == uuid {
			return &call
		}
	}

	return nil
}

/*
getCalls queries Leadspedia's public API for call data given parameters specified by the GetCallsRequest.
*/
func getCalls(req types.GetCallsRequest) (*types.AllCallsResponse, error) {
	client := &http.Client{}
	uri := fmt.Sprintf("https://api.leadspedia.com/core/v2/inboundCalls/getAll.do?api_key=%s&api_secret=%s&fromDate=%s&start=%d&limit=%d", env.API_KEY, env.API_SECRET, req.From, req.Start, req.Limit)
	request, _ := http.NewRequest("GET", uri, nil)
	res, e := client.Do(request)
	if e != nil {
		return nil, e
	}

	var data types.AllCallsResponse
	err := json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		log.Fatalln("failed to parse call data: ", err)
	}

	return &data, nil
}
