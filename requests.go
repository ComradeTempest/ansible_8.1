package main

import (
	"cdrsender/cdrs"
	"cdrsender/ilog"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

func DoPostRequestJson(url string, data []byte) (result []byte) {

	for nAttempt := 0; ; nAttempt++ {

		if nAttempt != 0 {
			g_statistics.OnCdrProcessingError()
			time.Sleep(time.Second * time.Duration(g_params.IntervalHttpRetry))
		}

		ilog.Log(ilog.DBG, "DoPostRequest, requesting %s (data: %s), attempt #%d", url, data, nAttempt)

		requestStartTime := time.Now()
		response, err := http.Post(url, JSON_CONTENT_TYPE, bytes.NewReader(data))
		if err != nil {
			ilog.Log(ilog.ERR, "DoPostRequest, error requesting %s: %s", url, err.Error())
			continue
		}

		if response.StatusCode != 200 {
			ilog.Log(ilog.ERR, "DoPostRequest, request failed (%d): %s", response.StatusCode, response.Status)
			response.Body.Close()
			continue
		}

		result, err = io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			ilog.Log(ilog.ERR, "DoPostRequest, error reading response body: %s", err.Error())
			continue
		}

		requestDuration := time.Since(requestStartTime)
		ilog.Log(ilog.DBG, "DoPostRequest, response received in %d ms: %s", requestDuration.Milliseconds(),
			strings.TrimSpace(string(result)))
		return
	}
}

func RequestGetStartPosition() (filename string, position int64) {

	requestData := SRequestStartPos{
		Host:     g_params.PcName,
		Instance: g_params.InstanceName,
	}

	jsonData, _ := json.Marshal(requestData)
	url := g_params.UrlCdrReceiver + URI_PATH_GETSTARTPOS
	for {

		data := DoPostRequestJson(url, jsonData)
		//data := []byte("{\"startFile\":\"TTT\",\"startPosition\":100}")

		var respJson SResponseJson
		if err := json.Unmarshal(data, &respJson); err != nil {
			ilog.Log(ilog.ERR, "RequestGetStartPosition, invalid JSON received (%s): %s", data, err.Error())
			continue
		}

		filename = respJson.Filename
		position = respJson.Position
		ilog.Log(ilog.INF, "RequestGetStartPosition, start file is %s:%d", filename, position)
		return
	}
}

// CdrProcessor returns only when it successfully processed previous CDR.
// If CDR cannot be processed for some reason (e.g. CDR receiver server
// is down, CdrProcessor will try again unless it succeeds).
// CdrProcessor returns file name and position to read the next CDR from.
func CdrProcessor(cdr *cdrs.SCdr) (nextFileName string, nextFilePos int64) {

	requestData := SProcessCdr{
		Host:         g_params.PcName,
		Instance:     g_params.InstanceName,
		Filename:     cdr.Filename,
		Position:     cdr.FilePosition,
		RecordLength: cdr.Length,
		Data:         string(cdr.Data),
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		ilog.Log(ilog.CRT, "CdrProcessor, cannot convert CDR (%#v) to JSON: %s", requestData, err.Error())
		panic("CdrProcessor, cannot convert CDR to JSON")
	}

	url := g_params.UrlCdrReceiver + URI_PATH_PROCESSCDR
	for {
		responseData := DoPostRequestJson(url, jsonData)

		var respJson SResponseJson
		if err := json.Unmarshal(responseData, &respJson); err != nil {
			ilog.Log(ilog.ERR, "CdrProcessor, invalid JSON received (%s): %s", responseData, err.Error())
			continue
		}

		nextFileName = respJson.Filename
		nextFilePos = respJson.Position
		ilog.Log(ilog.INF, "CdrProcessor, CDR %s:%d processed", cdr.Filename, cdr.FilePosition)
		return
	}
}
