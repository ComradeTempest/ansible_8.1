package main

import (
	"CdrSender/ilog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometeus metrics
type SStats struct {
	// Last time a CDR was read from a file
	lastCdrReadTime prometheus.Gauge

	// File position the last CDR was read from
	lastCdrFilePos prometheus.Gauge

	// Last time when a CDR was processed
	lastCdrProcessTime prometheus.Gauge

	// CDR processing summary
	cdrProcessingSummary prometheus.Summary

	// Average CDR processing time
	avgCdrProcessingTime prometheus.Gauge

	// Average CDR processing speed
	avgCdrProcessingSpeed prometheus.Gauge

	// Cdr processing errors
	cdrProcessingErrorsCount prometheus.Counter
}

func (stat *SStats) OnCdrRead(tm time.Time, filename string, position int64) {
	stat.lastCdrReadTime.Set(float64(tm.UnixNano()) / 1e9)
	stat.lastCdrFilePos.Set(float64(position))
}

func (stat *SStats) OnCdrProcessed(tm time.Time, cdrLength int) {
	stat.lastCdrProcessTime.Set(float64(tm.UnixNano()) / 1e9)
	stat.cdrProcessingSummary.Observe(float64(cdrLength))
}

func (stat *SStats) OnCdrProcessingError() {
	stat.cdrProcessingErrorsCount.Inc()
}

func (stat *SStats) OnAvgValues(avgProcessingTime, avgSpeed float64) {
	stat.avgCdrProcessingTime.Set(avgProcessingTime)
	stat.avgCdrProcessingSpeed.Set(avgSpeed)
}

func (stat *SStats) Init() {

	stat.lastCdrReadTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cdr_last_read_time",
		Help: "Last time a CDR was read from a file",
	})
	prometheus.MustRegister(stat.lastCdrReadTime)

	stat.lastCdrFilePos = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cdr_last_read_position",
		Help: "File position the last CDR was read from",
	})
	prometheus.MustRegister(stat.lastCdrFilePos)

	stat.lastCdrProcessTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cdr_last_processed_time",
		Help: "Last time when a CDR was processed",
	})
	prometheus.MustRegister(stat.lastCdrProcessTime)

	stat.cdrProcessingSummary = prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "cdr_processing_summary",
		MaxAge:     time.Second * 60,
		AgeBuckets: 1,
	})
	prometheus.MustRegister(stat.cdrProcessingSummary)

	stat.avgCdrProcessingTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cdr_avr_processing_time",
		Help: "Average CDR processing time, seconds",
	})
	prometheus.MustRegister(stat.avgCdrProcessingTime)

	stat.avgCdrProcessingSpeed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cdr_avr_processing_speed",
		Help: "Average CDR processing speed, CDRs/s",
	})
	prometheus.MustRegister(stat.avgCdrProcessingSpeed)

	stat.cdrProcessingErrorsCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cdr_processing_errors",
		Help: "Amount of CDR processing errors",
	})
	prometheus.MustRegister(stat.cdrProcessingErrorsCount)

	http.Handle("/metrics", promhttp.Handler())

	// Start Prometheus HTTP server
	go func() {
		err := http.ListenAndServe(g_params.PrometheusHttpUrl, nil)
		if err != nil {
			ilog.Log(ilog.ERR, "SStats::Init::func, cannot start HTTP server: %s", err.Error())
		}
	}()
}
