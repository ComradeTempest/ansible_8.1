package main

import (
	"CdrSender/ilog"
	"encoding/json"
	"time"
)

func RequestGetStartPosition() (filename string, position int64) {

	jsonData, _ := json.Marshal(map[string]string{
		"host":     g_params.PcName,
		"instance": g_params.InstanceName,
	})

	url := g_params.UrlCdrReceiver + URI_PATH_GETSTARTPOS

	for nAttempt := 0; ; nAttempt++ {

		if nAttempt != 0 {
			time.Sleep(time.Second * time.Duration(g_params.IntervalHttpRetry))
		}

		ilog.Log(ilog.DBG, "RequestGetStartPosition, requesting %s (data: %s), attempt #%d", url, jsonData, nAttempt)

		requestStartTime := time.Now()
		/* response, err := http.Post(url, JSON_CONTENT_TYPE, bytes.NewReader(jsonData))
		if err != nil {
			ilog.Log(ilog.ERR, "RequestGetStartPosition, error requesting %s: %s", url, err.Error())
			continue
		}

		if response.StatusCode != 200 {
			ilog.Log(ilog.ERR, "RequestGetStartPosition, request failed (%d): %s", response.StatusCode, response.Status)
			response.Body.Close()
			continue
		}

		data, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			ilog.Log(ilog.ERR, "RequestGetStartPosition, error reading response body: %s", err.Error())
			continue
		} */
		var err error
		data := []byte("{\"startFile\":\"TTT\",\"startPosition\":100}")

		requestDuration := time.Since(requestStartTime)
		ilog.Log(ilog.DBG, "RequestGetStartPosition, response received in %d ms: %s", requestDuration.Milliseconds(),
			data)

		// Simple local structure to conveniently parse JSON response
		var respJson struct {
			Filename string `json:"startFile"`
			Position int64  `json:"startPosition"`
		}

		if err = json.Unmarshal(data, &respJson); err != nil {
			ilog.Log(ilog.ERR, "RequestGetStartPosition, invalid JSON received (%s): %s", data, err.Error())
			continue
		}

		filename = respJson.Filename
		position = respJson.Position
		ilog.Log(ilog.INF, "RequestGetStartPosition, start file is %s:%d",
			filename, position)
		return
	}
}
