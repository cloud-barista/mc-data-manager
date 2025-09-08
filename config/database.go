package config

import (
	// "database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig 데이터베이스 설정
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

var DB *gorm.DB

// NewDatabaseConfig 데이터베이스 설정 생성
func NewDatabaseConfig() *DatabaseConfig {
	host := os.Getenv("MC_DATA_MANAGER_DATABASE_HOST")
	port := os.Getenv("MC_DATA_MANAGER_DATABASE_PORT")
	user := os.Getenv("MC_DATA_MANAGER_DATABASE_USER")
	password := os.Getenv("MC_DATA_MANAGER_DATABASE_PASSWORD")
	dbname := os.Getenv("MC_DATA_MANAGER_DATABASE_NAME")

	// 디버깅을 위한 로그 추가
	log.Printf("Database config - Host: %s, Port: %s, User: %s, DBName: %s",
		host, port, user, dbname)

	return &DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}
}

// GetDSN 데이터베이스 연결 문자열 반환
func (c *DatabaseConfig) GetDSN() string {
	if dsn := os.Getenv("MC_DATA_MANAGER_DATABASE_URL"); dsn != "" {
		log.Printf("Using MC_DATA_MANAGER_DATABASE_URL: %s", dsn)
		return dsn
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.DBName)
	log.Printf("Generated DSN: %s", dsn)
	return dsn
}

// func InitDB() (*sql.DB, error) {
// 	host := os.Getenv("MC_DATA_MANAGER_DATABASE_HOST")
// 	port := os.Getenv("MC_DATA_MANAGER_DATABASE_PORT")
// 	user := os.Getenv("MC_DATA_MANAGER_DATABASE_USER")
// 	password := os.Getenv("MC_DATA_MANAGER_DATABASE_PASSWORD")
// 	dbname := os.Getenv("MC_DATA_MANAGER_DATABASE_NAME")

// 	if dsn := os.Getenv("MC_DATA_MANAGER_DATABASE_URL"); dsn != "" {
// 		db, err := sql.Open("mysql", dsn)
// 		if err != nil {
// 			return nil, fmt.Errorf("데이터베이스 연결 실패: %v", err)
// 		}
// 		return db, nil
// 	}

// 	msqlInfo := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
// 		host, port, user, password, dbname)

// 	db, err := sql.Open("mysql", msqlInfo)
// 	if err != nil {
// 		return nil, fmt.Errorf("데이터베이스 연결 실패: %v", err)
// 	}

// 	if err = db.Ping(); err != nil {
// 		return nil, fmt.Errorf("데이터베이스 ping 실패: %v", err)
// 	}

// 	return db, nil
// }

func InitDB() {
	cfg := NewDatabaseConfig()
	dsn := cfg.GetDSN()

	// GORM 오픈 (에러만 로깅)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatalf("DB 연결 실패: %v", err)
	}

	// 커넥션 풀 설정
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("DB 핸들러 획득 실패: %v", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(1 * time.Hour)

	DB = db
	log.Println("DB 연결 성공")
}
