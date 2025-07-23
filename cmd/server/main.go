package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mandacode-com/golib/server"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
	kafkaserver "mandacode.com/accounts/mailer/cmd/server/kafka"
	"mandacode.com/accounts/mailer/config"
	mailhandler "mandacode.com/accounts/mailer/internal/handler/mail"
	"mandacode.com/accounts/mailer/internal/usecase/mail"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	validator := validator.New()

	cfg, err := config.LoadConfig(validator)
	if err != nil {
		logger.Fatal("failed to load configuration", zap.Error(err))
	}

	mailDialer := gomail.NewDialer(
		cfg.SMTP.Host,
		cfg.SMTP.Port,
		cfg.SMTP.Username,
		cfg.SMTP.Password,
	)
	mailDialer.TLSConfig = &tls.Config{
		InsecureSkipVerify: cfg.Env != "prod", // Skip TLS verification in non-production environments
		ServerName:         cfg.SMTP.Host,     // Set ServerName for TLS verification
	}

	// Initialize MailApp
	mailUsecase, err := mail.NewMailUsecase(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.SenderName, cfg.SMTP.SenderEmail, mailDialer, logger)
	if err != nil {
		logger.Fatal("failed to create MailApp", zap.Error(err))
	}

	dialer := &kafka.Dialer{
		Timeout: 10 * time.Second, // 10 seconds
	}
	// Create Kafka mailReader (consumer)
	mailReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaMail.Brokers,
		Topic:   cfg.KafkaMail.Topic,
		GroupID: cfg.KafkaMail.GroupID,
		Dialer:  dialer,
	})
	defer mailReader.Close()

	// Create MailHandler
	mailHandler := mailhandler.NewMailHandler(mailUsecase, validator)

	// Create Kafka server with reader and handler
	kafkaServer := kafkaserver.NewKafkaServer(
		logger,
		[]*kafkaserver.ReaderHandler{
			{
				Reader:  mailReader,
				Handler: mailHandler,
			},
		})

	// Create server manager
	manager := server.NewServerManager(
		[]server.Server{kafkaServer},
	)

	// Create context for server management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled on exit

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalCh
		log.Printf("Received signal: %s, shutting down...\n", sig)
		cancel() // Cancel the context to stop the server
	}()

	// Start the server manager
	if err := manager.Run(ctx); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
