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

	iniFile, err := ini.Load(INI_FILE_NAME)
	if err != nil {
		return
	}

	ilog.SetDefaultLogger(ilog.NewFromIni(iniFile, APP_NAME))
	defer ilog.StopDefaultLogger()

	ilog.Log(ilog.INF, "main, starting %s v.%s", APP_NAME, APP_VERSION)
	if !g_params.LoadFromIni(iniFile) {
		return
	}

	ilog.Log(ilog.DBG, "main, application settings: %#v", g_params)
	g_statistics.Init()

	// Initalize SCdrFile with default values
	cdrFile := cdrs.NewCdrFile(g_params.SrcCdrPath, g_params.CdrFilePrefix)

	// Get read position from DB
	cdrFile.CurrentFile, cdrFile.CurrentPosition = RequestGetStartPosition()

	// Start CDR pump
	CdrPump(cdrFile)
}
