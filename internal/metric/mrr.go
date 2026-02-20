package metric

// MRR (Mean Reciprocal Rank) 平均倒数排名
// 关注第一个相关文档的排名位置

// MRRCalculator MRR 计算器
type MRRCalculator struct{}

// NewMRRCalculator 创建 MRR 计算器
func NewMRRCalculator() *MRRCalculator {
	return &MRRCalculator{}
}

// Calculate 计算 MRR
// relevantDocs: 相关文档ID列表
// retrievedDocs: 检索到的文档ID列表（按排序）
func (m *MRRCalculator) Calculate(relevantDocs, retrievedDocs []int) float64 {
	if len(relevantDocs) == 0 || len(retrievedDocs) == 0 {
		return 0.0
	}

	// 创建相关文档集合
	relevantSet := make(map[int]bool)
	for _, docID := range relevantDocs {
		relevantSet[docID] = true
	}

	// 找到第一个相关文档的排名
	for i, docID := range retrievedDocs {
		if relevantSet[docID] {
			return 1.0 / float64(i+1)
		}
	}

	return 0.0
}

// AverageMRR 计算多个查询的平均 MRR
func (m *MRRCalculator) AverageMRR(queries []QueryResult) float64 {
	if len(queries) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, query := range queries {
		sum += m.Calculate(query.RelevantDocs, query.RetrievedDocs)
	}

	return sum / float64(len(queries))
}

// QueryResult 查询结果
type QueryResult struct {
	RelevantDocs  []int // 相关文档
	RetrievedDocs []int // 检索到的文档
}
