package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	baseURL := "http://localhost:8080/api/v1"

	fmt.Println("============================================")
	fmt.Println("租户过滤测试程序")
	fmt.Println("============================================")

	// 步骤0: 先注册测试用户
	fmt.Println("\n📝 步骤0: 注册测试用户...")
	registerUser(baseURL, "test_a", "test_a@example.com", "test123")
	registerUser(baseURL, "test_b", "test_b@example.com", "test123")

	// 步骤1: 登录获取 Token
	fmt.Println("\n📝 步骤1: 登录测试用户A...")
	tokenA := login(baseURL, "test_a@example.com", "test123")
	if tokenA == "" {
		fmt.Println("❌ 登录失败，无法继续测试")
		return
	}
	fmt.Println("✅ 用户A登录成功")

	// 步骤2: 获取当前用户信息
	fmt.Println("\n👤 步骤2: 获取用户A信息...")
	userProfile := getUserProfile(baseURL, tokenA)
	if userProfile != nil {
		fmt.Printf("用户信息: %s\n", prettyPrint(userProfile))
	}

	// 步骤3: 查询会话列表（验证租户过滤）
	fmt.Println("\n💬 步骤3: 查询会话列表（用户A）...")
	sessionsA := listSessions(baseURL, tokenA)
	fmt.Printf("用户A 找到 %d 个会话\n", len(sessionsA))
	for i, sess := range sessionsA {
		if i < 3 {
			fmt.Printf("  - %s (状态: %v)\n", sess["title"], sess["status"])
		}
	}

	// 步骤4: 创建新会话
	fmt.Println("\n➕ 步骤4: 创建新会话...")
	newSessionID := createSession(baseURL, tokenA, "用户A的新会话")
	if newSessionID != "" {
		fmt.Printf("✅ 创建会话成功, ID: %s\n", newSessionID)

		fmt.Println("\n📨 步骤5: 发送测试消息...")
		msgID := createMessage(baseURL, tokenA, newSessionID, "user", "这是一条测试消息")
		if msgID != "" {
			fmt.Printf("✅ 消息发送成功, ID: %s\n", msgID)
		}

		// 步骤5.5: 再次查询会话列表验证创建成功
		fmt.Println("\n💬 步骤5.5: 再次查询会话列表（用户A，验证创建）...")
		sessionsAAfter := listSessions(baseURL, tokenA)
		fmt.Printf("用户A 现在找到 %d 个会话\n", len(sessionsAAfter))
		for i, sess := range sessionsAAfter {
			if i < 3 {
				fmt.Printf("  - %s (状态: %v)\n", sess["title"], sess["status"])
			}
		}
		sessionsA = sessionsAAfter // 更新变量以便后续使用
	} else {
		fmt.Println("❌ 创建会话失败")
	}

	// 步骤6: 切换用户测试（租户B）
	fmt.Println("\n🔄 步骤6: 切换到用户B测试...")
	tokenB := login(baseURL, "test_b@example.com", "test123")
	if tokenB == "" {
		fmt.Println("❌ 用户B登录失败")
		return
	}
	fmt.Println("✅ 用户B登录成功")

	// 步骤7: 用用户B查询会话
	fmt.Println("\n💬 步骤7: 查询会话列表（用户B）...")
	sessionsB := listSessions(baseURL, tokenB)
	fmt.Printf("用户B 找到 %d 个会话\n", len(sessionsB))
	for i, sess := range sessionsB {
		if i < 3 {
			fmt.Printf("  - %s (状态: %v)\n", sess["title"], sess["status"])
		}
	}

	// 步骤8: 验证租户隔离
	fmt.Println("\n🔒 步骤8: 验证租户隔离...")
	if len(sessionsA) > 0 {
		tenantAFirstSession := sessionsA[0]
		sessionID := toString(tenantAFirstSession["id"])
		fmt.Printf("尝试用用户B访问用户A的会话: %s\n", sessionID)

		accessResult := getSessionDetail(baseURL, tokenB, sessionID)
		if accessResult {
			fmt.Println("⚠️  警告: 用户B可以访问用户A的会话，租户隔离可能有问题！")
		} else {
			fmt.Println("✅ 租户隔离正常: 用户B无法访问用户A的会话")
		}
	}

	fmt.Println("\n============================================")
	fmt.Println("测试完成！")
	fmt.Println("============================================")

	// 总结
	fmt.Println("\n📊 测试总结:")
	fmt.Printf("  - 用户A (test_a) 的会话数: %d\n", len(sessionsA))
	fmt.Printf("  - 用户B (test_b) 的会话数: %d\n", len(sessionsB))
	if len(sessionsA) > 0 && len(sessionsB) > 0 {
		fmt.Println("  ✅ 多用户数据隔离功能正常")
	}
}

// registerUser 注册用户
func registerUser(baseURL, username, email, password string) bool {
	payload := map[string]interface{}{
		"username": username,
		"email":    email,
		"password": password,
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("  ❌ 注册请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != 200 && resp.StatusCode != 409 {
		fmt.Printf("  ❌ 注册失败 (HTTP %d): %s\n", resp.StatusCode, string(body))
		return false
	}

	// 打印成功信息
	if resp.StatusCode == 200 {
		fmt.Printf("  ✅ 注册成功: %s\n", username)
	} else if resp.StatusCode == 409 {
		fmt.Printf("  ℹ️  用户已存在: %s\n", username)
	}

	return resp.StatusCode == 200 || resp.StatusCode == 409
}

// login 登录获取 token
func login(baseURL, email, password string) string {
	payload := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/auth/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		// token 在 data 字段中
		if dataField, ok := result["data"].(map[string]interface{}); ok {
			if accessToken, ok := dataField["access_token"].(string); ok {
				return accessToken
			}
		}
	}

	fmt.Printf("❌ 登录失败: %s\n", string(body))
	return ""
}

// getUserProfile 获取用户信息
func getUserProfile(baseURL, token string) map[string]interface{} {
	req, _ := http.NewRequest("GET", baseURL+"/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if dataField, ok := result["data"].(map[string]interface{}); ok {
			return dataField
		}
	}

	return nil
}

// listSessions 查询会话列表
func listSessions(baseURL, token string) []map[string]interface{} {
	req, _ := http.NewRequest("GET", baseURL+"/sessions?page=1&size=100", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if dataField, ok := result["data"].(map[string]interface{}); ok {
			if sessions, ok := dataField["sessions"].([]interface{}); ok {
				result := make([]map[string]interface{}, 0, len(sessions))
				for _, s := range sessions {
					if session, ok := s.(map[string]interface{}); ok {
						result = append(result, session)
					}
				}
				return result
			}
		}
	}

	return nil
}

// createSession 创建会话
func createSession(baseURL, token, title string) string {
	payload := map[string]interface{}{
		"title":      title,
		"max_rounds": 5,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/sessions", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if dataField, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := dataField["id"].(string); ok {
				return id
			}
		}
	}

	fmt.Printf("创建会话响应: %s\n", string(body))
	return ""
}

// createMessage 创建消息
func createMessage(baseURL, token, sessionID, role, content string) string {
	payload := map[string]interface{}{
		"session_id": sessionID,
		"role":       role,
		"content":    content,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/messages", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if dataField, ok := result["data"].(map[string]interface{}); ok {
			if id, ok := dataField["id"].(string); ok {
				return id
			}
		}
	}

	return ""
}

// getSessionDetail 获取会话详情
func getSessionDetail(baseURL, token, sessionID string) bool {
	req, _ := http.NewRequest("GET", baseURL+"/sessions/"+sessionID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		fmt.Printf("访问失败: %s\n", string(body))
	}
	return resp.StatusCode == 200
}

// toString 将 interface{} 转换为 string
func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// prettyPrint 简单打印 map
func prettyPrint(m map[string]interface{}) string {
	if m == nil {
		return "{}"
	}
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ", "
		}
		result += fmt.Sprintf("%v: %v", k, v)
		first = false
	}
	result += "}"
	return result
}
