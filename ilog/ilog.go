// ilog is a logger that creates log files in IPN style
package ilog

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/ini.v1"
)

// Available log levels
const (
	CRT = iota
	ERR
	WRN
	INF
	DBG
)

// Logger parameters that must be specified by the user when creating logger instance
type SParams struct {
	Path                 string
	FilePrefix           string
	Level                int
	DeleteFilesAfterDays int
	SourceFilePos        bool
}

// Logger interface
type ILog interface {
	Start()
	Stop()
	Log(logLevel int, format string, params ...interface{})
}

// *****************************************************************************************************

// sLogRecord describes one log record
type sLogRecord struct {
	// Record date and time
	m_dt time.Time

	// Record severity as string (CRT, ERR, etc.)
	m_severity string

	// Source file name and position of the logging code line (if enabled in settings)
	m_sourceFilePos string

	m_text string
}

type sLogger struct {
	// Logger parameters
	m_params SParams

	// Last log record date in the YYYYMMDD format. It is used
	// to check whether we need to create a new log file
	m_currentLogDateStr string

	// Last log file name
	m_currentLogFileName string

	// Log file handle
	m_logFile *os.File

	// Log records channel
	m_logRecordsChannel chan sLogRecord

	// Logger worker finished flag
	m_loggerFinished sync.WaitGroup
}

var (
	g_logLevelToString = [...]string{"CRT", "ERR", "WRN", "INF", "DBG"}
	g_stringToLogLevel = [...]string{"critical", "error", "warning", "info", "debug"}

	// Default logger
	g_defaultLogger ILog = nil
)

// *****************************************************************************************************

func New(params SParams) ILog {

	// Check params
	if params.Level > DBG {
		params.Level = DBG
	}

	if params.DeleteFilesAfterDays < 2 {
		params.DeleteFilesAfterDays = 2
	}

	logger := sLogger{m_params: params}
	logger.Start()

	return &logger
}

func NewFromIni(iniFile *ini.File, prefix string) ILog {
	params := sParamsFromIni(iniFile, prefix)
	return New(params)
}

func SetDefaultLogger(logger ILog) {

	if g_defaultLogger == logger {
		return
	}

	if g_defaultLogger != nil {
		g_defaultLogger.Stop()
	}

	g_defaultLogger = logger
}

func StopDefaultLogger() {
	SetDefaultLogger(nil)
}

func Log(logLevel int, format string, params ...interface{}) {
	if g_defaultLogger != nil {
		g_defaultLogger.Log(logLevel|0x100, format, params...)
	}
}

// *****************************************************************************************************

func sParamsDefault() SParams {
	return SParams{
		Path:                 ".",
		FilePrefix:           "AppLog",
		Level:                DBG,
		DeleteFilesAfterDays: 30,
		SourceFilePos:        true,
	}
}

func sParamsLogLevelFromString(logLevel string) int {
	// First, try symbolic names
	logLevel = strings.ToLower(logLevel)
	for i, val := range g_stringToLogLevel {
		if logLevel == val {
			return i
		}
	}

	res, err := strconv.Atoi(logLevel)

	// Debug is the default log level
	if err != nil || res > DBG {
		return DBG
	}

	if res < CRT {
		return CRT
	}

	return int(res)
}

func sParamsFromIni(iniFile *ini.File, prefix string) (res SParams) {

	res = sParamsDefault()
	section := iniFile.Section("LOGGER")
	if section == nil {
		return res
	}

	res.Path = section.Key("Path").String()
	if res.Path == "" {
		res.Path = "."
	}

	res.FilePrefix = prefix
	res.Level = sParamsLogLevelFromString(section.Key("Level").String())
	res.DeleteFilesAfterDays = section.Key("DeleteFilesAfterDays").RangeInt(30, 2, 10000)
	res.SourceFilePos, _ = section.Key("SourceFilePosition").Bool()

	return res
}

func (l *sLogger) Start() {

	// Check whether logger is already started - if channel is created, logger is started
	if l.m_logRecordsChannel != nil {
		return
	}

	// Create target log folder
	os.MkdirAll(l.m_params.Path, 0755)

	// Create channel and start logger
	l.m_logRecordsChannel = make(chan sLogRecord, 100)
	l.m_loggerFinished.Add(1)
	go l.Worker()

	l.Log(INF, "SLogger::Start, logging started")
	l.Log(DBG, "SLogger::Start, logging parameters: %#v", l.m_params)
}

func (l *sLogger) Stop() {

	if l.m_logRecordsChannel == nil {
		return
	}

	l.Log(INF, "Slogger::Stop, logging stopped")
	close(l.m_logRecordsChannel)

	// Here we must wait until the logger worker thread finishes its work and
	// closes the output log file correctly
	l.m_loggerFinished.Wait()
	l.m_logRecordsChannel = nil
}

func (l *sLogger) CloseLogFile() {

	if l.m_logFile != nil {
		l.m_logFile.Close()
		l.m_logFile = nil
	}
}

func (l *sLogger) DeleteOldLogFiles() error {

	mask, err := filepath.Abs(l.m_params.Path + "/" + l.m_params.FilePrefix + "_*.log")
	if err != nil {
		return err
	}

	files, err := filepath.Glob(mask)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, fileName := range files {

		var year, month, day int
		n, err := fmt.Sscanf(filepath.Base(fileName), l.m_params.FilePrefix+"_%04d%02d%02d", &year, &month, &day)
		if err != nil || n != 3 {
			continue
		}

		fileDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
		fileAge := int(now.Sub(fileDate).Seconds()) / 86400
		if fileAge > l.m_params.DeleteFilesAfterDays {
			os.Remove(fileName)
		}
	}

	return nil
}

func (l *sLogger) OpenLogFile(logTime time.Time) error {

	var err error
	newLogDateStr := logTime.Format("_20060102")

	// The file has not been opened yet
	if l.m_logFile == nil {

		l.m_currentLogDateStr = newLogDateStr
		l.m_currentLogFileName, _ = filepath.Abs(l.m_params.Path + "/" + l.m_params.FilePrefix +
			newLogDateStr + logTime.Format("_150405.log"))

		l.DeleteOldLogFiles()
		l.m_logFile, err = os.OpenFile(l.m_currentLogFileName, os.O_WRONLY|os.O_CREATE, 0644)
		return err
	}

	// The file is open and its date is the same as the current
	if newLogDateStr == l.m_currentLogDateStr {
		return nil
	}

	// Create new log file. Its name must have 000000 as time suffix
	newLogFileName, _ := filepath.Abs(l.m_params.Path + "/" + l.m_params.FilePrefix +
		newLogDateStr + "_000000.log")

	// Write new file name to the current log file
	l.m_logFile.WriteString("----------------------> " + newLogFileName)

	// Close current log file
	l.CloseLogFile()

	l.DeleteOldLogFiles()
	l.m_logFile, err = os.OpenFile(newLogFileName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	// Write previous log file name to the new log file
	l.m_logFile.WriteString("<---------------------- " + l.m_currentLogFileName)

	l.m_currentLogDateStr = newLogDateStr
	l.m_currentLogFileName = newLogFileName
	return nil
}

func (l *sLogger) Log(logLevel int, format string, params ...interface{}) {

	// Log level contains hidden caller depth in the bits 8 and above
	callerDepth := 1 + (logLevel >> 8)
	logLevel &= 0xFF

	// Check log level
	if logLevel < 0 || logLevel > l.m_params.Level {
		return
	}

	var caller string
	if l.m_params.SourceFilePos {
		if _, goFile, line, ok := runtime.Caller(callerDepth); ok {
			caller = fmt.Sprintf("%s:%d", filepath.Base(goFile), line)
		}
	}

	// Write new log record structure to the log channel
	l.m_logRecordsChannel <- sLogRecord{
		m_dt:            time.Now().Round(0),
		m_severity:      g_logLevelToString[logLevel],
		m_sourceFilePos: caller,
		m_text:          fmt.Sprintf(format, params...),
	}
}

func (l *sLogger) Worker() {

	for rec := range l.m_logRecordsChannel {

		// Open new log file if needed
		l.OpenLogFile(rec.m_dt)

		// Skip this record if there is some problem with the log file
		if l.m_logFile == nil {
			continue
		}

		str := rec.m_dt.Format("2006.01.02 15:04:05.000") + " [" + rec.m_severity + "] "
		if rec.m_sourceFilePos != "" {
			str += "(" + rec.m_sourceFilePos + ") "
		}
		str += rec.m_text + "\n"

		l.m_logFile.WriteString(str)
	}

	// Close log file
	l.CloseLogFile()
	l.m_loggerFinished.Done()
}
