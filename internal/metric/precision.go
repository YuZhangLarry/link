package metric

// Precision 精确率计算
// Precision = 检索到的相关文档数 / 检索到的总文档数

// PrecisionCalculator 精确率计算器
type PrecisionCalculator struct{}

// NewPrecisionCalculator 创建精确率计算器
func NewPrecisionCalculator() *PrecisionCalculator {
	return &PrecisionCalculator{}
}

// Calculate 计算精确率
// relevantDocs: 相关文档ID列表
// retrievedDocs: 检索到的文档ID列表（按排序）
func (p *PrecisionCalculator) Calculate(relevantDocs, retrievedDocs []int) float64 {
	if len(retrievedDocs) == 0 {
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

	return float64(matches) / float64(len(retrievedDocs))
}

// PrecisionAtK 计算@K的精确率
func (p *PrecisionCalculator) PrecisionAtK(relevantDocs, retrievedDocs []int, k int) float64 {
	if k > len(retrievedDocs) {
		k = len(retrievedDocs)
	}
	if k == 0 {
		return 0.0
	}

	return p.Calculate(relevantDocs, retrievedDocs[:k])
}

// AveragePrecision 计算平均精确率
func (p *PrecisionCalculator) AveragePrecision(relevantDocs, retrievedDocs []int) float64 {
	if len(relevantDocs) == 0 {
		return 0.0
	}

	relevantSet := make(map[int]bool)
	for _, docID := range relevantDocs {
		relevantSet[docID] = true
	}

	var sumPrecision float64
	relevantCount := 0

	for i, docID := range retrievedDocs {
		if relevantSet[docID] {
			relevantCount++
			precision := float64(relevantCount) / float64(i+1)
			sumPrecision += precision
		}
	}

	if relevantCount == 0 {
		return 0.0
	}

	return sumPrecision / float64(relevantCount)
}
