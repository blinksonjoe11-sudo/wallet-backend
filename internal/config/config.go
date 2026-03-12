package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	PORT        string
	DB_HOST     string
	DB_USER     string
	DB_PASSWORD string
	DB_PORT     string
	DB_NAME     string
	JWTSecret   string
}

func Load() (Config, error) {

	err := godotenv.Load()

	if err != nil {
		return Config{}, err
	}

	port, err := extractText("PORT")

	if err != nil {
		return Config{}, fmt.Errorf("Port cant be empty")
	}

	db_host, err := extractText("DB_HOST")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find host")
	}
	db_port, err := extractText("DB_PORT")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find port")
	}
	db_password, err := extractText("DB_PASSWORD")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find password")
	}

	db_name, err := extractText("DB_NAME")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find databae name")
	}

	db_user, err := extractText("DB_user")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find host")
	}

	jwtSecret, err := extractText("JWTSecret")

	if err != nil {
		return Config{}, fmt.Errorf("cannot find host")
	}

	return Config{
		PORT:        port,
		DB_HOST:     db_host,
		DB_USER:     db_user,
		DB_PASSWORD: db_password,
		DB_PORT:     db_port,
		DB_NAME:     db_name,
		JWTSecret:   jwtSecret,
	}, nil
}

func extractText(key string) (string, error) {

	cf := os.Getenv(key)

	if cf == "" {
		return "", fmt.Errorf("Cant be empty")
	}

	cfg := strings.TrimSpace(cf)

	return cfg, nil
}
