package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type MailConfig struct {
	Host       string `validate:"required"`
	Port       int    `validate:"required,min=1"`
	Username   string `validate:"required"`
	Password   string `validate:"required"`
	SenderEmail     string `validate:"required,email"`
	SenderName string `validate:"required"`
}

type KafkaConfig struct {
	Brokers []string `validate:"required,dive,required"`
	Topic   string   `validate:"required"`
	GroupID string   `validate:"required"`
}

type Config struct {
	Env       string      `validate:"required,oneof=dev prod"`
	SMTP      MailConfig  `validate:"required"`
	KafkaMail KafkaConfig `validate:"required"`
}

// LoadConfig loads env vars from .env (if exists) and returns structured config
func LoadConfig(v *validator.Validate) (*Config, error) {
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}

	mailPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	if err != nil {
		return nil, err
	}

	mailConfig := MailConfig{
		Host:     getEnv("SMTP_HOST", ""),
		Port:     mailPort,
		Username: getEnv("SMTP_USER", ""),
		Password: getEnv("SMTP_PASS", ""),
		SenderEmail: getEnv("SMTP_SENDER_EMAIL", ""),
		SenderName: getEnv("SMTP_SENDER_NAME", ""),
	}

	kafkaConfig := KafkaConfig{
		Brokers: strings.Split(getEnv("KAFKA_MAIL_BROKERS", ""), ","),
		Topic:   getEnv("KAFKA_MAIL_TOPIC", ""),
		GroupID: getEnv("KAFKA_MAIL_GROUP_ID", ""),
	}

	config := &Config{
		Env:       getEnv("ENV", "dev"),
		SMTP:      mailConfig,
		KafkaMail: kafkaConfig,
	}

	if err := v.Struct(config); err != nil {
		return nil, errors.New("invalid configuration: " + err.Error())
	}

	return config, nil
}

// getEnv returns env value or fallback
func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
