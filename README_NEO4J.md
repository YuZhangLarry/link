# Neo4j Go 操作工具

独立的 Neo4j 操作程序，采用标准的 Go 项目结构，每个命令都是独立的可执行程序。

## 📁 项目结构

```
link/
├── main.go                    # 原有的用户结构体示例
├── neo4j/                     # Neo4j 共享包
│   ├── neo4j.go              # 连接配置和驱动创建
│   └── helper.go             # 辅助函数
└── cmd/                       # 独立命令目录
    ├── insert/               # 插入数据命令
    │   └── main.go
    ├── query/                # 查询所有数据命令
    │   └── main.go
    ├── find/                 # 查找用户命令
    │   └── main.go
    ├── friends/              # 查询朋友命令
    │   └── main.go
    ├── common/               # 查询共同朋友命令
    │   └── main.go
    ├── stats/                # 统计信息命令
    │   └── main.go
    └── delete/               # 删除数据命令
        └── main.go
```

## 🚀 使用方法

### 1. 插入测试数据
```bash
go run cmd/insert/main.go
```
创建 5 个用户和 6 个朋友关系。

### 2. 查询所有数据
```bash
go run cmd/query/main.go
```
显示所有用户及其朋友列表。

### 3. 查找特定用户
```bash
go run cmd/find/main.go Alice
```
查找并显示用户的详细信息。

### 4. 查询某人的朋友
```bash
go run cmd/friends/main.go Alice
```
显示指定用户的所有朋友。

### 5. 查询共同朋友
```bash
go run cmd/common/main.go Bob Charlie
```
查询两个用户的共同朋友。

### 6. 显示统计信息
```bash
go run cmd/stats/main.go
```
显示数据库统计信息，包括：
- 总用户数
- 总关系数
- 平均年龄
- 朋友最多的用户排行
- 没有朋友的用户
- 年龄分布

### 7. 删除所有数据
```bash
go run cmd/delete/main.go
```
删除数据库中的所有节点和关系。

## 📊 示例数据

### 用户列表
- Alice (28岁)
- Bob (32岁)
- Charlie (25岁)
- David (30岁)
- Eve (27岁)

### 朋友关系
- Alice → Bob (2020年)
- Alice → Charlie (2021年)
- Bob → Charlie (2019年)
- Bob → David (2022年)
- Charlie → David (2020年)
- David → Eve (2023年)

## 🔧 连接配置

在 `neo4j/neo4j.go` 中的 `DefaultConfig()` 函数配置：
- URI: `bolt://localhost:7687`
- 用户名: `neo4j`
- 密码: `larry12345`

如需修改，编辑 `neo4j/neo4j.go` 文件。

## ✅ 完整测试流程示例

```bash
# 1. 插入数据
go run cmd/insert/main.go

# 2. 查询所有数据
go run cmd/query/main.go

# 3. 查找特定用户
go run cmd/find/main.go Alice

# 4. 查询朋友
go run cmd/friends/main.go Alice

# 5. 查询共同朋友
go run cmd/common/main.go Bob Charlie

# 6. 查看统计
go run cmd/stats/main.go

# 7. 清理数据（可选）
go run cmd/delete/main.go
```

## 🎯 优势

- ✅ 每个命令都是独立的程序
- ✅ 没有多个 main 函数冲突
- ✅ 符合 Go 项目标准结构
- ✅ 共享代码复用（neo4j 包）
- ✅ 易于维护和扩展

## 📝 扩展新命令

如需添加新命令：

1. 在 `cmd/` 下创建新目录，如 `cmd/mycmd/`
2. 创建 `main.go` 文件
3. 导入 `"link/neo4j"` 包
4. 使用 `neo4j.CreateDriver()` 创建连接
5. 实现 main 函数

示例：
```go
package main

import (
    "context"
    "log"
    "link/neo4j"
)

func main() {
    ctx := context.Background()
    driver, err := neo4j.CreateDriver(ctx, neo4j.DefaultConfig())
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    defer driver.Close(ctx)

    // 你的代码...
}
```
