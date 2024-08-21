package database

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/skinkvi/TestTaskBackDev/internal/config"
	"github.com/skinkvi/TestTaskBackDev/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDb(cfg *config.Config) (*gorm.DB, error) {
	if cfg == nil {
		logrus.Errorf("Error connecting to database: config is nil")
		return nil, fmt.Errorf("config is nil")
	}

	const op = "db.connectToDb"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password,
		cfg.Database.DB, cfg.Database.Port, cfg.Database.SSLMode)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Errorf("Error connecting to database: %v, op: %s", err, op)
		return nil, err
	}

	return DB, nil
}

func MigrateRefreshToken() error {
	if DB == nil {
		logrus.Errorf("Database connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	err := DB.AutoMigrate(&models.RefreshToken{})
	if err != nil {
		logrus.Errorf("Error migrating refresh token table: %v", err)
		return err
	}
	logrus.Info("Refresh token table migrated successfully")
	return nil
}

func MigrateUser() error {
	if DB == nil {
		logrus.Errorf("Database connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	err := DB.AutoMigrate(&models.User{})
	if err != nil {
		logrus.Errorf("Error migrating user table: %v", err)
		return err
	}
	logrus.Info("User table migrated successfully")
	return nil
}

func CloseDb() error {
	if DB == nil {
		logrus.Errorf("Database connection is nil")
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
