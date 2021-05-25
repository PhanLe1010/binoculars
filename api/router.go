package api

import (
	"github.com/gorilla/mux"
	"github.com/rancher/binoculars/binoculars"
)

func NewRouter(s *binoculars.Server) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.Methods("POST").Path("/v1/metrics").HandlerFunc(s.RecordMetrics)
	r.Methods("GET").Path("/v1/healthcheck").HandlerFunc(s.HealthCheck)

	return r
}
