package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/rancher/binoculars/client"
)

const (
	// example.com should be replaced by your Binoculars server address
	binocularsServerAddress = "https://example.com/v1/metrics"
)

func main() {
	done := make(chan struct{})

	if IsAllowCollectingTelemetrics() {
		binocularsClientAgent := client.NewBinocularsClientAgent(binocularsServerAddress, NewMyMetricHandler())
		binocularsClientAgent.Start()
		defer binocularsClientAgent.Stop()
	}

	registerShutdownChannel(done)
	<-done
}

func IsAllowCollectingTelemetrics() bool {
	// get users' consent
	return true
}

func registerShutdownChannel(done chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logrus.Infof("Receive %v to exit", sig)
		close(done)
	}()
}
