package common_repository

import (
	"time"

	"github.com/google/uuid"
)

// ========================================
// 辅助函数
// ========================================

// TimePtr 返回时间指针
func TimePtr(t time.Time) *time.Time {
	return &t
}

// NowPtr 返回当前时间指针
func NowPtr() *time.Time {
	now := time.Now()
	return &now
}

// GenerateUUID 生成 UUID
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateUUIDPrefix 生成带前缀的 UUID
func GenerateUUIDPrefix(prefix string) string {
	return prefix + "-" + uuid.New().String()
}
