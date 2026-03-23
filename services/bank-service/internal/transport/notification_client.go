// Package transport — outbound gRPC connections from bank-service.
package transport

import (
	"context"
	"fmt"
	"log"

	notifv1 "banka-backend/proto/notification"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NotificationServiceClient je gRPC klijent za notification-service.
// Koristi se za sinhronizovano slanje OTP emaila u Flow 2 (sa rollback podrškom).
type NotificationServiceClient struct {
	client notifv1.NotificationServiceClient
	conn   *grpc.ClientConn
}

// NewNotificationServiceClient otvara gRPC konekciju ka notification-service.
func NewNotificationServiceClient(addr string) (*NotificationServiceClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial notification-service at %s: %w", addr, err)
	}
	return &NotificationServiceClient{
		client: notifv1.NewNotificationServiceClient(conn),
		conn:   conn,
	}, nil
}

// SendCardOTP poziva notification-service gRPC SendEmail sa tipom "CARD_OTP".
// Notification-service renderuje template iz embedded fajla i šalje email.
func (c *NotificationServiceClient) SendCardOTP(ctx context.Context, toEmail, otpCode string) error {
	resp, err := c.client.SendEmail(ctx, &notifv1.SendEmailRequest{
		To:      toEmail,
		Subject: "CARD_OTP",
		Body:    otpCode,
	})
	if err != nil {
		return fmt.Errorf("notification-service gRPC: %w", err)
	}
	if !resp.GetSuccess() {
		return fmt.Errorf("notification-service nije uspešno poslao OTP email")
	}

	log.Printf("[notification-client] OTP email poslat na %s", toEmail)
	return nil
}

// Close oslobađa gRPC konekciju.
func (c *NotificationServiceClient) Close() error {
	return c.conn.Close()
}

// ─── NoOp implementacija ──────────────────────────────────────────────────────

// NoOpNotificationSender se koristi kada NOTIFICATION_SERVICE_ADDR nije konfigurisan.
// Svi pozivi vraćaju grešku — RequestKartica će failovati sa jasnom porukom.
type NoOpNotificationSender struct{}

func (s *NoOpNotificationSender) SendCardOTP(_ context.Context, toEmail, _ string) error {
	log.Printf("[notification-client] NoOp — OTP email nije poslat na %s (notification-service nije konfigurisan)", toEmail)
	return fmt.Errorf("notification-service nije konfigurisan")
}
