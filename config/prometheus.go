package config

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type customMetrics struct {
	memoryTotal prometheus.Gauge
	threadCount prometheus.Gauge
	coreCount   prometheus.Gauge
}

func newCustomMetrics() customMetrics {
	newMetrics := &customMetrics{
		memoryTotal: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "memstats_sys_available_bytes",
			Help: "Number of bytes available from system.",
		}),
		threadCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpuinfo_sys_thread_total",
			Help: "Number of cpus available from system.",
		}),
		coreCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "cpuinfo_sys_cores_total",
			Help: "Number of cpus available from system.",
		}),
	}

	prometheus.Register(newMetrics.memoryTotal)
	prometheus.Register(newMetrics.threadCount)
	prometheus.Register(newMetrics.coreCount)

	return *newMetrics
}

func Prometheus(cfg Config, router *http.ServeMux) http.Handler {

	slokMiddleware := middleware.New(middleware.Config{
		Service:  cfg.Server.Service,
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})

	router.Handle(cfg.Server.Metrics, promhttp.Handler())

	handler := std.Handler("", slokMiddleware, router)

	custom := newCustomMetrics()

	memTotal := readProcInfoRegexNumber("meminfo", "MemTotal")
	custom.memoryTotal.Set(float64(memTotal) * 1000)

	threadCount := readProcInfoRegexCount("cpuinfo", "processor")
	custom.threadCount.Set(float64(threadCount))

	coreCount := readProcInfoRegexNumber("cpuinfo", "cpu cores")
	custom.coreCount.Set(float64(coreCount))

	return handler
}

func readProcInfoRegexCount(fname string, property string) int64 {
	path := fmt.Sprintf("/proc/%s", fname)

	file, err := os.ReadFile(path)
	if err != nil {
		return int64(-1)
	}

	file_str := string(file)
	rx := fmt.Sprintf("%s[ \\t]+:[ \\t]+(\\d+)", property)
	regex, err := regexp.Compile(rx)
	if err != nil {
		// handle error
	}

	if matches := regex.FindAllStringSubmatch(file_str, -1); matches != nil {
		return int64(len(matches))
	}

	return int64(-1)
}

func readProcInfoRegexNumber(fname string, property string) int64 {
	path := fmt.Sprintf("/proc/%s", fname)

	file, err := os.ReadFile(path)
	if err != nil {
		return int64(-1)
	}

	file_str := string(file)
	rx := fmt.Sprintf("%s:?[\\s]+:?[\\s]+(\\d+)", property)
	regex := regexp.MustCompile(rx)

	matches := regex.FindStringSubmatch(file_str)
	if len(matches) == 0 {
		return -1
	}

	number, err := strconv.Atoi(matches[1])
	if err != nil {
		return -4
	}

	return int64(number)
}
