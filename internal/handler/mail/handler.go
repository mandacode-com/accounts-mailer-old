package mailhandler

import (
	"context"

	"github.com/go-playground/validator/v10"
	mailerv1 "github.com/mandacode-com/accounts-proto/mailer/v1"
	kafka "github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
	kafkaserver "mandacode.com/accounts/mailer/cmd/server/kafka"
	"mandacode.com/accounts/mailer/internal/usecase/mail"
)

type MailHandler struct {
	MailApp   *mail.MailUsecase
	validator *validator.Validate
}

// HandleMessage implements kafkaserver.KafkaHandler.
func (h *MailHandler) HandleMessage(ctx context.Context, m kafka.Message) error {
	event := &mailerv1.EmailVerificationEvent{}
	if err := proto.Unmarshal(m.Value, event); err != nil {
		return err
	}
	if err := h.MailApp.SendEmailVerificationMail(event.Email, event.VerificationLink); err != nil {
		return err
	}
	return nil
}

func NewMailHandler(mail *mail.MailUsecase, validator *validator.Validate) kafkaserver.KafkaHandler {
	return &MailHandler{
		MailApp:   mail,
		validator: validator,
	}
}
