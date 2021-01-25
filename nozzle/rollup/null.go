package rollup

type nullRollup struct {
}

func NewNullRollup() *nullRollup {
	return &nullRollup{}
}

func (h *nullRollup) Record(string, map[string]string, int64) {
}

func (h *nullRollup) Rollup(_ int64) []*PointsBatch {
	return []*PointsBatch{}
}
