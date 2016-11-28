package metrics

const (
	TotalEnvelopesReceivedKey               = "TotalEnvelopesReceived"
	LastEnvelopReceivedTimestampKey         = "LastEnvelopReceivedTimestamp"
	TotalMetricsReceivedKey                 = "TotalMetricsReceived"
	LastMetricReceivedTimestampKey          = "LastMetricReceivedTimestamp"
	TotalContainerMetricsReceivedKey        = "TotalContainerMetricsReceived"
	TotalContainerMetricsProcessedKey       = "TotalContainerMetricsProcessed"
	LastContainerMetricReceivedTimestampKey = "LastContainerMetricReceivedTimestamp"
	TotalCounterEventsReceivedKey           = "TotalCounterEventsReceived"
	TotalCounterEventsProcessedKey          = "TotalCounterEventsProcessed"
	LastCounterEventReceivedTimestampKey    = "LastCounterEventReceivedTimestamp"
	TotalHttpStartStopReceivedKey           = "TotalHttpStartStopReceived"
	TotalHttpStartStopProcessedKey          = "TotalHttpStartStopProcessed"
	LastHttpStartStopReceivedTimestampKey   = "LastHttpStartStopReceivedTimestamp"
	TotalValueMetricsReceivedKey            = "TotalValueMetricsReceived"
	TotalValueMetricsProcessedKey           = "TotalValueMetricsProcessed"
	LastValueMetricReceivedTimestampKey     = "LastValueMetricReceivedTimestamp"
	SlowConsumerAlertKey                    = "SlowConsumerAlert"
	LastSlowConsumerAlertTimestampKey       = "LastSlowConsumerAlertTimestamp"
)

type InternalMetrics struct {
	TotalEnvelopesReceived               int64
	LastEnvelopReceivedTimestamp         int64
	TotalMetricsReceived                 int64
	LastMetricReceivedTimestamp          int64
	TotalContainerMetricsReceived        int64
	TotalContainerMetricsProcessed       int64
	TotalContainerMetricsCached          int64
	LastContainerMetricReceivedTimestamp int64
	TotalCounterEventsReceived           int64
	TotalCounterEventsProcessed          int64
	TotalCounterEventsCached             int64
	LastCounterEventReceivedTimestamp    int64
	TotalHttpStartStopReceived           int64
	TotalHttpStartStopProcessed          int64
	TotalHttpStartStopCached             int64
	LastHttpStartStopReceivedTimestamp   int64
	TotalValueMetricsReceived            int64
	TotalValueMetricsProcessed           int64
	TotalValueMetricsCached              int64
	LastValueMetricReceivedTimestamp     int64
	SlowConsumerAlert                    bool
	LastSlowConsumerAlertTimestamp       int64
}

type ContainerMetrics []*ContainerMetric

type ContainerMetric struct {
	Origin           string
	Timestamp        int64
	Deployment       string
	Job              string
	Index            string
	IP               string
	Tags             map[string]string
	ApplicationId    string
	InstanceIndex    int32
	CpuPercentage    float64
	MemoryBytes      uint64
	DiskBytes        uint64
	MemoryBytesQuota uint64
	DiskBytesQuota   uint64
}

type CounterEvents []*CounterEvent

type CounterEvent struct {
	Origin     string
	Timestamp  int64
	Deployment string
	Job        string
	Index      string
	IP         string
	Tags       map[string]string
	Name       string
	Delta      uint64
	Total      uint64
}

type HttpStartStops []*HttpStartStop

type HttpStartStop struct {
	Origin               string
	Timestamp            int64
	Deployment           string
	Job                  string
	Index                string
	IP                   string
	Tags                 map[string]string
	RequestId            string
	Method               string
	Uri                  string
	RemoteAddress        string
	UserAgent            string
	StatusCode           int32
	ContentLength        int64
	ApplicationId        string
	InstanceIndex        int32
	InstanceId           string
	ClientStartTimestamp int64
	ClientStopTimestamp  int64
	ServerStartTimestamp int64
	ServerStopTimestamp  int64
}

type ValueMetrics []*ValueMetric

type ValueMetric struct {
	Origin     string
	Timestamp  int64
	Deployment string
	Job        string
	Index      string
	IP         string
	Tags       map[string]string
	Name       string
	Value      float64
	Unit       string
}
