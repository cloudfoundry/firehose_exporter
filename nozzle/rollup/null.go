package rollup

type NullRollup struct {
}

func NewNullRollup() *NullRollup {
	return &NullRollup{}
}

func (h *NullRollup) Record(string, map[string]string, int64) {
}

func (h *NullRollup) Rollup(_ int64) []*PointsBatch {
	return []*PointsBatch{}
}
