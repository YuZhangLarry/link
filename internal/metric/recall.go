package metric

// Recall 召回率计算
// Recall = 检索到的相关文档数 / 相关文档总数

// RecallCalculator 召回率计算器
type RecallCalculator struct{}

// NewRecallCalculator 创建召回率计算器
func NewRecallCalculator() *RecallCalculator {
	return &RecallCalculator{}
}

// Calculate 计算召回率
// relevantDocs: 相关文档ID列表
// retrievedDocs: 检索到的文档ID列表
func (r *RecallCalculator) Calculate(relevantDocs, retrievedDocs []int) float64 {
	if len(relevantDocs) == 0 {
		return 0.0
	}

	// 创建相关文档集合
	relevantSet := make(map[int]bool)
	for _, docID := range relevantDocs {
		relevantSet[docID] = true
	}

	// 计算检索到的相关文档数
	matches := 0
	for _, docID := range retrievedDocs {
		if relevantSet[docID] {
			matches++
		}
	}

	return float64(matches) / float64(len(relevantDocs))
}

// RecallAtK 计算@K的召回率
func (r *RecallCalculator) RecallAtK(relevantDocs, retrievedDocs []int, k int) float64 {
	if k > len(retrievedDocs) {
		k = len(retrievedDocs)
	}
	if k == 0 {
		return 0.0
	}

	return r.Calculate(relevantDocs, retrievedDocs[:k])
}
