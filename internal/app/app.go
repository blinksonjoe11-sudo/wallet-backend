package app

import (
	"fmt"
	"wallet-api/internal/config"
	"wallet-api/internal/database"
	"wallet-api/internal/model"

	"gorm.io/gorm"
)

type App struct {
	Config config.Config
	DB     *gorm.DB
}

func NewApp() (*App, error) {

	cfg, err := config.Load()

	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	// connect to database

	db, err := database.Connect(cfg)

	// automigrate user model
	db.AutoMigrate(&model.User{})
	db.AutoMigrate(&model.Wallet{})
	db.AutoMigrate(&model.Transaction{})

	if err != nil {
		return nil, fmt.Errorf("error connecting to database, %v", err)
	}

	return &App{
		Config: cfg,
		DB:     db,
	}, nil
}
