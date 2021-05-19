package main

import (
	"confd-fork/backends"
	"confd-fork/log"
	"confd-fork/resource/template"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()

	log.SetTag("Test")
	log.SetLevel("debug")

	if err := initConfig(); err != nil {
		log.Fatal(err.Error())
	}

	log.Info("Starting ...")

	storeClient, err := backends.New(config.BackendsConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	config.TemplateConfig.StoreClient = storeClient

	stopChan := make(chan bool)
	doneChan := make(chan bool)
	errChan := make(chan error, 10)

	var processor template.Processor
	switch {
	case config.Watch:
		processor = template.WatchProcessor(config.TemplateConfig, stopChan, doneChan, errChan)
	default:
		processor = template.IntervalProcessor(config.TemplateConfig, stopChan, doneChan, errChan, config.Interval)
	}

	go processor.Process()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case err := <-errChan:
			log.Error(err.Error())
		case s := <-signalChan:
			log.Info(fmt.Sprintf("Captured %v. Exiting...", s))
			close(doneChan)
		case <-doneChan:
			log.Info("Captured done!")
			os.Exit(0)
		}
	}
}
