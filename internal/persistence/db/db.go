package db

import (
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbConnMaxLifetime = 5 * time.Minute
	dbMaxIdleConns    = 10
	dbMaxOpenConns    = 25
)

func NewPostgres(url string) (*gorm.DB, error) {
	gormLogger := logger.New(
		&gormZerologWriter{},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		},
	)

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxLifetime(dbConnMaxLifetime)
	sqlDB.SetMaxIdleConns(dbMaxIdleConns)
	sqlDB.SetMaxOpenConns(dbMaxOpenConns)

	return db, nil
}

type gormZerologWriter struct{}

func (w *gormZerologWriter) Printf(format string, args ...any) {
	log.Warn().Msgf(format, args...)
}
