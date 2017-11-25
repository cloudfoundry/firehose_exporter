package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"

	"github.com/bosh-prometheus/firehose_exporter/collectors"
	"github.com/bosh-prometheus/firehose_exporter/filters"
	"github.com/bosh-prometheus/firehose_exporter/firehosenozzle"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/uaatokenrefresher"
)

var (
	uaaUrl = flag.String(
		"uaa.url", "",
		"Cloud Foundry UAA URL ($FIREHOSE_EXPORTER_UAA_URL).",
	)

	uaaClientID = flag.String(
		"uaa.client-id", "",
		"Cloud Foundry UAA Client ID ($FIREHOSE_EXPORTER_UAA_CLIENT_ID).",
	)

	uaaClientSecret = flag.String(
		"uaa.client-secret", "",
		"Cloud Foundry UAA Client Secret ($FIREHOSE_EXPORTER_UAA_CLIENT_SECRET).",
	)

	dopplerUrl = flag.String(
		"doppler.url", "",
		"Cloud Foundry Doppler URL ($FIREHOSE_EXPORTER_DOPPLER_URL).",
	)

	dopplerSubscriptionID = flag.String(
		"doppler.subscription-id", "prometheus",
		"Cloud Foundry Doppler Subscription ID ($FIREHOSE_EXPORTER_DOPPLER_SUBSCRIPTION_ID).",
	)

	dopplerIdleTimeout = flag.Duration(
		"doppler.idle-timeout", 0,
		"Cloud Foundry Doppler Idle Timeout duration ($FIREHOSE_EXPORTER_DOPPLER_IDLE_TIMEOUT).",
	)

	dopplerMinRetryDelay = flag.Duration(
		"doppler.min-retry-delay", 0,
		"Cloud Foundry Doppler min retry delay duration ($FIREHOSE_EXPORTER_DOPPLER_MIN_RETRY_DELAY).",
	)

	dopplerMaxRetryDelay = flag.Duration(
		"doppler.max-retry-delay", 0,
		"Cloud Foundry Doppler max retry delay duration ($FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_DELAY).",
	)

	dopplerMaxRetryCount = flag.Int(
		"doppler.max-retry-count", 0,
		"Cloud Foundry Doppler max retry count ($FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_COUNT).",
	)

	dopplerMetricExpiration = flag.Duration(
		"doppler.metric-expiration", 5*time.Minute,
		"How long a Cloud Foundry Doppler metric is valid ($FIREHOSE_EXPORTER_DOPPLER_METRIC_EXPIRATION).",
	)

	filterDeployments = flag.String(
		"filter.deployments", "",
		"Comma separated deployments to filter ($FIREHOSE_EXPORTER_FILTER_DEPLOYMENTS)",
	)

	filterEvents = flag.String(
		"filter.events", "",
		"Comma separated events to filter (ContainerMetric,CounterEvent,ValueMetric) ($FIREHOSE_EXPORTER_FILTER_EVENTS).",
	)

	metricsNamespace = flag.String(
		"metrics.namespace", "firehose",
		"Metrics Namespace ($FIREHOSE_EXPORTER_METRICS_NAMESPACE).",
	)

	metricsEnvironment = flag.String(
		"metrics.environment", "",
		"Environment label to be attached to metrics ($FIREHOSE_EXPORTER_METRICS_ENVIRONMENT).",
	)

	metricsCleanupInterval = flag.Duration(
		"metrics.cleanup-interval", 2*time.Minute,
		"Metrics clean up interval ($FIREHOSE_EXPORTER_METRICS_CLEANUP_INTERVAL).",
	)

	showVersion = flag.Bool(
		"version", false,
		"Print version information.",
	)

	skipSSLValidation = flag.Bool(
		"skip-ssl-verify", false,
		"Disable SSL Verify ($FIREHOSE_EXPORTER_SKIP_SSL_VERIFY).",
	)

	listenAddress = flag.String(
		"web.listen-address", ":9186",
		"Address to listen on for web interface and telemetry ($FIREHOSE_EXPORTER_WEB_LISTEN_ADDRESS).",
	)

	metricsPath = flag.String(
		"web.telemetry-path", "/metrics",
		"Path under which to expose Prometheus metrics ($FIREHOSE_EXPORTER_WEB_TELEMETRY_PATH).",
	)

	authUsername = flag.String(
		"web.auth.username", "",
		"Username for web interface basic auth ($FIREHOSE_EXPORTER_WEB_AUTH_USERNAME).",
	)

	authPassword = flag.String(
		"web.auth.password", "",
		"Password for web interface basic auth ($FIREHOSE_EXPORTER_WEB_AUTH_PASSWORD).",
	)

	tlsCertFile = flag.String(
		"web.tls.cert_file", "",
		"Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($FIREHOSE_EXPORTER_WEB_TLS_CERTFILE).",
	)

	tlsKeyFile = flag.String(
		"web.tls.key_file", "",
		"Path to a file that contains the TLS private key (PEM format) ($FIREHOSE_EXPORTER_WEB_TLS_KEYFILE).",
	)
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

func overrideFlagsWithEnvVars() {
	overrideWithEnvVar("FIREHOSE_EXPORTER_UAA_URL", uaaUrl)
	overrideWithEnvVar("FIREHOSE_EXPORTER_UAA_CLIENT_ID", uaaClientID)
	overrideWithEnvVar("FIREHOSE_EXPORTER_UAA_CLIENT_SECRET", uaaClientSecret)
	overrideWithEnvVar("FIREHOSE_EXPORTER_DOPPLER_URL", dopplerUrl)
	overrideWithEnvVar("FIREHOSE_EXPORTER_DOPPLER_SUBSCRIPTION_ID", dopplerSubscriptionID)
	overrideWithEnvDuration("FIREHOSE_EXPORTER_DOPPLER_IDLE_TIMEOUT", dopplerIdleTimeout)
	overrideWithEnvDuration("FIREHOSE_EXPORTER_DOPPLER_MIN_RETRY_DELAY", dopplerMinRetryDelay)
	overrideWithEnvDuration("FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_DELAY", dopplerMaxRetryDelay)
	overrideWithEnvInt("FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_COUNT", dopplerMaxRetryCount)
	overrideWithEnvDuration("FIREHOSE_EXPORTER_DOPPLER_METRIC_EXPIRATION", dopplerMetricExpiration)
	overrideWithEnvVar("FIREHOSE_EXPORTER_FILTER_DEPLOYMENTS", filterDeployments)
	overrideWithEnvVar("FIREHOSE_EXPORTER_FILTER_EVENTS", filterEvents)
	overrideWithEnvVar("FIREHOSE_EXPORTER_METRICS_NAMESPACE", metricsNamespace)
	overrideWithEnvVar("FIREHOSE_EXPORTER_METRICS_ENVIRONMENT", metricsEnvironment)
	overrideWithEnvDuration("FIREHOSE_EXPORTER_METRICS_CLEANUP_INTERVAL", metricsCleanupInterval)
	overrideWithEnvBool("FIREHOSE_EXPORTER_SKIP_SSL_VERIFY", skipSSLValidation)
	overrideWithEnvVar("FIREHOSE_EXPORTER_WEB_LISTEN_ADDRESS", listenAddress)
	overrideWithEnvVar("FIREHOSE_EXPORTER_WEB_TELEMETRY_PATH", metricsPath)
	overrideWithEnvVar("FIREHOSE_EXPORTER_WEB_AUTH_USERNAME", authUsername)
	overrideWithEnvVar("FIREHOSE_EXPORTER_WEB_AUTH_PASSWORD", authPassword)
	overrideWithEnvVar("FIREHOSE_EXPORTER_WEB_TLS_CERTFILE", tlsCertFile)
	overrideWithEnvVar("FIREHOSE_EXPORTER_WEB_TLS_KEYFILE", tlsKeyFile)
}

func overrideWithEnvVar(name string, value *string) {
	envValue := os.Getenv(name)
	if envValue != "" {
		*value = envValue
	}
}

func overrideWithEnvInt(name string, value *int) {
	envValue := os.Getenv(name)
	if envValue != "" {
		intValue, err := strconv.Atoi(envValue)
		if err != nil {
			log.Fatalf("Invalid `%s`: %s", name, err)
		}
		*value = int(intValue)
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

type basicAuthHandler struct {
	handler  http.HandlerFunc
	username string
	password string
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != h.username || password != h.password {
		log.Errorf("Invalid HTTP auth from `%s`", r.RemoteAddr)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"metrics\"")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	h.handler(w, r)
	return
}

func prometheusHandler() http.Handler {
	handler := prometheus.Handler()

	if *authUsername != "" && *authPassword != "" {
		handler = &basicAuthHandler{
			handler:  prometheus.Handler().ServeHTTP,
			username: *authUsername,
			password: *authPassword,
		}
	}

	return handler
}

func main() {
	flag.Parse()
	overrideFlagsWithEnvVars()

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
		log.Errorf("Error creating UAA client: %s", err.Error())
		os.Exit(1)
	}

	var deployments []string
	if *filterDeployments != "" {
		deployments = strings.Split(*filterDeployments, ",")
	}
	deploymentFilter := filters.NewDeploymentFilter(deployments)

	var events []string
	if *filterEvents != "" {
		events = strings.Split(*filterEvents, ",")
	}
	eventFilter, err := filters.NewEventFilter(events)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	metricsStore := metrics.NewStore(*dopplerMetricExpiration, *metricsCleanupInterval, deploymentFilter, eventFilter)

	nozzle := firehosenozzle.New(
		*dopplerUrl,
		*skipSSLValidation,
		*dopplerSubscriptionID,
		*dopplerIdleTimeout,
		*dopplerMinRetryDelay,
		*dopplerMaxRetryDelay,
		*dopplerMaxRetryCount,
		authTokenRefresher,
		metricsStore,
	)
	go func() {
		log.Fatal(nozzle.Start())
	}()

	internalMetricsCollector := collectors.NewInternalMetricsCollector(*metricsNamespace, *metricsEnvironment, metricsStore)
	prometheus.MustRegister(internalMetricsCollector)

	containerMetricsCollector := collectors.NewContainerMetricsCollector(*metricsNamespace, *metricsEnvironment, metricsStore)
	prometheus.MustRegister(containerMetricsCollector)

	counterEventsCollector := collectors.NewCounterEventsCollector(*metricsNamespace, *metricsEnvironment, metricsStore)
	prometheus.MustRegister(counterEventsCollector)

	httpStartStopCollector := collectors.NewHttpStartStopCollector(*metricsNamespace, *metricsEnvironment, metricsStore)
	prometheus.MustRegister(httpStartStopCollector)

	valueMetricsCollector := collectors.NewValueMetricsCollector(*metricsNamespace, *metricsEnvironment, metricsStore)
	prometheus.MustRegister(valueMetricsCollector)

	handler := prometheusHandler()
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Cloud Foundry Firehose Exporter</title></head>
             <body>
             <h1>Cloud Foundry Firehose Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	if *tlsCertFile != "" && *tlsKeyFile != "" {
		log.Infoln("Listening TLS on", *listenAddress)
		log.Fatal(http.ListenAndServeTLS(*listenAddress, *tlsCertFile, *tlsKeyFile, nil))
	} else {
		log.Infoln("Listening on", *listenAddress)
		log.Fatal(http.ListenAndServe(*listenAddress, nil))
	}
}
