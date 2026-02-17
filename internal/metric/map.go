package metric

// MAP (Mean Average Precision) 平均精确率均值
// 所有查询的平均精确率

// MAPCalculator MAP 计算器
type MAPCalculator struct {
	precisionCalc *PrecisionCalculator
}

// NewMAPCalculator 创建 MAP 计算器
func NewMAPCalculator() *MAPCalculator {
	return &MAPCalculator{
		precisionCalc: NewPrecisionCalculator(),
	}
}

// Calculate 计算单个查询的平均精确率
func (m *MAPCalculator) Calculate(relevantDocs, retrievedDocs []int) float64 {
	return m.precisionCalc.AveragePrecision(relevantDocs, retrievedDocs)
}

// AverageMAP 计算多个查询的 MAP
func (m *MAPCalculator) AverageMAP(queries []QueryResult) float64 {
	if len(queries) == 0 {
		return 0.0
	}

	sum := 0.0
	count := 0

	for _, query := range queries {
		if len(query.RelevantDocs) > 0 {
			ap := m.Calculate(query.RelevantDocs, query.RetrievedDocs)
			sum += ap
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}
