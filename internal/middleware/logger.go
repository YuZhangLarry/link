package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxBodySize = 1024 * 10 // 最大记录10KB的body内容
)

// loggerResponseBodyWriter 自定义ResponseWriter用于捕获响应内容
type loggerResponseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，同时写入buffer和原始writer
func (r loggerResponseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// sanitizeBody 清理敏感信息
func sanitizeBody(body string) string {
	result := body
	// 替换常见的敏感字段（JSON格式）
	sensitivePatterns := []struct {
		pattern     string
		replacement string
	}{
		{`"password"\s*:\s*"[^"]*"`, `"password":"***"`},
		{`"token"\s*:\s*"[^"]*"`, `"token":"***"`},
		{`"access_token"\s*:\s*"[^"]*"`, `"access_token":"***"`},
		{`"refresh_token"\s*:\s*"[^"]*"`, `"refresh_token":"***"`},
		{`"authorization"\s*:\s*"[^"]*"`, `"authorization":"***"`},
		{`"api_key"\s*:\s*"[^"]*"`, `"api_key":"***"`},
		{`"secret"\s*:\s*"[^"]*"`, `"secret":"***"`},
		{`"apikey"\s*:\s*"[^"]*"`, `"apikey":"***"`},
		{`"apisecret"\s*:\s*"[^"]*"`, `"apisecret":"***"`},
	}

	for _, p := range sensitivePatterns {
		re := regexp.MustCompile(p.pattern)
		result = re.ReplaceAllString(result, p.replacement)
	}

	return result
}

// readRequestBody 读取请求体（限制大小用于日志，但完整读取用于重置）
func readRequestBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}

	// 检查Content-Type，只记录JSON类型
	contentType := c.GetHeader("Content-Type")
	if !strings.Contains(contentType, "application/json") &&
		!strings.Contains(contentType, "application/x-www-form-urlencoded") &&
		!strings.Contains(contentType, "text/") {
		return "[非文本类型，已跳过]"
	}

	// 完整读取body内容
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "[读取请求体失败]"
	}

	// 重置request body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 用于日志的body（限制大小）
	var logBodyBytes []byte
	if len(bodyBytes) > maxBodySize {
		logBodyBytes = bodyBytes[:maxBodySize]
	} else {
		logBodyBytes = bodyBytes
	}

	bodyStr := string(logBodyBytes)
	if len(bodyBytes) > maxBodySize {
		bodyStr += "... [内容过长，已截断]"
	}

	return sanitizeBody(bodyStr)
}

// RequestID 中间件 - 添加请求ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header or generate a new one
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in header
		c.Header("X-Request-ID", requestID)

		// Set request ID in context
		SetRequestID(c, requestID)

		// Set request ID in the global context for logging
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), RequestIDKey, requestID),
		)

		c.Next()
	}
}

// Logger 中间件 - 记录请求详情
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 读取请求体
		var requestBody string
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			requestBody = readRequestBody(c)
		}

		// 创建响应体捕获器
		responseBody := &bytes.Buffer{}
		responseWriter := &loggerResponseBodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get client IP and status code
		clientIP := c.ClientIP()
		statusCode := c.Writer.Status()
		method := c.Request.Method
		requestID := GetRequestID(c)

		if raw != "" {
			path = path + "?" + raw
		}

		// 读取响应体
		responseBodyStr := ""
		if responseBody.Len() > 0 {
			contentType := c.Writer.Header().Get("Content-Type")
			if strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "text/") {
				bodyBytes := responseBody.Bytes()
				if len(bodyBytes) > maxBodySize {
					responseBodyStr = string(bodyBytes[:maxBodySize]) + "... [内容过长，已截断]"
				} else {
					responseBodyStr = string(bodyBytes)
				}
				responseBodyStr = sanitizeBody(responseBodyStr)
			} else {
				responseBodyStr = "[非文本类型，已跳过]"
			}
		}

		// 构建日志
		logAttrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status_code", statusCode),
			slog.Int("size", c.Writer.Size()),
			slog.String("latency", latency.String()),
			slog.String("client_ip", clientIP),
		}

		// 添加租户和用户信息
		if tenantID := GetTenantID(c); tenantID > 0 {
			logAttrs = append(logAttrs, slog.Int64("tenant_id", tenantID))
		}
		if userID, exists := GetUserID(c); exists {
			logAttrs = append(logAttrs, slog.Int64("user_id", userID))
		}

		// 添加请求体
		if requestBody != "" {
			logAttrs = append(logAttrs, slog.String("request_body", requestBody))
		}

		// 添加响应体
		if responseBodyStr != "" {
			logAttrs = append(logAttrs, slog.String("response_body", responseBodyStr))
		}

		// 根据状态码决定日志级别
		msg := fmt.Sprintf("%s %s", method, path)
		if statusCode >= 500 {
			slog.LogAttrs(c.Request.Context(), slog.LevelError, msg, logAttrs...)
		} else if statusCode >= 400 {
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, msg, logAttrs...)
		} else {
			slog.LogAttrs(c.Request.Context(), slog.LevelInfo, msg, logAttrs...)
		}
	}
}
