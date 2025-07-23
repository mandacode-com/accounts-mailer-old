package kafkaserver

import (
	"context"
	"sync"

	"github.com/mandacode-com/golib/server"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type ReaderHandler struct {
	Reader  *kafka.Reader
	Handler KafkaHandler
}

type kafkaServer struct {
	readerHandlers []*ReaderHandler
	logger         *zap.Logger
	wg             *sync.WaitGroup
}

// Start implements KafkaServer.
func (k *kafkaServer) Start(ctx context.Context) error {
	k.logger.Info("Starting Kafka server")

	for _, readerHandler := range k.readerHandlers {
		k.wg.Add(1)
		go k.runReader(ctx, readerHandler)
	}

	k.wg.Wait()
	return nil
}

// Stop implements KafkaServer.
func (k *kafkaServer) Stop(ctx context.Context) error {
	k.logger.Info("Stopping Kafka server")
	k.wg.Wait() // Wait for all readers to finish
	k.logger.Info("Kafka server stopped")
	return nil
}

func (k *kafkaServer) runReader(ctx context.Context, rh *ReaderHandler) {
	defer func() {
		k.wg.Done()
		if err := rh.Reader.Close(); err != nil {
			k.logger.Error("Failed to close reader", zap.Error(err), zap.String("topic", rh.Reader.Config().Topic))
		} else {
			k.logger.Info("Reader closed", zap.String("topic", rh.Reader.Config().Topic))
		}
	}()

	k.logger.Info("Reader started", zap.String("topic", rh.Reader.Config().Topic))

	for {
		m, err := rh.Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				k.logger.Info("Context cancelled", zap.String("topic", rh.Reader.Config().Topic))
				return
			}
			k.logger.Error("Failed to read message", zap.Error(err))
			continue
		}
		if err := rh.Handler.HandleMessage(ctx, m); err != nil {
			k.logger.Error("Failed to handle message", zap.Error(err))
		}
	}
}

func NewKafkaServer(logger *zap.Logger, readerHandlers []*ReaderHandler) server.Server {
	for _, rh := range readerHandlers {
		reader := rh.Reader
		dialer := reader.Config().Dialer

		for _, broker := range reader.Config().Brokers {
			if _, err := dialer.DialContext(context.Background(), "tcp", broker); err != nil {
				logger.Fatal("Failed to connect to Kafka broker", zap.Error(err), zap.String("broker", broker))
			}
		}
		logger.Info("Connected to Kafka broker", zap.Strings("brokers", reader.Config().Brokers))
	}
	return &kafkaServer{
		readerHandlers: readerHandlers,
		logger:         logger,
		wg:             &sync.WaitGroup{},
	}
}
