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
	fmt.Println("API 测试程序")
	fmt.Println("============================================")

	// 测试健康检查
	testHealthCheck(baseURL)

	// 测试注册
	token := testRegister(baseURL)

	if token == "" {
		fmt.Println("❌ 注册失败，无法继续测试")
		return
	}

	// 测试登录
	token = testLogin(baseURL)
	if token == "" {
		fmt.Println("❌ 登录失败，无法继续测试")
		return
	}

	// 获取用户信息
	testGetProfile(baseURL, token)

	// 创建测试会话
	sessionID := testCreateSession(baseURL, token)

	if sessionID != "" {
		// 获取会话列表
		testListSessions(baseURL, token)

		// 创建测试消息
		messageID := testCreateMessage(baseURL, token, sessionID)

		if messageID != "" {
			// 获取消息列表
			testListMessages(baseURL, token, sessionID)
		}

		// 更新会话
		testUpdateSession(baseURL, token, sessionID)
	}

	fmt.Println("============================================")
	fmt.Println("测试完成！")
	fmt.Println("============================================")
}

func testHealthCheck(baseURL string) {
	fmt.Println("\n🔍 测试健康检查...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ 健康检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("✅ 健康检查成功: %s\n", string(body))
}

func testRegister(baseURL string) string {
	fmt.Println("\n📝 测试注册...")
	payload := map[string]interface{}{
		"username": "testuser_" + time.Now().Format("20060102150405"),
		"email":    fmt.Sprintf("test_%d@example.com", time.Now().Unix()),
		"password": "test123",
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/auth/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("❌ 注册请求失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		fmt.Printf("✅ 注册成功: %v\n", result)
		if accessToken, ok := result["access_token"].(string); ok {
			return accessToken
		}
	} else {
		fmt.Printf("❌ 注册失败: %s\n", string(body))
	}

	return ""
}

func testLogin(baseURL string) string {
	fmt.Println("\n🔑 测试登录...")
	payload := map[string]interface{}{
		"email":    "admin@link.com",
		"password": "admin123",
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
		fmt.Printf("✅ 登录成功\n")
		if accessToken, ok := result["access_token"].(string); ok {
			return accessToken
		}
	} else {
		fmt.Printf("❌ 登录失败: %s\n", string(body))
	}

	return ""
}

func testGetProfile(baseURL, token string) {
	fmt.Println("\n👤 测试获取用户信息...")
	req, _ := http.NewRequest("GET", baseURL+"/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 获取用户信息失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		fmt.Printf("✅ 获取用户信息成功: %s\n", string(body))
	} else {
		fmt.Printf("❌ 获取用户信息失败: %s\n", string(body))
	}
}

func testCreateSession(baseURL, token string) string {
	fmt.Println("\n💬 测试创建会话...")
	payload := map[string]interface{}{
		"title":   "测试会话",
		"kb_id":   nil,
		"max_rounds": 5,
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/sessions", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 创建会话失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if id, ok := result["id"].(string); ok {
			fmt.Printf("✅ 创建会话成功, ID: %s\n", id)
			return id
		}
	} else {
		fmt.Printf("❌ 创建会话失败: %s\n", string(body))
	}

	return ""
}

func testListSessions(baseURL, token string) {
	fmt.Println("\n📋 测试获取会话列表...")
	req, _ := http.NewRequest("GET", baseURL+"/sessions?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 获取会话列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		fmt.Printf("✅ 获取会话列表成功: %s\n", string(body))
	} else {
		fmt.Printf("❌ 获取会话列表失败: %s\n", string(body))
	}
}

func testCreateMessage(baseURL, token, sessionID string) string {
	fmt.Println("\n💬 测试创建消息...")
	payload := map[string]interface{}{
		"session_id": sessionID,
		"role":       "user",
		"content":    "你好，这是一条测试消息",
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", baseURL+"/messages", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 创建消息失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode == 200 {
		if id, ok := result["id"].(string); ok {
			fmt.Printf("✅ 创建消息成功, ID: %s\n", id)
			return id
		}
	} else {
		fmt.Printf("❌ 创建消息失败: %s\n", string(body))
	}

	return ""
}

func testListMessages(baseURL, token, sessionID string) {
	fmt.Println("\n📨 测试获取消息列表...")
	req, _ := http.NewRequest("GET", baseURL+"/messages?session_id="+sessionID+"&page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 获取消息列表失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		fmt.Printf("✅ 获取消息列表成功: %s\n", string(body))
	} else {
		fmt.Printf("❌ 获取消息列表失败: %s\n", string(body))
	}
}

func testUpdateSession(baseURL, token, sessionID string) {
	fmt.Println("\n✏️  测试更新会话...")
	payload := map[string]interface{}{
		"title": "测试会话（已更新）",
	}

	data, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PUT", baseURL+"/sessions/"+sessionID, bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 更新会话失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ 更新会话成功")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 更新会话失败: %s\n", string(body))
	}
}
