package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type BinocularsClientAgent struct {
	ServerAddress          string
	MetricHandler          MetricHandler
	DefaultRequestInterval time.Duration
	stopCh                 chan struct{}
	isRunning              bool
	lock                   *sync.RWMutex
}

type MetricHandler interface {
	GatherMetrics() Metrics
	HandleResponse(response *ServerResponse, err error)
}

func NewBinocularsClientAgent(address string, metricHandler MetricHandler) *BinocularsClientAgent {
	return &BinocularsClientAgent{
		ServerAddress:          address,
		MetricHandler:          metricHandler,
		DefaultRequestInterval: 1 * time.Hour,
		isRunning:              false,
		lock:                   &sync.RWMutex{},
	}
}

func (c *BinocularsClientAgent) Start() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.isRunning {
		return
	}
	c.stopCh = make(chan struct{})
	go c.run()
	c.isRunning = true
}

func (c *BinocularsClientAgent) Stop() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.isRunning {
		return
	}
	close(c.stopCh)
	c.isRunning = false
}

func (c *BinocularsClientAgent) SetDefaultRequestInterval(interval time.Duration) {
	c.DefaultRequestInterval = interval
}

func (c *BinocularsClientAgent) run() {
	requestInterval := c.DefaultRequestInterval

	doWork := func() {
		resp, err := c.SendMetrics(c.MetricHandler.GatherMetrics())
		if err == nil && resp.RequestIntervalInMinutes > 0 {
			requestInterval = time.Duration(resp.RequestIntervalInMinutes) * time.Minute
		}
		c.MetricHandler.HandleResponse(resp, err)
	}

	doWork()

	for {
		select {
		case <-time.After(requestInterval):
			doWork()
		case <-c.stopCh:
			return
		}
	}
}

func (c *BinocularsClientAgent) SendMetrics(metrics Metrics) (*ServerResponse, error) {
	var (
		resp    ServerResponse
		content bytes.Buffer
	)

	if err := json.NewEncoder(&content).Encode(&metrics); err != nil {
		return nil, err
	}

	r, err := http.Post(c.ServerAddress, "application/json", &content)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	if r.StatusCode != http.StatusOK {
		message := ""
		messageBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			message = err.Error()
		} else {
			message = string(messageBytes)
		}
		return nil, fmt.Errorf("query return status code %v, message %v", r.StatusCode, message)
	}
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
