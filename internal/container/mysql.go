package container

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/go-sql-driver/mysql"
	"link/internal/config"
)

var (
	DB    *SQLDatabase // database/sql 封装
	GORM  *GORMDatabase // GORM 封装
)

// SQLDatabase database/sql 封装
type SQLDatabase struct {
	*sql.DB
}

// GORMDatabase GORM 封装
type GORMDatabase struct {
	*gorm.DB
}

// ========================================
// database/sql 方式
// ========================================

// InitSQLDatabase 初始化 database/sql 连接
func InitSQLDatabase(cfg *config.DatabaseConfig) (*SQLDatabase, error) {
	dsn := cfg.GetDSN()

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)

	DB = &SQLDatabase{DB: db}
	log.Printf("✅ 数据库连接成功 (sql): %s@%s:%s/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.Database)

	return DB, nil
}

// ========================================
// GORM 方式
// ========================================

// InitGORMDatabase 初始化 GORM 连接
func InitGORMDatabase(cfg *config.DatabaseConfig, logLevel string) (*GORMDatabase, error) {
	dsn := cfg.GetDSN()

	// 配置日志级别
	var gormLogLevel logger.LogLevel
	switch logLevel {
	case "silent":
		gormLogLevel = logger.Silent
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	default:
		gormLogLevel = logger.Info
	}

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("连接 GORM 数据库失败: %w", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取 sql.DB 失败: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("GORM 数据库 ping 失败: %w", err)
	}

	GORM = &GORMDatabase{DB: db}
	log.Printf("✅ 数据库连接成功 (GORM): %s@%s:%s/%s\n", cfg.User, cfg.Host, cfg.Port, cfg.Database)

	return GORM, nil
}

// ========================================
// 兼容旧接口
// ========================================

// InitDatabase 初始化数据库连接（GORM）
func InitDatabase(cfg *config.DatabaseConfig) error {
	_, err := InitGORMDatabase(cfg, "info")
	return err
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if GORM != nil {
		sqlDB, err := GORM.DB.DB()
		if err == nil {
			return sqlDB.Close()
		}
	}
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDB 获取 GORM 数据库连接
func GetDB() *gorm.DB {
	if GORM != nil {
		return GORM.DB
	}
	return nil
}

// GetSQLDB 获取 database/sql 连接
func GetSQLDB() *sql.DB {
	if DB != nil {
		return DB.DB
	}
	return nil
}
