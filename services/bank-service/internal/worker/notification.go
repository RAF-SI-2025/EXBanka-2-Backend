package worker

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// emailQueue mora biti isti naziv kao u notification-service consumer-u.
const emailQueue = "email_notifications"

// KreditEmailEvent je payload koji bank-service šalje na email_notifications queue.
//
// Polja Type, Email i Token su kompatibilna sa postojećim EmailEvent contractom
// koji user-service i notification-service već koriste (JSON tagovi se poklapaju).
// Polja VlasnikID, KreditID, IznosRate, Valuta, Subject i Body su proširenja
// koja notification-service mora da podrži kada doda templateove za kreditne evente.
//
// TODO: Polje Email ne može biti popunjeno u bank-service jer se email adresa
// čuva isključivo u user-service. Da bi se email resolvo·ao, potrebno je dodati
// gRPC klijent za user-service ili prihvatiti email kao deo JWT claims-a.
// Do tada se event publikuje sa praznim Email poljem i VlasnikID poljem
// koje consumer može da iskoristi za dohvat adrese.
type KreditEmailEvent struct {
	// Kompatibilna jezgra (user-service / notification-service ugovor).
	Type  string `json:"type"`  // CREDIT_RATA_USPEH | CREDIT_RATA_UPOZORENJE | CREDIT_RATA_KAZNA
	Email string `json:"email"` // email primaoca; prazno dok se ne doda user-service lookup
	Token string `json:"token"` // prazno za kreditne evente; zadržano radi kompatibilnosti

	// Proširena polja — notification-service templateovi treba da ih podrže.
	VlasnikID int64   `json:"vlasnik_id"`
	KreditID  int64   `json:"kredit_id"`
	IznosRate float64 `json:"iznos_rate"`
	Valuta    string  `json:"valuta"`
	Subject   string  `json:"subject"`
	Body      string  `json:"body"`
}

// NotificationPublisher apstrahuje slanje kreditnih notifikacija.
// Produkcijska implementacija šalje na RabbitMQ; za testove se injektuje mock.
type NotificationPublisher interface {
	Publish(event KreditEmailEvent) error
}

// =============================================================================
// AMQPKreditPublisher — produkcijska implementacija
// =============================================================================

// AMQPKreditPublisher publikuje KreditEmailEvent na RabbitMQ email_notifications queue.
// Prati isti obrazac kao user-service AMQPPublisher: dial-per-publish, fire-and-forget.
// Ako RabbitMQ nije dostupan, greška se loguje i operacija naplate se nastavlja.
type AMQPKreditPublisher struct {
	amqpURL string
}

// NewAMQPKreditPublisher kreira publisher vezan za dati RabbitMQ URL.
func NewAMQPKreditPublisher(amqpURL string) *AMQPKreditPublisher {
	return &AMQPKreditPublisher{amqpURL: amqpURL}
}

// Publish serijalizuje event u JSON i šalje ga na email_notifications queue.
// Veza i kanal se zatvaraju via defer posle svakog poziva.
func (p *AMQPKreditPublisher) Publish(event KreditEmailEvent) error {
	conn, err := amqp.Dial(p.amqpURL)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("amqp channel: %w", err)
	}
	defer ch.Close()

	// Deklariši queue kao durable — poruke preživljavaju restart brokera.
	_, err = ch.QueueDeclare(
		emailQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	return ch.Publish(
		"",         // default exchange
		emailQueue, // routing key = queue name
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // poruka preživljava restart brokera
			Body:         body,
		},
	)
}

// =============================================================================
// NoOpNotificationPublisher — za slučaj kada RabbitMQ URL nije konfigurisan
// =============================================================================

// NoOpNotificationPublisher loguje event ali ništa ne šalje.
// Koristi se kada RABBITMQ_URL nije postavljen u env-u.
type NoOpNotificationPublisher struct{}

func (p *NoOpNotificationPublisher) Publish(event KreditEmailEvent) error {
	log.Printf("[worker/notif] NoOp — event tipa %q za vlasnik_id=%d kredit_id=%d nije poslat (RabbitMQ nije konfigurisan)",
		event.Type, event.VlasnikID, event.KreditID)
	return nil
}
