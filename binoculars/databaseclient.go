package binoculars

import (
	"fmt"
	influxcli "github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	InfluxDBPrecisionNanosecond = "ns"
)

type databaseClient interface {
	createDatabase(name string) error
	addRecords(databaseName string, metrics ...Metric) error
	close() error
}

type influxClient struct {
	httpClient influxcli.Client
}

func newInfluxClient(url, username, password string) (*influxClient, error) {
	if url == "" {
		return nil, fmt.Errorf("empty url")
	}
	cfg := influxcli.HTTPConfig{
		Addr:               url,
		InsecureSkipVerify: true,
	}
	if username != "" {
		cfg.Username = username
	}
	if password != "" {
		cfg.Password = password
	}
	c, err := influxcli.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}
	return &influxClient{httpClient: c}, nil
}

func (c *influxClient) createDatabase(name string) error {
	q := influxcli.NewQuery("CREATE DATABASE "+name, "", "")
	response, err := c.httpClient.Query(q)
	if err != nil {
		return err
	}
	if response.Error() != nil {
		return response.Error()
	}
	return nil
}

func (c *influxClient) addRecords(databaseName string, metrics ...Metric) error {
	bp, err := influxcli.NewBatchPoints(influxcli.BatchPointsConfig{
		Database:  databaseName,
		Precision: InfluxDBPrecisionNanosecond,
	})
	if err != nil {
		return err
	}
	for _, m := range metrics {
		ifpt, err := influxcli.NewPoint(m.Name, m.Tags, m.Fields, time.Now())
		if err != nil {
			logrus.WithError(err).Warnf("error creating new InfluxDB point for metric: %v. Will skip adding this point", m)
			continue
		}
		bp.AddPoint(ifpt)
	}
	if err = c.httpClient.Write(bp); err != nil {
		return err
	}
	return nil
}

func (c *influxClient) close() error {
	return c.httpClient.Close()
}
