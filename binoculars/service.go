package binoculars

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	DatabaseSuffix = "_binoculars"
)

type Server struct {
	done           chan struct{}
	databaseClient databaseClient
	databaseName   string
	queryPeriod    string
}

func NewServer(done chan struct{}, applicationName, dbURL, dbUser, dbPass, queryPeriod string) (*Server, error) {
	dbClient, err := newInfluxClient(dbURL, dbUser, dbPass)
	if err != nil {
		return nil, errors.Wrap(err, "fail to create db client")
	}
	logrus.Debugf("Database connection established")

	s := &Server{
		done:           done,
		databaseClient: dbClient,
		databaseName:   applicationName + DatabaseSuffix,
		queryPeriod:    queryPeriod,
	}

	go func() {
		<-done
		if err := s.databaseClient.close(); err != nil {
			logrus.Debugf("Failed to close DB connection: %v", err)
		} else {
			logrus.Debug("DB connection closed")
		}
	}()

	if err := s.initDB(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) initDB() error {
	return s.databaseClient.createDatabase(s.databaseName)
}

func (s *Server) HealthCheck(rw http.ResponseWriter, req *http.Request) {
	rw.WriteHeader(http.StatusOK)
}

func (s *Server) RecordMetrics(rw http.ResponseWriter, req *http.Request) {
	var (
		err     error
		metrics Metrics
	)

	defer func() {
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
		}
	}()

	if err = json.NewDecoder(req.Body).Decode(&metrics); err != nil {
		return
	}

	logrus.Debugf("Metrics from request's body: %+v", metrics)

	if err = s.databaseClient.addRecords(s.databaseName, metrics...); err != nil {
		return
	}

	resp, err := s.GenerateResponse()
	if err != nil {
		logrus.Errorf("Failed to GenerateResponse: %v", err)
		return
	}

	if err = respondWithJSON(rw, resp); err != nil {
		logrus.Errorf("Failed to repsondWithJSON: %v", err)
		return
	}
}

func (s *Server) GenerateResponse() (*ServerResponse, error) {
	resp := &ServerResponse{}
	d, err := time.ParseDuration(s.queryPeriod)
	if err != nil {
		logrus.Errorf("fail to parse queryPeriod while building the response: %v", err)
		resp.RequestIntervalInMinutes = 60
	} else {
		resp.RequestIntervalInMinutes = int(d.Minutes())
	}
	return resp, nil
}

func respondWithJSON(rw http.ResponseWriter, obj interface{}) error {
	response, err := json.Marshal(obj)
	if err != nil {
		return errors.Wrapf(err, "fail to marshal %v", obj)
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, err = rw.Write(response)
	return err
}
