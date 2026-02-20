package metric

import (
	"strings"
)

// ROUGE 指标计算
// ROUGE (Recall-Oriented Understudy for Gisting Evaluation) 是评估文本摘要质量的指标

// ROUGEScore ROUGE 分数
type ROUGEScore struct {
	Precision float64
	Recall    float64
	F1        float64
}

// ROUGECalculator ROUGE 计算器
type ROUGECalculator struct {
	maxN int // 最大 n-gram 数
}

// NewROUGECalculator 创建 ROUGE 计算器
func NewROUGECalculator(maxN int) *ROUGECalculator {
	if maxN < 1 {
		maxN = 2
	}
	return &ROUGECalculator{maxN: maxN}
}

// Calculate 计算 ROUGE 分数
func (r *ROUGECalculator) Calculate(reference, candidate string, n int) *ROUGEScore {
	if reference == "" || candidate == "" {
		return &ROUGEScore{Precision: 0, Recall: 0, F1: 0}
	}

	refTokens := tokenize(reference)
	candTokens := tokenize(candidate)

	if n > r.maxN {
		n = r.maxN
	}

	refNGrams := getNGrams(refTokens, n)
	candNGrams := getNGrams(candTokens, n)

	if len(refNGrams) == 0 || len(candNGrams) == 0 {
		return &ROUGEScore{Precision: 0, Recall: 0, F1: 0}
	}

	// 计算匹配的 n-gram
	matches := 0
	for ngram, count := range candNGrams {
		refCount, exists := refNGrams[ngram]
		if exists {
			if count <= refCount {
				matches += count
			} else {
				matches += refCount
			}
		}
	}

	// 计算召回率
	recall := float64(matches) / float64(len(refNGrams))

	// 计算精确率
	precision := float64(matches) / float64(len(candNGrams))

	// 计算 F1 分数
	var f1 float64
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return &ROUGEScore{
		Precision: precision,
		Recall:    recall,
		F1:        f1,
	}
}

// ROUGE1 计算 ROUGE-1 分数
func ROUGE1(reference, candidate string) float64 {
	calc := NewROUGECalculator(1)
	score := calc.Calculate(reference, candidate, 1)
	return score.F1 // 使用 F1 分数
}

// ROUGE2 计算 ROUGE-2 分数
func ROUGE2(reference, candidate string) float64 {
	calc := NewROUGECalculator(2)
	score := calc.Calculate(reference, candidate, 2)
	return score.F1 // 使用 F1 分数
}

// ROUGEL 计算 ROUGE-L 分数
func ROUGEL(reference, candidate string) float64 {
	if reference == "" || candidate == "" {
		return 0.0
	}

	// 计算最长公共子序列
	lcs := longestCommonSubsequence(reference, candidate)

	refLen := len(strings.Fields(reference))
	candLen := len(strings.Fields(candidate))

	if refLen == 0 || candLen == 0 {
		return 0.0
	}

	// ROUGE-L 使用召回率和精确率的调和平均
	recall := float64(lcs) / float64(refLen)
	precision := float64(lcs) / float64(candLen)

	var f1 float64
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}

	return f1
}

// longestCommonSubsequence 计算最长公共子序列长度
func longestCommonSubsequence(s1, s2 string) int {
	// 简化版本：按词计算
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	m := len(words1)
	n := len(words2)

	// 创建 DP 表
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// 填充 DP 表
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if words1[i-1] == words2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}

	return dp[m][n]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
