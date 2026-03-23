// Package transport manages RabbitMQ consumption for notification-service.
package transport

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"banka-backend/services/notification-service/internal/config"
	"banka-backend/services/notification-service/internal/domain"
)

const emailQueue = "email_notifications"

// StartConsumer dials RabbitMQ, declares the queue, and begins consuming
// EmailEvent messages. It is designed to be called as a goroutine from main.
// It blocks until the connection is closed.
//
// emailSvc is the domain.NotificationService interface so this layer is not
// coupled to the concrete *service.EmailService implementation.
func StartConsumer(cfg *config.Config, emailSvc domain.NotificationService) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("[rabbitmq] failed to connect: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("[rabbitmq] failed to open channel: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		emailQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("[rabbitmq] failed to declare queue: %v", err)
	}

	msgs, err := ch.Consume(
		emailQueue,
		"",    // consumer tag (auto-generated)
		false, // auto-ack — we ack manually
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("[rabbitmq] failed to register consumer: %v", err)
	}

	log.Printf("[rabbitmq] consumer started, waiting for messages on queue %q", emailQueue)

	go func() {
		for msg := range msgs {
			var event domain.EmailEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Printf("[rabbitmq] failed to unmarshal message: %v — discarding", err)
				// Malformed JSON can never be fixed by requeuing — discard it.
				msg.Ack(false)
				continue
			}

			if err := emailSvc.SendEmail(event); err != nil {
				log.Printf("[rabbitmq] failed to send email to %s (type=%s): %v — requeueing", event.Email, event.Type, err)
				// Nack with requeue=true so the message is not lost on transient
				// SMTP failures (e.g. temporary connection error). If SMTP is
				// permanently unavailable the message will cycle until it recovers.
				msg.Nack(false, true)
				continue
			}

			log.Printf("[rabbitmq] email sent to %s (type=%s)", event.Email, event.Type)
			msg.Ack(false)
		}
	}()

	// Block until the broker closes the connection.
	connErr := <-conn.NotifyClose(make(chan *amqp.Error, 1))
	if connErr != nil {
		log.Printf("[rabbitmq] connection closed: %v", connErr)
	}
}
