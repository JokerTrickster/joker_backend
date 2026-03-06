package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_aws "github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/JokerTrickster/joker_backend/shared/logger"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MysqlDB *sql.DB
var GormMysqlDB *gorm.DB

const DBTimeOut = 15 * time.Second

func InitMySQL() error {
	var connectionString string
	var err error
	isLocal := os.Getenv("IS_LOCAL")

	logger.Info("MySQL init",
		zap.String("IS_LOCAL", isLocal),
		zap.String("MYSQL_HOST", os.Getenv("MYSQL_HOST")),
		zap.String("MYSQL_PORT", os.Getenv("MYSQL_PORT")),
		zap.String("MYSQL_USER", os.Getenv("MYSQL_USER")),
		zap.String("MYSQL_DATABASE", os.Getenv("MYSQL_DATABASE")),
	)

	if isLocal == "true" {
		connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASSWORD"),
			os.Getenv("MYSQL_HOST"),
			os.Getenv("MYSQL_PORT"),
			os.Getenv("MYSQL_DATABASE"),
		)
	} else {
		logger.Info("Using AWS SSM parameters for database connection")
		dbInfos, err := _aws.AwsSsmGetParams(context.Background(), []string{"dev_backend_mysql_user", "dev_backend_mysql_password", "dev_backend_mysql_host", "dev_backend_mysql_port", "dev_backend_mysql_db"})
		if err != nil {
			return err
		}
		connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			dbInfos[4], //user
			dbInfos[1], //password
			dbInfos[0], //host
			dbInfos[2], //port
			dbInfos[3], //db
		)
	}

	MysqlDB, err = sql.Open("mysql", connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}
	MysqlDB.SetMaxOpenConns(25)
	MysqlDB.SetMaxIdleConns(10)
	MysqlDB.SetConnMaxLifetime(5 * time.Minute)
	logger.Info("Connected to MySQL")

	GormMysqlDB, err = gorm.Open(mysql.New(mysql.Config{
		Conn: MysqlDB,
	}), &gorm.Config{
		SkipDefaultTransaction: false,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize Gorm MySQL: %w", err)
	}

	logger.Info("Gorm MySQL connected successfully")
	return nil
}

func PKIDGenerate() string {
	//uuid 로 생성
	result := (uuid.New()).String()
	return result
}

func NowDateGenerate() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func EpochToTime(t int64) time.Time {
	return time.Unix(t, t%1000*1000000)
}
func EpochToTimeString(t int64) string {
	return time.Unix(t, t%1000*1000000).String()
}

func TimeStringToEpoch(t string) (int64, error) {
	date, err := time.Parse("2006-01-02 15:04:05 -0700 MST", t)
	if err != nil {
		return 0, err
	}
	return date.Unix(), nil
}

func TimeToEpoch(t time.Time) int64 {
	return t.Unix()
}

// 트랜잭션 처리 미들웨어
func Transaction(db *gorm.DB, fc func(tx *gorm.DB) error) (err error) {
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			err = fmt.Errorf("panic occurred: %v", r)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()

	if err = tx.Error; err != nil {
		return err
	}

	err = fc(tx)
	return
}
