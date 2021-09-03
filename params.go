package main

import (
	"CdrSender/ilog"
	"math"
	"os"

	"gopkg.in/ini.v1"
)

type SParams struct {
	// Application instance name
	InstanceName string

	// PC name
	PcName string

	// Source CDR path
	SrcCdrPath string

	// CDR file prefix
	CdrFilePrefix string

	// CDR receiver URL
	UrlCdrReceiver string

	// CDR check interval, milliseconds. If no CDR is read, the program
	// will sleep this number of milliseconds before attempting to read
	// CDR again.
	InvervalCdrCheck int

	// Switch to a new file on EOF interval, milliseconds. If no CDR can
	// be read from the current file because of EOF, the program will wait
	// this number of milliseconds before trying to switch to a new file
	InvervalSwitchNewFile int

	// Interval until next CDR reading attempt after a read error, milliseconds
	IntervalErrorReread int

	// How many times to try to re-read CDR if we got any other error but
	// EOF
	CdrRereadErrorCount int

	// Statistics update interval, seconds
	IntervalStatsUpdate int

	// HTTP request retry interval, seconds
	IntervalHttpRetry int
}

func readIniString(section *ini.Section, keyName string, value *string) bool {
	*value = section.Key(keyName).String()
	if *value != "" {
		return true
	}

	ilog.Log(ilog.CRT, "SParams::readIniString, %s::%s is empty", section.Name(), keyName)
	return false
}

func (params *SParams) LoadFromIni(iniFile *ini.File) bool {

	pcName, err := os.Hostname()
	if err != nil {
		ilog.Log(ilog.CRT, "SParams::LoadFromIni, cannot get PC name: %s", err.Error())
		return false
	}

	params.PcName = pcName

	section := iniFile.Section("MAIN")

	if !readIniString(section, "InstanceName", &params.InstanceName) ||
		!readIniString(section, "CdrReceiverUrl", &params.UrlCdrReceiver) ||
		!readIniString(section, "SourceCdrPath", &params.SrcCdrPath) {

		return false
	}

	params.CdrFilePrefix = section.Key("CdrFilePrefix").String()
	params.InvervalCdrCheck = section.Key("CdrCheckInterval").RangeInt(100, 0, 3600000)
	params.InvervalSwitchNewFile = section.Key("SwitchToNewFileOnEofInterval").RangeInt(5000, 0, 3600000)
	params.IntervalErrorReread = section.Key("ErrorRereadInterval").RangeInt(1000, 100, 3600000)
	params.CdrRereadErrorCount = section.Key("ErrorRetryReadCount").RangeInt(5, 1, math.MaxInt32)
	params.IntervalStatsUpdate = section.Key("StatisticsUpdateInerval").RangeInt(60, 1, 300)
	params.IntervalHttpRetry = section.Key("HttpRetryInterval").RangeInt(10, 1, 300)

	return true
}
