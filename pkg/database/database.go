package database

import (
	"errors"
	"sync"
	"time"

	"github.com/krispeckt/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	logger = logx.New()
	db     *gorm.DB
	mu     sync.Mutex
)

func InitDatabase(DSN string) error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		logger.Infof("database connection already initialized")
		return nil
	}

	newDB, err := gorm.Open(postgres.Open(DSN), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return err
	}

	sqlDB, err := newDB.DB()
	if err != nil {
		return errors.New("could not connect to database: " + err.Error())
	}

	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	db = newDB
	logger.Infof("created a connection to db")
	return nil
}

// GetDB возвращает экземпляр GORM
func GetDB() *gorm.DB {
	return db
}

// CloseDatabase закрывает соединение с БД
func CloseDatabase() {
	mu.Lock()
	defer mu.Unlock()

	if db == nil {
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Warnf("error when retrieving db: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		logger.Warnf("error while closing connection: %v", err)
		return
	}

	db = nil
	logger.Infof("closed the connection to db")
}
