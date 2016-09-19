package metrics

type InternalMetrics struct {
	TotalEnvelopesReceived        float64
	TotalMetricsReceived          float64
	TotalContainerMetricsReceived float64
	TotalCounterEventsReceived    float64
	TotalValueMetricsReceived     float64
	SlowConsumerAlert             bool
}

type ContainerMetrics map[string]ContainerMetric

type ContainerMetric struct {
	Origin           string
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

type CounterMetrics map[string]CounterMetric

type CounterMetric struct {
	Origin     string
	Deployment string
	Job        string
	Index      string
	IP         string
	Tags       map[string]string
	Name       string
	Delta      uint64
	Total      uint64
}

type ValueMetrics map[string]ValueMetric

type ValueMetric struct {
	Origin     string
	Deployment string
	Job        string
	Index      string
	IP         string
	Tags       map[string]string
	Name       string
	Value      float64
	Unit       string
}
