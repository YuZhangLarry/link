package metric

import (
	"math"
)

// NDCG (Normalized Discounted Cumulative Gain) 归一化折损累积增益
// 用于评估排序质量

// NDCGCalculator NDCG 计算器
type NDCGCalculator struct{}

// NewNDCGCalculator 创建 NDCG 计算器
func NewNDCGCalculator() *NDCGCalculator {
	return &NDCGCalculator{}
}

// Calculate 计算 NDCG@k
// relevantDocs: 相关文档ID列表（可以包含相关性分数，这里简化为二值：相关=1，不相关=0）
// retrievedDocs: 检索到的文档ID列表（按排序）
// k: 截断位置
func (n *NDCGCalculator) Calculate(relevantDocs, retrievedDocs []int, k int) float64 {
	if len(retrievedDocs) == 0 {
		return 0.0
	}

	if k > len(retrievedDocs) {
		k = len(retrievedDocs)
	}

	// 创建相关文档集合
	relevantSet := make(map[int]bool)
	for _, docID := range relevantDocs {
		relevantSet[docID] = true
	}

	// 计算 DCG
	dcg := n.calculateDCG(relevantSet, retrievedDocs, k)

	// 计算 IDCG（理想排序）
	idealDocs := make([]int, 0, len(relevantDocs))
	for _, docID := range retrievedDocs {
		if relevantSet[docID] {
			idealDocs = append(idealDocs, docID)
		}
	}
	// 添加相关但未检索到的文档（补到k）
	for _, docID := range relevantDocs {
		found := false
		for _, id := range idealDocs {
			if id == docID {
				found = true
				break
			}
		}
		if !found && len(idealDocs) < k {
			idealDocs = append(idealDocs, docID)
		}
	}

	idcg := n.calculateDCCGivenRelevant(idealDocs, k)

	if idcg == 0 {
		return 0.0
	}

	return dcg / idcg
}

// calculateDCG 计算 DCG
func (n *NDCGCalculator) calculateDCG(relevantSet map[int]bool, retrievedDocs []int, k int) float64 {
	dcg := 0.0
	for i := 0; i < k && i < len(retrievedDocs); i++ {
		if relevantSet[retrievedDocs[i]] {
			// 相关文档的增益为 1
			if i == 0 {
				dcg += 1.0
			} else {
				dcg += 1.0 / math.Log2(float64(i+1))
			}
		}
	}
	return dcg
}

// calculateDCCGivenRelevant 给定全是相关文档的列表计算 DCG（即 IDCG）
func (n *NDCGCalculator) calculateDCCGivenRelevant(docs []int, k int) float64 {
	if k > len(docs) {
		k = len(docs)
	}

	dcg := 0.0
	for i := 0; i < k; i++ {
		if i == 0 {
			dcg += 1.0
		} else {
			dcg += 1.0 / math.Log2(float64(i+1))
		}
	}
	return dcg
}

// NDCG3 计算 NDCG@3
func (n *NDCGCalculator) NDCG3(relevantDocs, retrievedDocs []int) float64 {
	return n.Calculate(relevantDocs, retrievedDocs, 3)
}

// NDCG10 计算 NDCG@10
func (n *NDCGCalculator) NDCG10(relevantDocs, retrievedDocs []int) float64 {
	return n.Calculate(relevantDocs, retrievedDocs, 10)
}
