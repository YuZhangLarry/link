package container

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"link/internal/config"
)

var DB *sql.DB

// InitDatabase 初始化数据库连接
func InitDatabase(cfg *config.DatabaseConfig) error {
	dsn := cfg.GetDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(100)              // 最大打开连接数
	db.SetMaxIdleConns(10)               // 最大空闲连接数
	db.SetConnMaxLifetime(time.Hour * 1) // 连接最大生存时间

	DB = db
	log.Printf("✅ 数据库连接成功: %s@%s:%s/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.Database)

	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB 获取数据库连接
func GetDB() *sql.DB {
	return DB
}
