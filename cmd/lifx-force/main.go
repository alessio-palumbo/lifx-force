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
	"path/filepath"
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to load homedir:", err)
	}
	cfg, err := config.LoadConfig(filepath.Join(homeDir, ".lifx-force", "config.toml"))
	if err != nil {
		log.Fatal("Failed to load config file:", err)
	}

	logger := logger.SetupLogger(cfg)
	logger.Info("Starting lifx-force")

	exePath, err := runtime.EnsureFingertrackInstalled(logger)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Initializing LIFX LAN Controller")

	ctrl, err := controller.New(controller.WithHFStateRefreshPeriod(time.Second))
	if err != nil {
		log.Fatal("Failed to initialize lifxlan-go controller:", err)
	}
	defer ctrl.Close()

	// Allow discovery to occur so that Labels and Groups are available to consumer.
	// TODO Add some logic to periodically refresh the list of devices inside the consumer.
	time.Sleep(2 * time.Second)

	cmd := exec.CommandContext(ctx, exePath, runtime.ArgsFromConfig(cfg)...)
	cmd.Cancel = func() error {
		logger.Info("Process cancelled: terminating fingertrack")
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("Failed to create stdout pipe:", err)
	}
	cmd.Stderr = os.Stderr

	logger.Info("Starting Fingertrack program")

	if err := cmd.Start(); err != nil {
		log.Fatal("Failed to start fingertrack:", err)
	}

	logger.Info("Starting consumer")
	c := consumer.New(cfg, ctrl, logger)

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
