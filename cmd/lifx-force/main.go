package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/alessio-palumbo/lifx-force/internal/config"
	"github.com/alessio-palumbo/lifx-force/internal/consumer"
	"github.com/alessio-palumbo/lifx-force/internal/logger"
	"github.com/alessio-palumbo/lifx-force/internal/runtime"
	"github.com/alessio-palumbo/lifxlan-go/pkg/controller"

	_ "embed"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := logger.SetupLogger(cfg)
	logger.Info("Starting lifx-force")

	exePath, err := runtime.EnsureFingertrackInstalled()
	if err != nil {
		log.Fatal(err)
	}

	ctrl, err := controller.New(controller.WithHFStateRefreshPeriod(time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer ctrl.Close()

	time.Sleep(2 * time.Second)

	cmd := exec.CommandContext(ctx, exePath)
	cmd.Cancel = func() error {
		logger.Info("Terminating fingertrack")
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal("Failed to start fingertrack:", err)
	}

	c := consumer.New(cfg, ctrl)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			var event consumer.Event
			if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
				continue
			}

			c.HandleEvent(&event)
		}
		if err := scanner.Err(); err != nil {
			logger.Error(fmt.Sprintf("Scanner error: %v", err))
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down...")

	// Graceful stop
	stop()
	done := make(chan struct{})
	go func() {
		// wait for fingertrack to exit
		cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		logger.Info("Force killing fingertrack")
		cmd.Process.Kill()
	}
}
