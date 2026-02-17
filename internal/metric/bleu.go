package metric

import (
	"math"
	"strings"
)

// BLEU 指标计算
// BLEU (Bilingual Evaluation Understudy) 是评估机器翻译质量的指标

// BLEUCalculator BLEU 计算器
type BLEUCalculator struct {
	maxN int // 最大 n-gram 数
}

// NewBLEUCalculator 创建 BLEU 计算器
func NewBLEUCalculator(maxN int) *BLEUCalculator {
	if maxN < 1 {
		maxN = 4
	}
	return &BLEUCalculator{maxN: maxN}
}

// Calculate 计算 BLEU 分数
// weights: 各 n-gram 的权重，如 [1.0, 0, 0, 0] 表示 BLEU-1
func (b *BLEUCalculator) Calculate(reference, candidate string, weights []float64) float64 {
	if reference == "" || candidate == "" {
		return 0.0
	}

	refTokens := tokenize(reference)
	candTokens := tokenize(candidate)

	if len(candTokens) == 0 {
		return 0.0
	}

	// 填充权重
	for len(weights) < b.maxN {
		weights = append(weights, 0.0)
	}

	// 计算各 n-gram 的精确率
	var logPSum float64
	for n := 1; n <= b.maxN; n++ {
		if n > len(weights) || weights[n-1] == 0.0 {
			continue
		}

		prec := b.precision(refTokens, candTokens, n)
		if prec > 0 {
			logPSum += weights[n-1] * math.Log(prec)
		}
	}

	// 计算简短惩罚
	bp := b.brevityPenalty(refTokens, candTokens)

	// 计算 BLEU 分数
	if logPSum == 0 {
		return 0.0
	}

	// 归一化权重
	sumWeights := 0.0
	for _, w := range weights {
		sumWeights += w
	}
	if sumWeights == 0 {
		return 0.0
	}

	return bp * math.Exp(logPSum/sumWeights)
}

// precision 计算 n-gram 精确率
func (b *BLEUCalculator) precision(refTokens, candTokens []string, n int) float64 {
	if n > len(candTokens) {
		return 0.0
	}

	// 获取候选文本的 n-gram
	candNGrams := getNGrams(candTokens, n)

	// 获取参考文本的 n-gram
	refNGrams := getNGrams(refTokens, n)

	if len(candNGrams) == 0 {
		return 0.0
	}

	// 计算匹配的 n-gram 数量
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

	return float64(matches) / float64(len(candNGrams))
}

// brevityPenalty 计算简短惩罚
func (b *BLEUCalculator) brevityPenalty(refTokens, candTokens []string) float64 {
	candLen := len(candTokens)
	refLen := len(refTokens)

	if candLen > refLen {
		return 1.0
	}

	if candLen == 0 {
		return 0.0
	}

	return math.Exp(1.0 - float64(refLen)/float64(candLen))
}

// BLEU1 计算 BLEU-1 分数
func BLEU1(reference, candidate string) float64 {
	calc := NewBLEUCalculator(1)
	return calc.Calculate(reference, candidate, []float64{1.0})
}

// BLEU2 计算 BLEU-2 分数
func BLEU2(reference, candidate string) float64 {
	calc := NewBLEUCalculator(2)
	return calc.Calculate(reference, candidate, []float64{0.5, 0.5})
}

// BLEU4 计算 BLEU-4 分数
func BLEU4(reference, candidate string) float64 {
	calc := NewBLEUCalculator(4)
	return calc.Calculate(reference, candidate, []float64{0.25, 0.25, 0.25, 0.25})
}

// getNGrams 获取 n-gram
func getNGrams(tokens []string, n int) map[string]int {
	ngrams := make(map[string]int)
	if n > len(tokens) {
		return ngrams
	}

	for i := 0; i <= len(tokens)-n; i++ {
		ngram := strings.Join(tokens[i:i+n], " ")
		ngrams[ngram]++
	}

	return ngrams
}

// tokenize 分词
func tokenize(text string) []string {
	// 简单分词：按空格分割
	// 对于中文，可以按字符分割
	tokens := strings.Fields(text)
	if len(tokens) == 0 {
		// 如果没有空格，按字符分割（适用于中文）
		return strings.Split(text, "")
	}
	return tokens
}
