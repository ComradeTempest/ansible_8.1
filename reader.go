package main

import (
	"cdrsender/cdrs"
	"cdrsender/ilog"
	"time"
)

func CdrRead(cdrFile *cdrs.SCdrFile) *cdrs.SCdr {

	cdr := cdrs.NewCdr(cdrFile)
	eof, err := cdr.ReadFromFile()

	// Any error but EOF?
	if err != nil && !eof {

		// Increment error counter and check it against maximum error count
		g_cdrReadErrorCount++
		if g_cdrReadErrorCount < g_params.CdrRereadErrorCount {
			time.Sleep(time.Millisecond * time.Duration(g_params.IntervalErrorReread))
			return nil
		}

		g_cdrReadErrorCount = 0

		// OK, this file seems to be corrupted, switch to the next one
		if cdrFile.NextFile() {
			// The actual reading attempt will be made on the next CdrRead() call
			ilog.Log(ilog.INF, "CdrRead, %d error(s), switching to the next file: %s", g_params.CdrRereadErrorCount,
				cdrFile.CurrentFile)
			return nil
		}

		// If there is no next file, we have nothing more to do but try to re-read the
		// current file again. At least, wait some time before it
		time.Sleep(time.Millisecond * time.Duration(g_params.IntervalErrorReread))
		return nil
	}

	// EOF is not an error, so reset error counter
	g_cdrReadErrorCount = 0
	now := time.Now()

	// Is it EOF?
	if eof {

		// Do nothing within some interval after the last successful CDR read
		cdrInterval := now.Sub(g_lastCdrTime).Milliseconds()
		if cdrInterval < int64(g_params.InvervalSwitchNewFile) {
			time.Sleep(time.Millisecond * time.Duration(g_params.InvervalCdrCheck))
			return nil
		}

		// Try switching to a new file. Update g_lastCdrTime so if the switching does not succeed now,
		// we won't try switching too often
		g_lastCdrTime = now
		newCdrFile := *cdrFile

		// No new file? Return
		if !newCdrFile.NextFile() {
			time.Sleep(time.Millisecond * time.Duration(g_params.InvervalCdrCheck))
			return nil
		}

		// Try read a CDR from the new file
		cdr = cdrs.NewCdr(&newCdrFile)
		_, err = cdr.ReadFromFile()

		// Any error? We don't switch to the new file now
		if err != nil {
			return nil
		}

		// Switch to the new file permanently
		*cdrFile = newCdrFile
		ilog.Log(ilog.INF, "CdrRead, switched to the next file: %s", cdrFile.CurrentFile)
		return cdr
	}

	// We've read a CDR
	g_lastCdrTime = now
	return cdr
}

func CdrPump(cdrFile *cdrs.SCdrFile) {

	ilog.Log(ilog.INF, "CdrPump, starting with path %s", cdrFile.Path)

	// Initialize cdrFile if it is not already initialized
	cdrFile.FirstFile()
	ilog.Log(ilog.INF, "CdrPump, starting with file %s at position %d", cdrFile.CurrentFile, cdrFile.CurrentPosition)

	g_cdrReadErrorCount = 0
	g_lastCdrTime = time.Now()

	lastStatUpdateTime := time.Now()
	nCdrs := 0
	totalProcessingTime := time.Duration(0)

	// Read CDRs in an endless loop
	for {

		cdr := CdrRead(cdrFile)
		if cdr != nil {
			cdrReadTime := time.Now()
			g_statistics.OnCdrRead(cdrReadTime, cdrFile.CurrentFile, cdrFile.CurrentPosition)

			// Process CDR and update current file name and position
			cdrFile.SetNewFilePosition(CdrProcessor(cdr))

			cdrProcessedTime := time.Now()
			g_statistics.OnCdrProcessed(cdrProcessedTime, cdr.Length)

			totalProcessingTime += cdrProcessedTime.Sub(cdrReadTime)
			nCdrs++
		}

		// Do we need to update statistics?
		now := time.Now()
		statInterval := now.Sub(lastStatUpdateTime)
		if statInterval >= time.Second*time.Duration(g_params.IntervalStatsUpdate) {

			// Count local statistics
			var avgProcessingTime float64
			avgSpeed := float64(nCdrs) / statInterval.Seconds()
			if nCdrs > 0 {
				avgProcessingTime = totalProcessingTime.Seconds() / float64(nCdrs)
			}

			g_statistics.OnAvgValues(avgProcessingTime, avgSpeed)

			ilog.Log(ilog.DBG,
				"CdrPump, statistics: %d CDRs processed, avg. speed %f CDRs/s, avg. processing time %f ms",
				nCdrs, avgSpeed, avgProcessingTime*1000.0)

			lastStatUpdateTime = now
			nCdrs = 0
			totalProcessingTime = time.Duration(0)
		}
	}
}
