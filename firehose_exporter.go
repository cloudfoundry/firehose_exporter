package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	"github.com/cloudfoundry-community/firehose_exporter/collectors"
	"github.com/cloudfoundry-community/firehose_exporter/firehosenozzle"
	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/uaatokenrefresher"
)

var (
	listenAddress = flag.String(
		"web.listen-address", ":9114",
		"Address to listen on for web interface and telemetry.",
	)
	metricsPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose Prometheus metrics.",
	)
	showVersion = flag.Bool(
		"version", false,
		"Print version information.",
	)

	uaaUrl = flag.String(
		"uaa.url", "",
		"Cloud Foundry UAA URL.",
	)
	uaaClientID = flag.String(
		"uaa.client-id", "",
		"Cloud Foundry UAA Client ID.",
	)
	uaaClientSecret = flag.String(
		"uaa.client-secret", "",
		"Cloud Foundry UAA Client Secret.",
	)

	dopplerUrl = flag.String(
		"doppler.url", "",
		"Cloud Foundry Doppler URL.",
	)
	dopplerSubscriptionID = flag.String(
		"doppler.subscription-id", "prometheus",
		"Cloud Foundry Doppler Subscription ID.",
	)
	dopplerIdleTimeoutSeconds = flag.Uint(
		"doppler.idle-timeout-seconds", 5,
		"Cloud Foundry Doppler Idle Timeout (in seconds).",
	)

	skipSSLValidation = flag.Bool(
		"skip-ssl-verify", false,
		"Disable SSL Verify.",
	)
)

func init() {
	prometheus.MustRegister(version.NewCollector("firehose_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("firehose_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting firehose_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	authTokenRefresher, err := uaatokenrefresher.New(
		*uaaUrl,
		*uaaClientID,
		*uaaClientSecret,
		*skipSSLValidation,
	)
	if err != nil {
		log.Errorf("Error creating uaa client: %s", err.Error())
		os.Exit(1)
	}

	metricsStore := metrics.NewStore()

	nozzle := firehosenozzle.New(
		*dopplerUrl,
		*skipSSLValidation,
		*dopplerSubscriptionID,
		uint32(*dopplerIdleTimeoutSeconds),
		authTokenRefresher,
		metricsStore,
	)
	go func() {
		log.Fatal(nozzle.Start())
	}()

	internalMetricsCollector := collectors.NewInternalMetricsCollector(metricsStore)
	prometheus.MustRegister(internalMetricsCollector)

	containerMetricsCollector := collectors.NewContainerMetricsCollector(metricsStore)
	prometheus.MustRegister(containerMetricsCollector)

	counterMetricsCollector := collectors.NewCounterMetricsCollector(metricsStore)
	prometheus.MustRegister(counterMetricsCollector)

	valueMetricsCollector := collectors.NewValueMetricsCollector(metricsStore)
	prometheus.MustRegister(valueMetricsCollector)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Cloud Foundry Firehose Exporter</title></head>
             <body>
             <h1>Cloud Foundry Firehose Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
