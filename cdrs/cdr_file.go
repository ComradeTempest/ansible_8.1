package cdrs

import (
	"cdrsender/ilog"
	"path/filepath"
	"sort"
	"time"
)

type SCdrFile struct {
	// Path to CDR files
	Path string

	// CDR file name prefix
	Prefix string

	// Current CDR file (without path, file name only)
	CurrentFile string

	// Current file position
	CurrentPosition int64
}

// *****************************************************************************************************

func NewCdrFile(path, prefix string) *SCdrFile {

	result := SCdrFile{
		Prefix: prefix,
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		ilog.Log(ilog.ERR, "cdrs::NewCdrFileList, cannot transform path %s to absolute: %s", path, err.Error())
		return nil
	}

	result.Path = absPath
	return &result
}

// *****************************************************************************************************

func (sfl *SCdrFile) SetNewFilePosition(filename string, position int64) {

	sfl.CurrentFile = filename
	sfl.CurrentPosition = position
}

func (sfl *SCdrFile) getFileList() []string {

	result, err := filepath.Glob(sfl.Path + "/" + sfl.Prefix + "*" + ".cdr")
	if err != nil {
		ilog.Log(ilog.ERR, "cdrs::getCdrFileList, cannot get file list of %s: %s", sfl.Path, err.Error())
		return nil
	}

	// Return nil pointer instead of an empty slice
	if len(result) == 0 {
		return nil
	}

	// Strip anything but file names from the file list
	for index, fileName := range result {
		result[index] = filepath.Base(fileName)
	}

	// Sort file list in ascending order
	sort.StringSlice(result).Sort()

	return result
}

func (sfl *SCdrFile) NextFile() bool {

	files := sfl.getFileList()
	if files == nil {
		return false
	}

	// If we are called for the first time, use the first file
	if sfl.CurrentFile == "" {
		sfl.CurrentFile = files[0]
		sfl.CurrentPosition = 0
		return true
	}

	// Try to find specified file in the file list
	index := sort.StringSlice(files).Search(sfl.CurrentFile)

	// If we cannot find it (the wrong file was specified?), use the first file from the list
	if index == len(files) || files[index] != sfl.CurrentFile {
		ilog.Log(ilog.WRN, "cdrs::GetNextCdrFile, file %s was not found in %s", sfl.CurrentFile, sfl.Path)
		sfl.CurrentFile = files[0]
		sfl.CurrentPosition = 0
		return true
	}

	// If this is the last file, there is no next file yet
	if index == len(files)-1 {
		return false
	}

	// Switch to the next file
	sfl.CurrentFile = files[index+1]
	sfl.CurrentPosition = 0
	return true
}

func (sfl *SCdrFile) Empty() bool {
	return sfl.CurrentFile == ""
}

func (sfl *SCdrFile) FirstFile() {

	// If SCdrFile is already initialized with a file name, do nothing
	if !sfl.Empty() {
		return
	}

	ilog.Log(ilog.INF, "SCdrFile::FirstFile, using first file in %s", sfl.Path)
	if sfl.NextFile() {
		return
	}

	ilog.Log(ilog.INF, "SCdrFile::FirstFile, no CDR files in %s, waiting...", sfl.Path)

	for {
		time.Sleep(time.Second * 3)
		if sfl.NextFile() {
			return
		}
	}
}
