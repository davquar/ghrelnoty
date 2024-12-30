package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	yaml "gopkg.in/yaml.v3"
	internal "it.davquar/gitrelnoty/internal/ghrelnoty"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	var configPath string
	flag.StringVar(&configPath, "config-path", "", "Path to ghrelnotify's YAML configuration file")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("GHRELNOTY_CONFIG_PATH")
	}
	if configPath == "" {
		return fmt.Errorf("config path not given: use --config-path or GHRELNOTY_CONFIG_PATH")
	}

	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	svc, err := internal.New(config)
	if err != nil {
		slog.Error("cannot initialize service", slog.Any("err", err))
		return err
	}
	defer svc.Close()

	shutdown := make(chan os.Signal, 1)
	serviceErrors := make(chan error, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func(chan<- error) {
		http.Handle("/metrics", promhttp.Handler())
		metricsServer := &http.Server{
			Addr:              fmt.Sprintf(":%d", config.MetricsPort),
			ReadHeaderTimeout: 3 * time.Second,
			Handler:           http.DefaultServeMux,
		}
		slog.Info("starting metrics server")
		err := metricsServer.ListenAndServe()
		if err != nil {
			serviceErrors <- fmt.Errorf("cannot start metrics server: %w", err)
		}
	}(serviceErrors)

	slog.Info("service ready to work")
	go svc.Work()

	select {
	case err := <-serviceErrors:
		slog.Error("service error", slog.Any("err", err))
		svc.Close()
		return err
	case <-shutdown:
		slog.Info("gracefully shutting down")
		svc.Close()
	}

	return nil
}

func loadConfig(path string) (internal.Config, error) {
	bytes, err := readYamlBytes(path)
	if err != nil {
		return internal.Config{}, fmt.Errorf("can't load config: %w", err)
	}

	config := internal.Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return internal.Config{}, fmt.Errorf("cannot unmarshal yaml: %w", err)
	}

	return config, nil
}

func readYamlBytes(path string) ([]byte, error) {
	fp, err := os.Open(path)
	defer func() {
		_ = fp.Close()
	}()

	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("config file found but can't read it: %w", err)
	} else if err == nil {
		yamlFile, err := io.ReadAll(fp)
		if err != nil {
			return nil, fmt.Errorf("can't read config file: %w", err)
		}
		return yamlFile, nil
	}

	return nil, err
}
