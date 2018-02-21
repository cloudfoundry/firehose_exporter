package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bosh-prometheus/firehose_exporter/collectors"
	"github.com/bosh-prometheus/firehose_exporter/filters"
	"github.com/bosh-prometheus/firehose_exporter/firehosenozzle"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/uaatokenrefresher"
)

var (
	uaaUrl = kingpin.Flag(
		"uaa.url", "Cloud Foundry UAA URL ($FIREHOSE_EXPORTER_UAA_URL)",
	).Envar("FIREHOSE_EXPORTER_UAA_URL").Required().String()

	uaaClientID = kingpin.Flag(
		"uaa.client-id", "Cloud Foundry UAA Client ID ($FIREHOSE_EXPORTER_UAA_CLIENT_ID)",
	).Envar("FIREHOSE_EXPORTER_UAA_CLIENT_ID").Required().String()

	uaaClientSecret = kingpin.Flag(
		"uaa.client-secret", "Cloud Foundry UAA Client Secret ($FIREHOSE_EXPORTER_UAA_CLIENT_SECRET)",
	).Envar("FIREHOSE_EXPORTER_UAA_CLIENT_SECRET").Required().String()

	dopplerUrl = kingpin.Flag(
		"doppler.url", "Cloud Foundry Doppler URL ($FIREHOSE_EXPORTER_DOPPLER_URL)",
	).Envar("FIREHOSE_EXPORTER_DOPPLER_URL").Required().String()

	dopplerSubscriptionID = kingpin.Flag(
		"doppler.subscription-id", "Cloud Foundry Doppler Subscription ID ($FIREHOSE_EXPORTER_DOPPLER_SUBSCRIPTION_ID)",
	).Envar("FIREHOSE_EXPORTER_DOPPLER_SUBSCRIPTION_ID").Default("prometheus").String()

	dopplerIdleTimeout = kingpin.Flag(
		"doppler.idle-timeout", "Cloud Foundry Doppler Idle Timeout duration ($FIREHOSE_EXPORTER_DOPPLER_IDLE_TIMEOUT)",
	).Envar("FIREHOSE_EXPORTER_DOPPLER_IDLE_TIMEOUT").Default("0").Duration()

	dopplerMinRetryDelay = kingpin.Flag(
		"doppler.min-retry-delay", "Cloud Foundry Doppler min retry delay duration ($FIREHOSE_EXPORTER_DOPPLER_MIN_RETRY_DELAY)",
	).Envar("FIREHOSE_EXPORTER_DOPPLER_MIN_RETRY_DELAY").Default("0").Duration()

	dopplerMaxRetryDelay = kingpin.Flag(
		"doppler.max-retry-delay", "Cloud Foundry Doppler max retry delay duration ($FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_DELAY)",
	).Envar("FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_DELAY").Default("0").Duration()

	dopplerMaxRetryCount = kingpin.Flag(
		"doppler.max-retry-count", "Cloud Foundry Doppler max retry count ($FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_COUNT)",
	).Envar("FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_COUNT").Default("0").Int()

	dopplerMetricExpiration = kingpin.Flag(
		"doppler.metric-expiration", "How long a Cloud Foundry Doppler metric is valid ($FIREHOSE_EXPORTER_DOPPLER_METRIC_EXPIRATION)",
	).Envar("FIREHOSE_EXPORTER_DOPPLER_METRIC_EXPIRATION").Default("5m").Duration()

	filterDeployments = kingpin.Flag(
		"filter.deployments", "Comma separated deployments to filter ($FIREHOSE_EXPORTER_FILTER_DEPLOYMENTS)",
	).Envar("FIREHOSE_EXPORTER_FILTER_DEPLOYMENTS").Default("").String()

	filterEvents = kingpin.Flag(
		"filter.events", "Comma separated events to filter (ContainerMetric,CounterEvent,ValueMetric) ($FIREHOSE_EXPORTER_FILTER_EVENTS)",
	).Envar("FIREHOSE_EXPORTER_FILTER_EVENTS").Default("").String()

	metricsNamespace = kingpin.Flag(
		"metrics.namespace", "Metrics Namespace ($FIREHOSE_EXPORTER_METRICS_NAMESPACE)",
	).Envar("FIREHOSE_EXPORTER_METRICS_NAMESPACE").Default("firehose").String()

	metricsEnvironment = kingpin.Flag(
		"metrics.environment", "Environment label to be attached to metrics ($FIREHOSE_EXPORTER_METRICS_ENVIRONMENT)",
	).Envar("FIREHOSE_EXPORTER_METRICS_ENVIRONMENT").Required().String()

	metricsCleanupInterval = kingpin.Flag(
		"metrics.cleanup-interval", "Metrics clean up interval ($FIREHOSE_EXPORTER_METRICS_CLEANUP_INTERVAL)",
	).Envar("FIREHOSE_EXPORTER_METRICS_CLEANUP_INTERVAL").Default("2m").Duration()

	skipSSLValidation = kingpin.Flag(
		"skip-ssl-verify", "Disable SSL Verify ($FIREHOSE_EXPORTER_SKIP_SSL_VERIFY)",
	).Envar("FIREHOSE_EXPORTER_SKIP_SSL_VERIFY").Default("false").Bool()

	listenAddress = kingpin.Flag(
		"web.listen-address", "Address to listen on for web interface and telemetry ($FIREHOSE_EXPORTER_WEB_LISTEN_ADDRESS)",
	).Envar("FIREHOSE_EXPORTER_WEB_LISTEN_ADDRESS").Default(":9186").String()

	metricsPath = kingpin.Flag(
		"web.telemetry-path", "Path under which to expose Prometheus metrics ($FIREHOSE_EXPORTER_WEB_TELEMETRY_PATH)",
	).Envar("FIREHOSE_EXPORTER_WEB_TELEMETRY_PATH").Default("/metrics").String()

	authUsername = kingpin.Flag(
		"web.auth.username", "Username for web interface basic auth ($FIREHOSE_EXPORTER_WEB_AUTH_USERNAME)",
	).Envar("FIREHOSE_EXPORTER_WEB_AUTH_USERNAME").String()

	authPassword = kingpin.Flag(
		"web.auth.password", "Password for web interface basic auth ($FIREHOSE_EXPORTER_WEB_AUTH_PASSWORD)",
	).Envar("FIREHOSE_EXPORTER_WEB_AUTH_PASSWORD").String()

	tlsCertFile = kingpin.Flag(
		"web.tls.cert_file", "Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($FIREHOSE_EXPORTER_WEB_TLS_CERTFILE)",
	).Envar("FIREHOSE_EXPORTER_WEB_TLS_CERTFILE").ExistingFile()

	tlsKeyFile = kingpin.Flag(
		"web.tls.key_file", "Path to a file that contains the TLS private key (PEM format) ($FIREHOSE_EXPORTER_WEB_TLS_KEYFILE)",
	).Envar("FIREHOSE_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
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
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("firehose_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

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
		nozzle.Start()
		os.Exit(1)
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
