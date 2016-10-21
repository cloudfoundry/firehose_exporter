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
	LastContainerMetricReceivedTimestamp int64
	TotalCounterEventsReceived           int64
	TotalCounterEventsProcessed          int64
	LastCounterEventReceivedTimestamp    int64
	TotalValueMetricsReceived            int64
	TotalValueMetricsProcessed           int64
	LastValueMetricReceivedTimestamp     int64
	SlowConsumerAlert                    bool
	LastSlowConsumerAlertTimestamp       int64
}

type ContainerMetrics []ContainerMetric

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

type CounterEvents []CounterEvent

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

type ValueMetrics []ValueMetric

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
