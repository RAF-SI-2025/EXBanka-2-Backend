package worker

import (
	"context"
	"log"
	"time"

	"banka-backend/services/bank-service/internal/domain"
)

// DailyLimitResetWorker resetuje used_limit svih agenata svake noći u 23:59.
// Prati isti pattern kao InstallmentWorker: goroutine + ticker + graceful shutdown.
type DailyLimitResetWorker struct {
	service domain.ActuaryService
}

// NewDailyLimitResetWorker konstruktor.
func NewDailyLimitResetWorker(service domain.ActuaryService) *DailyLimitResetWorker {
	return &DailyLimitResetWorker{service: service}
}

// Start pokreće worker u tekućoj goroutini (pozivati sa go worker.Start(ctx)).
// Blokira sve dok ctx ne bude otkazan.
// Worker se ne pokreće odmah pri startu — čeka do narednog 23:59.
func (w *DailyLimitResetWorker) Start(ctx context.Context) {
	log.Printf("[worker] DailyLimitResetWorker pokrenut — čeka 23:59 za reset agenata")

	for {
		next := nextResetTime()
		waitDuration := time.Until(next)
		log.Printf("[worker] DailyLimitResetWorker: sledeći reset u %s (za %s)",
			next.Format("2006-01-02 15:04:05"), waitDuration.Round(time.Second))

		select {
		case <-time.After(waitDuration):
			w.runReset(ctx)

		case <-ctx.Done():
			log.Printf("[worker] DailyLimitResetWorker prima signal za zaustavljanje — kraj")
			return
		}
	}
}

// nextResetTime vraća sledeći trenutak 23:59:00 lokalno vreme.
// Ako je sada posle 23:59, sledeći reset je sutra.
func nextResetTime() time.Time {
	now := time.Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, now.Location())
	if now.After(target) {
		target = target.Add(24 * time.Hour)
	}
	return target
}

func (w *DailyLimitResetWorker) runReset(ctx context.Context) {
	log.Printf("[worker] DailyLimitResetWorker: pokretanje reseta used_limit za sve agente")
	if err := w.service.ResetAllAgentsUsedLimit(ctx); err != nil {
		log.Printf("[worker] DailyLimitResetWorker GREŠKA pri resetovanju used_limit: %v", err)
		return
	}
	log.Printf("[worker] DailyLimitResetWorker: used_limit uspešno resetovan za sve agente")
}
