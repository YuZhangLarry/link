package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// responseBodyWriter 用于捕获响应内容
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，同时写入buffer和原始writer
func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// TracingMiddleware 追踪中间件
// 注意：当前版本为简化版，不包含 OpenTelemetry 集成
// 如果需要分布式追踪，需要先安装 OpenTelemetry SDK 并创建 internal/tracing 包
func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求信息到 context
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())

		// 可以在这里添加 OpenTelemetry span
		// ctx, span := tracing.ContextWithSpan(c.Request.Context(), spanName)
		// defer span.End()

		// 记录请求信息
		c.Set("span_name", spanName)

		// 记录请求头（跳过敏感头）
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if strings.ToLower(key) != "authorization" && strings.ToLower(key) != "cookie" {
				headers[key] = strings.Join(values, ";")
			}
		}
		c.Set("request_headers", headers)

		// 记录请求体
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if c.Request.Body != nil {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				c.Set("request_body", string(bodyBytes))
				// 重置请求体
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 记录查询参数
		if len(c.Request.URL.RawQuery) > 0 {
			c.Set("query_params", c.Request.URL.RawQuery)
		}

		// 创建响应体捕获器
		responseBody := &bytes.Buffer{}
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           responseBody,
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 记录响应状态
		statusCode := c.Writer.Status()
		c.Set("status_code", statusCode)

		// 记录响应体
		if responseBody.Len() > 0 {
			c.Set("response_body", responseBody.String())
		}

		// 可以在这里添加 span 属性和状态
		// span.SetAttributes(...)
		// if statusCode >= 400 {
		//     span.SetStatus(codes.Error, ...)
		// }
	}
}

// GetTraceInfo 获取追踪信息（用于调试）
func GetTraceInfo(c *gin.Context) map[string]interface{} {
	info := make(map[string]interface{})

	if v, exists := c.Get("span_name"); exists {
		info["span_name"] = v
	}
	if v, exists := c.Get("request_headers"); exists {
		info["request_headers"] = v
	}
	if v, exists := c.Get("request_body"); exists {
		info["request_body"] = v
	}
	if v, exists := c.Get("query_params"); exists {
		info["query_params"] = v
	}
	if v, exists := c.Get("status_code"); exists {
		info["status_code"] = v
	}
	if v, exists := c.Get("response_body"); exists {
		info["response_body"] = v
	}

	return info
}
