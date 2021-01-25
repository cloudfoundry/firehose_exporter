package main

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"code.cloudfoundry.org/go-loggregator"
	"github.com/bosh-prometheus/firehose_exporter/collectors"
	"github.com/bosh-prometheus/firehose_exporter/metricmaker"
	"github.com/bosh-prometheus/firehose_exporter/metrics"
	"github.com/bosh-prometheus/firehose_exporter/nozzle"
	"github.com/prometheus/common/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	retroCompatDisable = kingpin.Flag("retro_compat.disable", "Disable retro compatibility",
	).Envar("FIREHOSE_EXPORTER_RETRO_COMPAT_DISABLE").Default("false").Bool()

	enableRetroCompatDelta = kingpin.Flag("retro_compat.enable_delta", "Enable retro compatibility delta in counter",
	).Envar("FIREHOSE_EXPORTER_RETRO_COMPAT_ENABLE_DELTA").Default("false").Bool()

	loggingURL = kingpin.Flag(
		"logging.url", "Cloud Foundry Logging endpoint ($FIREHOSE_EXPORTER_LOGGING_URL)",
	).Envar("FIREHOSE_EXPORTER_LOGGING_URL").Required().String()

	loggingTLSCa = kingpin.Flag(
		"logging.tls.ca", "Path to ca cert to connect to rlp",
	).Envar("FIREHOSE_EXPORTER_LOGGING_TLS_CA").Default("").String()

	loggingTLSCert = kingpin.Flag(
		"logging.tls.cert", "Path to cert to connect to rlp in mtls",
	).Envar("FIREHOSE_EXPORTER_LOGGING_TLS_CERT").Default("").String()

	loggingTLSKey = kingpin.Flag(
		"logging.tls.key", "Path to key to connect to rlp in mtls",
	).Envar("FIREHOSE_EXPORTER_LOGGING_TLS_KEY").Default("").String()

	metricsNamespace = kingpin.Flag(
		"metrics.namespace", "Metrics Namespace ($FIREHOSE_EXPORTER_METRICS_NAMESPACE)",
	).Envar("FIREHOSE_EXPORTER_METRICS_NAMESPACE").Default("firehose").String()

	metricsBatchSize = kingpin.Flag(
		"metrics.batch_size", "Batch size for nozzle envelop buffer ($FIREHOSE_EXPORTER_METRICS_NAMESPACE)",
	).Envar("FIREHOSE_EXPORTER_METRICS_BATCH_SIZE").Default("-1").Int()

	metricsShardId = kingpin.Flag(
		"metrics.shard_id", "The sharding group name to use for egress from RLP ($FIREHOSE_EXPORTER_SHARD_ID)",
	).Envar("FIREHOSE_EXPORTER_SHARD_ID").Default("firehose_exporter").String()

	metricsNodeIndex = kingpin.Flag(
		"metrics.node_index", "Node index to use ($FIREHOSE_EXPORTER_NODE_INDEX)",
	).Envar("FIREHOSE_EXPORTER_NODE_INDEX").Default("0").Int()

	metricsTimerRollup = kingpin.Flag(
		"metrics.timer_rollup_buffer_size", "The number of envelopes that will be allowed to be buffered while timer metric aggregations are running ($FIREHOSE_EXPORTER_TIMER_ROLLUP_BUFFER_SIZE)",
	).Envar("FIREHOSE_EXPORTER_TIMER_ROLLUP_BUFFER_SIZE").Default("16384").Uint()

	metricsEnvironment = kingpin.Flag(
		"metrics.environment", "Environment label to be attached to metrics ($FIREHOSE_EXPORTER_METRICS_ENVIRONMENT)",
	).Envar("FIREHOSE_EXPORTER_METRICS_ENVIRONMENT").Required().String()

	metricExpiration = kingpin.Flag(
		"metrics.expiration", "How long a Cloud Foundry metric is valid ($FIREHOSE_EXPORTER_METRICS_EXPIRATION)",
	).Envar("FIREHOSE_EXPORTER_METRICS_EXPIRATION").Default("10m").Duration()

	skipSSLValidation = kingpin.Flag(
		"skip-ssl-verify", "Disable SSL Verify ($FIREHOSE_EXPORTER_SKIP_SSL_VERIFY)",
	).Envar("FIREHOSE_EXPORTER_SKIP_SSL_VERIFY").Default("false").Bool()

	filterDeployments = kingpin.Flag(
		"filter.deployments", "Comma separated deployments to filter ($FIREHOSE_EXPORTER_FILTER_DEPLOYMENTS)",
	).Envar("FIREHOSE_EXPORTER_FILTER_DEPLOYMENTS").Default("").String()

	filterEvents = kingpin.Flag(
		"filter.events", "Comma separated events to filter (ContainerMetric,CounterEvent,ValueMetric,Http) ($FIREHOSE_EXPORTER_FILTER_EVENTS)",
	).Envar("FIREHOSE_EXPORTER_FILTER_EVENTS").Default("").String()

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

	enableProfiler = kingpin.Flag("profiler.enable", "Enable pprof profiling on app on /debug/pprof",
	).Envar("FIREHOSE_EXPORTER_ENABLE_PROFILER").Default("false").Bool()

	logLevel = kingpin.Flag("log.level", "Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]",
	).Envar("FIREHOSE_EXPORTER_LOG_LEVEL").Default("info").String()

	logInJson = kingpin.Flag("log.in_json", "Log in json",
	).Envar("FIREHOSE_EXPORTER_LOG_IN_JSON").Default("false").Bool()
)

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

func initLog() {
	logLvl, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Panic(err.Error())
	}
	log.SetLevel(logLvl)
	if *logInJson {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

func initMetricMaker() {
	metricmaker.SetEnableEnvelopCounterDelta(*enableRetroCompatDelta)
	metricmaker.PrependMetricConverter(metricmaker.AddNamespace(*metricsNamespace))
	metricmaker.PrependMetricConverter(metricmaker.InjectMapLabel(map[string]string{
		"environment": *metricsEnvironment,
	}))
	if !*retroCompatDisable {
		metricmaker.PrependMetricConverter(metricmaker.RetroCompatMetricNames)
	} else {
		metricmaker.PrependMetricConverter(metricmaker.SuffixCounterWithTotal)
	}

}

func MakeStreamer() (*loggregator.EnvelopeStreamConnector, error) {
	loggregatorTLSConfig, err := loggregator.NewEgressTLSConfig(*loggingTLSCa, *loggingTLSCert, *loggingTLSKey)
	if err != nil {
		return nil, err
	}

	loggregatorTLSConfig.InsecureSkipVerify = *skipSSLValidation
	return loggregator.NewEnvelopeStreamConnector(
		*loggingURL,
		loggregatorTLSConfig,
		loggregator.WithEnvelopeStreamLogger(log.StandardLogger()),
		loggregator.WithEnvelopeStreamBuffer(10000, func(missed int) {
			log.Infof("dropped %d envelope batches", missed)
		}),
	), nil
}

func main() {
	kingpin.Version(version.Print("firehose_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	initLog()
	initMetricMaker()

	log.Info("Starting firehose_exporter", version.Info())
	log.Info("Build context", version.BuildContext())

	var pointBuffer chan []*metrics.RawMetric
	if *metricsBatchSize <= 0 {
		pointBuffer = make(chan []*metrics.RawMetric)
	} else {
		pointBuffer = make(chan []*metrics.RawMetric, *metricsBatchSize)
	}

	streamer, err := MakeStreamer()
	if err != nil {
		log.Panicf("Could not create streamer: %s", err.Error())
	}

	var events []string
	if *filterEvents != "" {
		events = strings.Split(*filterEvents, ",")
	}

	var deployments []string
	if *filterDeployments != "" {
		deployments = strings.Split(*filterDeployments, ",")
	}

	im := metrics.NewInternalMetrics(*metricsNamespace, *metricsEnvironment)
	nozz := nozzle.NewNozzle(
		streamer,
		*metricsShardId,
		*metricsNodeIndex,
		pointBuffer,
		im,
		nozzle.WithNozzleTimerRollup(
			10*time.Second,
			[]string{
				"status_code", "app_name", "app_id", "space_name",
				"space_id", "organization_name", "organization_id",
				"process_id", "process_instance_id", "process_type",
				"instance_id", "method", "scheme", "host",
			},
			[]string{
				"app_name", "app_id", "space_name", "space_id",
				"organization_name", "organization_id", "process_id",
				"process_instance_id", "process_type", "instance_id",
				"method", "scheme", "host",
			},
		),
		nozzle.WithNozzleTimerRollupBufferSize(*metricsTimerRollup),
		nozzle.WithFilterSelector(nozzle.NewFilterSelector(events...)),
		nozzle.WithFilterDeployment(nozzle.NewFilterDeployment(deployments...)),
	)
	collector := collectors.NewRawMetricsCollector(pointBuffer, *metricExpiration)
	nozz.Start()
	collector.Start()

	router := http.NewServeMux()
	router.Handle(*metricsPath, prometheusHandler(collector))

	if *enableProfiler {
		router.HandleFunc("/debug/pprof/", pprof.Index)
		router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		router.HandleFunc("/debug/pprof/profile", pprof.Profile)
		router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		router.HandleFunc("/debug/pprof/trace", pprof.Trace)
		router.Handle("/debug/vars", expvar.Handler())
	}
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
		log.Fatal(http.ListenAndServeTLS(*listenAddress, *tlsCertFile, *tlsKeyFile, router))
	} else {
		log.Infoln("Listening on", *listenAddress)
		log.Fatal(http.ListenAndServe(*listenAddress, router))
	}
}

func prometheusHandler(collector *collectors.RawMetricsCollector) http.Handler {
	var handler http.Handler = http.HandlerFunc(collector.RenderExpFmt)

	if *authUsername != "" && *authPassword != "" {
		handler = &basicAuthHandler{
			handler:  handler.ServeHTTP,
			username: *authUsername,
			password: *authPassword,
		}
	}

	return handler
}
