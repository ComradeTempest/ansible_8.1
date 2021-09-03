package main

import (
	"CdrSender/cdrs"
	"CdrSender/ilog"
	"time"

	"gopkg.in/ini.v1"
)

var (
	g_params SParams

	g_cdrReadErrorCount int
	g_lastCdrTime       time.Time

	g_statistics SStats
)

// *****************************************************************************************************

func main() {

	iniFile, error := ini.Load(INI_FILE_NAME)
	if error != nil {
		return
	}

	ilog.SetDefaultLogger(ilog.NewFromIni(iniFile, APP_NAME))
	defer ilog.StopDefaultLogger()

	ilog.Log(ilog.INF, "main, starting %s v.%s", APP_NAME, APP_VERSION)
	if !g_params.LoadFromIni(iniFile) {
		return
	}

	ilog.Log(ilog.DBG, "main, application settings: %#v", g_params)

	cdrFile := cdrs.NewCdrFile(g_params.SrcCdrPath, g_params.CdrFilePrefix)
	cdrFile.CurrentFile, cdrFile.CurrentPosition = RequestGetStartPosition()

	//CdrPump(cdrFile)
}

// CdrProcessor returns only when it successfully processed previous CDR.
// If CDR cannot be processed for some reason (e.g. CDR receiver server
// is down, CdrProcessor will try again unless it succeeds).
// CdrProcessor returns file name and position to read the next CDR from.
func CdrProcessor(cdr *cdrs.SCdr) (nextFileName string, nextFilePos int64) {

	// Simulate CDR processing
	ilog.Log(ilog.DBG, "CdrProcessor, processing CDR %s", cdr.Data)
	time.Sleep(time.Second)
	nextFileName = cdr.Filename
	nextFilePos = cdr.FilePosition + int64(cdr.Length)
	return
}
