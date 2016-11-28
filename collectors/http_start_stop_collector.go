package collectors

import (
	"strconv"

	"github.com/bmizerany/perks/quantile"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/cloudfoundry-community/firehose_exporter/metrics"
	"github.com/cloudfoundry-community/firehose_exporter/utils"
)

type Applications map[string]*Application

type Application struct {
	Instances map[string]*Instance
}

type Instance struct {
	Uris map[string]*Uri
}

type Uri struct {
	Methods map[string]*Method
}

type Method struct {
	StatusCodes          map[int32]int64
	ContentLength        *quantile.Stream
	LastRequestTimestamp float64
	ClientDuration       *quantile.Stream
	ServerDuration       *quantile.Stream
}

type HttpStartStopCollector struct {
	namespace                        string
	metricsStore                     *metrics.Store
	requestTotalDesc                 *prometheus.Desc
	responseSizeBytesDesc            *prometheus.Desc
	lastRequestTimestampDesc         *prometheus.Desc
	clientRequestDurationSecondsDesc *prometheus.Desc
	serverRequestDurationSecondsDesc *prometheus.Desc
}

func NewHttpStartStopCollector(
	namespace string,
	metricsStore *metrics.Store,
) *HttpStartStopCollector {
	requestTotalDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, http_start_stop_subsystem, "request_total"),
		"Cloud Foundry Firehose http start stop total requests.",
		[]string{"application_id", "instance_id", "uri", "method", "status_code"},
		nil,
	)

	responseSizeBytesDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, http_start_stop_subsystem, "response_size_bytes"),
		"Summary of Cloud Foundry Firehose http start stop request size in bytes.",
		[]string{"application_id", "instance_id", "uri", "method"},
		nil,
	)

	lastRequestTimestampDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, http_start_stop_subsystem, "last_request_timestamp"),
		"Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose.",
		[]string{"application_id", "instance_id", "uri", "method"},
		nil,
	)

	clientRequestDurationSecondsDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, http_start_stop_subsystem, "client_request_duration_seconds"),
		"Summary of Cloud Foundry Firehose http start stop client request duration in seconds.",
		[]string{"application_id", "instance_id", "uri", "method"},
		nil,
	)

	serverRequestDurationSecondsDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, http_start_stop_subsystem, "server_request_duration_seconds"),
		"Summary of Cloud Foundry Firehose http start stop server request duration in seconds.",
		[]string{"application_id", "instance_id", "uri", "method"},
		nil,
	)

	return &HttpStartStopCollector{
		namespace:                        namespace,
		metricsStore:                     metricsStore,
		requestTotalDesc:                 requestTotalDesc,
		responseSizeBytesDesc:            responseSizeBytesDesc,
		lastRequestTimestampDesc:         lastRequestTimestampDesc,
		clientRequestDurationSecondsDesc: clientRequestDurationSecondsDesc,
		serverRequestDurationSecondsDesc: serverRequestDurationSecondsDesc,
	}
}

func (c HttpStartStopCollector) Collect(ch chan<- prometheus.Metric) {
	applications := c.calculateMetrics(c.metricsStore.GetHttpStartStops())
	c.reportMetrics(applications, ch)
}

func (c HttpStartStopCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.requestTotalDesc
	ch <- c.responseSizeBytesDesc
	ch <- c.lastRequestTimestampDesc
	ch <- c.clientRequestDurationSecondsDesc
	ch <- c.serverRequestDurationSecondsDesc
}

func (c HttpStartStopCollector) calculateMetrics(httpStartStops metrics.HttpStartStops) *Applications {
	applications := Applications{}

	for _, httpStartStop := range httpStartStops {
		if httpStartStop.ApplicationId == "" {
			continue
		}

		var application *Application
		application, ok := applications[httpStartStop.ApplicationId]
		if !ok {
			application = &Application{
				Instances: make(map[string]*Instance),
			}
			applications[httpStartStop.ApplicationId] = application
		}

		var instance *Instance
		instance, ok = application.Instances[httpStartStop.InstanceId]
		if !ok {
			instance = &Instance{
				Uris: make(map[string]*Uri),
			}
			application.Instances[httpStartStop.InstanceId] = instance
		}

		var uri *Uri
		uri, ok = instance.Uris[httpStartStop.Uri]
		if !ok {
			uri = &Uri{
				Methods: make(map[string]*Method),
			}
			instance.Uris[httpStartStop.Uri] = uri
		}

		var method *Method
		method, ok = uri.Methods[httpStartStop.Method]
		if !ok {
			method = &Method{
				StatusCodes:    make(map[int32]int64),
				ContentLength:  quantile.NewTargeted(0.50, 0.90, 0.99),
				ClientDuration: quantile.NewTargeted(0.50, 0.90, 0.99),
				ServerDuration: quantile.NewTargeted(0.50, 0.90, 0.99),
			}
			uri.Methods[httpStartStop.Method] = method
		}

		method.StatusCodes[httpStartStop.StatusCode]++
		method.ContentLength.Insert(float64(httpStartStop.ContentLength))
		if httpStartStop.ClientStartTimestamp > 0 {
			method.LastRequestTimestamp = utils.NanosecondsToSeconds(httpStartStop.ClientStartTimestamp)
		} else {
			method.LastRequestTimestamp = utils.NanosecondsToSeconds(httpStartStop.ServerStartTimestamp)
		}
		clientDuration := httpStartStop.ClientStopTimestamp - httpStartStop.ClientStartTimestamp
		if clientDuration > 0 {
			method.ClientDuration.Insert(utils.NanosecondsToSeconds(clientDuration))
		}
		serverDuration := httpStartStop.ServerStopTimestamp - httpStartStop.ServerStartTimestamp
		if serverDuration > 0 {
			method.ServerDuration.Insert(utils.NanosecondsToSeconds(serverDuration))
		}
	}

	return &applications
}

func (c HttpStartStopCollector) reportMetrics(applications *Applications, ch chan<- prometheus.Metric) {
	for applicationID, application := range *applications {
		for instanceID, instance := range application.Instances {
			for uriKey, uri := range instance.Uris {
				for methodKey, method := range uri.Methods {
					c.reportResponseSize(method.ContentLength, applicationID, instanceID, uriKey, methodKey, ch)
					c.reportLastRequestTimestamp(method.LastRequestTimestamp, applicationID, instanceID, uriKey, methodKey, ch)
					c.reportClientRequestDuration(method.ClientDuration, applicationID, instanceID, uriKey, methodKey, ch)
					c.reportServerRequestDuration(method.ServerDuration, applicationID, instanceID, uriKey, methodKey, ch)

					for statusCode, requestTotal := range method.StatusCodes {
						c.reportRequestTotal(
							requestTotal,
							applicationID,
							instanceID,
							uriKey,
							methodKey,
							strconv.Itoa(int(statusCode)),
							ch,
						)
					}
				}
			}
		}
	}
}

func (c HttpStartStopCollector) reportRequestTotal(
	requestTotal int64,
	applicationID string,
	instanceID string,
	uri string,
	method string,
	statusCode string,
	ch chan<- prometheus.Metric,
) {
	ch <- prometheus.MustNewConstMetric(
		c.requestTotalDesc,
		prometheus.CounterValue,
		float64(requestTotal),
		applicationID,
		instanceID,
		uri,
		method,
		statusCode,
	)
}

func (c HttpStartStopCollector) reportResponseSize(
	responseSize *quantile.Stream,
	applicationID string,
	instanceID string,
	uri string,
	method string,
	ch chan<- prometheus.Metric,
) {
	var responseSizeSum float64
	for _, sample := range responseSize.Samples() {
		responseSizeSum = responseSizeSum + sample.Value
	}

	responseSizeQuantiles := map[float64]float64{
		float64(0.50): float64(responseSize.Query(0.50)),
		float64(0.90): float64(responseSize.Query(0.90)),
		float64(0.99): float64(responseSize.Query(0.99)),
	}

	ch <- prometheus.MustNewConstSummary(
		c.responseSizeBytesDesc,
		uint64(responseSize.Count()),
		responseSizeSum,
		responseSizeQuantiles,
		applicationID,
		instanceID,
		uri,
		method,
	)
}

func (c HttpStartStopCollector) reportLastRequestTimestamp(
	lastRequestTimestamp float64,
	applicationID string,
	instanceID string,
	uri string,
	method string,
	ch chan<- prometheus.Metric,
) {
	ch <- prometheus.MustNewConstMetric(
		c.lastRequestTimestampDesc,
		prometheus.GaugeValue,
		lastRequestTimestamp,
		applicationID,
		instanceID,
		uri,
		method,
	)
}

func (c HttpStartStopCollector) reportClientRequestDuration(
	clientRequestDuration *quantile.Stream,
	applicationID string,
	instanceID string,
	uri string,
	method string,
	ch chan<- prometheus.Metric,
) {
	var clientRequestDurationSum float64
	for _, sample := range clientRequestDuration.Samples() {
		clientRequestDurationSum = clientRequestDurationSum + sample.Value
	}

	clientRequestDurationQuantiles := map[float64]float64{
		float64(0.50): float64(clientRequestDuration.Query(0.50)),
		float64(0.90): float64(clientRequestDuration.Query(0.90)),
		float64(0.99): float64(clientRequestDuration.Query(0.99)),
	}

	ch <- prometheus.MustNewConstSummary(
		c.clientRequestDurationSecondsDesc,
		uint64(clientRequestDuration.Count()),
		clientRequestDurationSum,
		clientRequestDurationQuantiles,
		applicationID,
		instanceID,
		uri,
		method,
	)
}

func (c HttpStartStopCollector) reportServerRequestDuration(
	serverRequestDuration *quantile.Stream,
	applicationID string,
	instanceID string,
	uri string,
	method string,
	ch chan<- prometheus.Metric,
) {
	var serverRequestDurationSum float64
	for _, sample := range serverRequestDuration.Samples() {
		serverRequestDurationSum = serverRequestDurationSum + sample.Value
	}

	serverRequestDurationQuantiles := map[float64]float64{
		float64(0.50): float64(serverRequestDuration.Query(0.50)),
		float64(0.90): float64(serverRequestDuration.Query(0.90)),
		float64(0.99): float64(serverRequestDuration.Query(0.99)),
	}

	ch <- prometheus.MustNewConstSummary(
		c.serverRequestDurationSecondsDesc,
		uint64(serverRequestDuration.Count()),
		serverRequestDurationSum,
		serverRequestDurationQuantiles,
		applicationID,
		instanceID,
		uri,
		method,
	)
}
