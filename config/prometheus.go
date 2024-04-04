package config

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	chiprometheus "github.com/toshi0607/chi-prometheus"
	"log"
)

func Prometheus(router *chi.Mux) {
	reg := prometheus.NewRegistry()
	if err := reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
		log.Fatal(err)
	}
	if err := reg.Register(collectors.NewGoCollector()); err != nil {
		log.Fatal(err)
	}
	prometheusMiddleware := chiprometheus.New("test")
	reg.MustRegister(prometheusMiddleware.Collectors()...)
	promHandler := promhttp.InstrumentMetricHandler(
		reg, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	)
	router.Use(prometheusMiddleware.Handler)
	router.Handle("/metrics", promHandler)
}
