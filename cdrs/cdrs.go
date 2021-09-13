package cdrs

import (
	"bufio"
	"cdrsender/ilog"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrInvalidCdr = errors.New("invalid CDR")
)

type SCdr struct {
	// File path, name and position where this CDR was read from.
	FilePath     string
	Filename     string
	FilePosition int64

	// CDR data
	Data []byte

	// Original CDR length
	Length int
}

// *****************************************************************************************************

func NewCdr(cdrFile *SCdrFile) *SCdr {

	if cdrFile.CurrentFile == "" {
		return nil
	}

	return &SCdr{
		FilePath:     cdrFile.Path,
		Filename:     cdrFile.CurrentFile,
		FilePosition: cdrFile.CurrentPosition,
	}
}

// *****************************************************************************************************

func (cdr *SCdr) ReadFromFile() (eof bool, err error) {

	eof = false

	filename, err := filepath.Abs(cdr.FilePath + "/" + cdr.Filename)
	if err != nil {
		ilog.Log(ilog.ERR, "SCdr::ReadFromFile, cannot convert file name to absolute: %s", err.Error())

		// Continue with original file name and path
		filename = cdr.FilePath + "/" + cdr.Filename
	}

	file, err := os.Open(filename)
	if err != nil {
		ilog.Log(ilog.ERR, "SCdr::ReadFromFile, cannot open file %s: %s", cdr.Filename, err.Error())
		return
	}

	// Close the file before exiting this function in any case
	defer file.Close()

	_, err = file.Seek(cdr.FilePosition, 0)
	if err != nil {
		ilog.Log(ilog.ERR, "SCdr::ReadFromFile, cannot seek to position %d, file %s: %s", cdr.FilePosition, cdr.Filename,
			err.Error())
		return
	}

	cdr.Data, err = bufio.NewReader(file).ReadBytes('\n')
	if err != nil {

		// Return empty CDR on error
		cdr.Data = cdr.Data[:0]

		if err == io.EOF {
			eof = true
			return
		}

		ilog.Log(ilog.ERR, "SCdr::ReadFromFile, error reading file %s at position %d: %s", cdr.Filename,
			cdr.FilePosition, err.Error())
		return
	}

	cdr.Length = len(cdr.Data)
	ilog.Log(ilog.INF, "SCdr::ReadFromFile, CDR (%d bytes) was successfuly read from file %s:%d",
		cdr.Length, cdr.Filename, cdr.FilePosition)

	// Remove trailing special symbols
	for pos := cdr.Length - 1; pos >= 0; pos-- {
		if cdr.Data[pos] > ' ' {
			cdr.Data = cdr.Data[0 : pos+1]
			break
		}
	}

	if len(cdr.Data) < 2 || cdr.Data[0] != '{' || cdr.Data[len(cdr.Data)-1] != '}' {
		ilog.Log(ilog.WRN, "SCdr::ReadFromFile, invalid CDR in file %s:%d", cdr.Filename, cdr.FilePosition)
		cdr.Data = cdr.Data[:0]
		err = ErrInvalidCdr
		return
	}

	return
}
