package main

import (
	"sync"
	"time"
)

type SStats struct {
	// Unix time when last CDR was successfully read
	LastCdrReadUts int64

	// Last CDR file name and position
	LastCdrFileName     string
	LastCdrFilePosition int64

	// Unix time when last CDR was successfully processed
	LastCdrProcessedUts int64

	// Average CDR reading/processing speed, CDRs/s
	AverageCdrSpeed float64

	// Average CDR processing time, seconds
	AverageCdrProcessingTime float64

	// Mutex, private member
	mutex sync.Mutex
}

func (stat *SStats) Lock() {
	stat.mutex.Lock()
}

func (stat *SStats) Unlock() {
	stat.mutex.Unlock()
}

func (stat *SStats) SetLastCdrReadTime(tm time.Time, filename string, position int64) {
	stat.Lock()
	stat.LastCdrReadUts = tm.Unix()
	stat.LastCdrFileName = filename
	stat.LastCdrFilePosition = position
	stat.Unlock()
}

func (stat *SStats) SetLastCdrProcessedTime(tm time.Time) {
	stat.Lock()
	stat.LastCdrProcessedUts = tm.Unix()
	stat.Unlock()
}

func (stat *SStats) UpdateAverages(interval time.Duration, totalProcessing time.Duration, nCdrs int) {

	if interval == time.Duration(0) {
		return
	}

	stat.Lock()
	stat.AverageCdrSpeed = float64(nCdrs) / interval.Seconds()
	if nCdrs > 0 {
		stat.AverageCdrProcessingTime = totalProcessing.Seconds() / float64(nCdrs)
	} else {
		stat.AverageCdrProcessingTime = 0
	}
	stat.Unlock()
}

func (stat *SStats) GetStat() *SStats {
	stat.Lock()
	result := SStats{
		LastCdrReadUts:           stat.LastCdrReadUts,
		LastCdrFileName:          stat.LastCdrFileName,
		LastCdrFilePosition:      stat.LastCdrFilePosition,
		LastCdrProcessedUts:      stat.LastCdrProcessedUts,
		AverageCdrSpeed:          stat.AverageCdrSpeed,
		AverageCdrProcessingTime: stat.AverageCdrProcessingTime,
	}
	stat.Unlock()

	return &result
}

func (stat *SStats) Reset() {
	stat.Lock()
	stat.LastCdrReadUts = 0
	stat.LastCdrFileName = ""
	stat.LastCdrFilePosition = 0
	stat.LastCdrProcessedUts = 0
	stat.AverageCdrSpeed = 0
	stat.AverageCdrProcessingTime = 0
	stat.Unlock()
}
