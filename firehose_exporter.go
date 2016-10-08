package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

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
		"web.listen-address", ":9186",
		"Address to listen on for web interface and telemetry.",
	)
	metricsPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose Prometheus metrics.",
	)
	metricsNamespace = flag.String(
		"metrics.namespace", "firehose_exporter",
		"Metrics Namespace.",
	)
	metricsGarbage = flag.Duration(
		"metrics.garbage", 1*time.Minute,
		"How long to run the metrics garbage.",
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
	dopplerMetricExpiry = flag.Duration(
		"doppler.metric-expiry", 5*time.Minute,
		"How long a Cloud Foundry Doppler metric is valid.",
	)

	skipSSLValidation = flag.Bool(
		"skip-ssl-verify", false,
		"Disable SSL Verify.",
	)
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

func overrideWithEnvVar(name string, value *string) {
	envValue := os.Getenv(name)
	if envValue != "" {
		*value = envValue
	}
}

func overrideWithEnvUint(name string, value *uint) {
	envValue := os.Getenv(name)
	if envValue != "" {
		intValue, err := strconv.Atoi(envValue)
		if err != nil {
			log.Fatalf("Invalid `%s`: %s", name, err)
		}
		*value = uint(intValue)
	}
}

func overrideWithEnvDuration(name string, value *time.Duration) {
	envValue := os.Getenv(name)
	if envValue != "" {
		var err error
		*value, err = time.ParseDuration(envValue)
		if err != nil {
			log.Fatalf("Invalid `%s`: %s", name, err)
		}
	}
}

func overrideWithEnvBool(name string, value *bool) {
	envValue := os.Getenv(name)
	if envValue != "" {
		var err error
		*value, err = strconv.ParseBool(envValue)
		if err != nil {
			log.Fatalf("Invalid `%s`: %s", name, err)
		}
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("firehose_exporter"))
		os.Exit(0)
	}

	log.Infoln("Starting firehose_exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())

	overrideWithEnvVar("NOZZLE_WEB_LISTEN_ADDRESS", listenAddress)
	overrideWithEnvVar("NOZZLE_WEB_TELEMETRY_PATH", metricsPath)
	overrideWithEnvVar("NOZZLE_METRICS_NAMESPACE", metricsNamespace)
	overrideWithEnvDuration("NOZZLE_METRICS_GARBAGE", metricsGarbage)
	overrideWithEnvVar("NOZZLE_UAA_URL", uaaUrl)
	overrideWithEnvVar("NOZZLE_UAA_CLIENT_ID", uaaClientID)
	overrideWithEnvVar("NOZZLE_UAA_CLIENT_SECRET", uaaClientSecret)
	overrideWithEnvVar("NOZZLE_DOPPLER_URL", dopplerUrl)
	overrideWithEnvVar("NOZZLE_DOPPLER_SUBSCRIPTION_ID", dopplerSubscriptionID)
	overrideWithEnvUint("NOZZLE_DOPPLER_IDLE_TIMEOUT_SECONDS", dopplerIdleTimeoutSeconds)
	overrideWithEnvDuration("NOZZLE_DOPPLER_METRIC_EXPIRY", dopplerMetricExpiry)
	overrideWithEnvBool("NOZZLE_SKIP_SSL_VERIFY", skipSSLValidation)

	authTokenRefresher, err := uaatokenrefresher.New(
		*uaaUrl,
		*uaaClientID,
		*uaaClientSecret,
		*skipSSLValidation,
	)
	if err != nil {
		log.Errorf("Error creating UAA client: %s", err.Error())
		os.Exit(1)
	}

	metricsStore := metrics.NewStore(*metricsGarbage, *dopplerMetricExpiry)
	go metricsStore.Start()

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

	internalMetricsCollector := collectors.NewInternalMetricsCollector(*metricsNamespace, metricsStore)
	prometheus.MustRegister(internalMetricsCollector)

	containerMetricsCollector := collectors.NewContainerMetricsCollector(*metricsNamespace, metricsStore)
	prometheus.MustRegister(containerMetricsCollector)

	counterMetricsCollector := collectors.NewCounterMetricsCollector(*metricsNamespace, metricsStore)
	prometheus.MustRegister(counterMetricsCollector)

	valueMetricsCollector := collectors.NewValueMetricsCollector(*metricsNamespace, metricsStore)
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
