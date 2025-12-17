// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/btouchard/shm/sdk/golang"
)

func main() {
	cfg := golang.Config{
		ServerURL:      "http://localhost:8080",
		AppName:        "Ackify",
		AppVersion:     "2.0.0-simulation",
		Environment:    "production",
		Enabled:        true,
		DataDir:        ".",
		ReportInterval: 10 * time.Second,
	}

	client, err := golang.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	docs := int64(10)
	sigs := int64(5)
	reminds := int64(4)
	whs := int64(2)

	client.SetProvider(func() map[string]interface{} {
		docs += int64(rand.Intn(5))
		sigs += int64(rand.Intn(3))
		reminds += int64(rand.Intn(2))
		whs += int64(rand.Intn(1))

		log.Printf("📊 Collecting metrics: Docs=%d, Sigs=%d, Reminder=%d, Webhooks=%d", docs, sigs, reminds, whs)

		return map[string]interface{}{
			"documents_total":   docs,
			"signatures_total":  sigs,
			"remind_sent_total": reminds,
			"webhooks_total":    whs,
		}
	})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		log.Println("🛑 Stopping simulator...")
		cancel()
	}()

	log.Println("🚀 Starting Simulator (Register -> Activate -> Loop)...")
	client.Start(ctx)
	log.Println("👋 Simulator stopped.")
}
