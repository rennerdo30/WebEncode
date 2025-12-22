package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/rennerdo30/webencode/internal/worker"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/config"
	"github.com/rennerdo30/webencode/pkg/hardware"
	"github.com/rennerdo30/webencode/pkg/logger"
)

func main() {
	l := logger.New("worker")
	if err := run(l); err != nil {
		l.Error("Worker crashed", "error", err)
		os.Exit(1)
	}
}

func run(l *logger.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Config & Identity
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	workerID := os.Getenv("WORKER_ID")
	if workerID == "" {
		hostname, _ := os.Hostname()
		if hostname == "" || hostname == "localhost" {
			workerID = fmt.Sprintf("worker-%s", uuid.NewString()[:8])
		} else {
			workerID = hostname
		}
	}
	l.Info("Worker starting", "id", workerID)

	// 2. Capabilities
	caps := hardware.Detect()
	l.Info("Hardware capabilities", "nvidia", caps.HasNvidia, "vaapi", caps.HasVAAPI)

	// 3. Connect to Bus
	b, err := bus.Connect(cfg.NatsURL, l)
	if err != nil {
		return err
	}
	defer b.Close()

	// 4. Task Processing
	w := worker.New(workerID, b, caps, l, cfg.PluginDir)

	// 5. Heartbeat Loop
	go startHeartbeat(ctx, b, workerID, caps, w, l)

	if err := w.Start(ctx); err != nil {
		return err
	}

	// 6. Wait for stop
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	l.Info("Shutting down...")
	return nil
}

func startHeartbeat(ctx context.Context, b *bus.Bus, id string, caps *hardware.Capabilities, w *worker.Worker, l *logger.Logger) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			payload, _ := json.Marshal(map[string]interface{}{
				"id":        id,
				"timestamp": time.Now(),
				"capacity":  caps,
				"status":    w.Status(),
			})
			b.Publish(ctx, bus.SubjectWorkerHeartbeat, payload)
		}
	}
}
