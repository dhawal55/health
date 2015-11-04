package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"

	"git.nordstrom.net/rfid/version"
)

// Interface for retrieving health status of various components.
type HealthChecker interface {
	IsHealthy() (bool, error, []string)
}

type healthService struct {
	healthCheckers []HealthChecker
	version        string
	checksum       string
}

type healthCheckReponse struct {
	isHealthy       bool
	OverallHealth   string            `json:"overallHealth"`
	Version         string            `json:"version"`
	Checksum        string            `json:"checksum"`
	Hostname        string            `json:"hostname"`
	GoRoutinesCount int               `json:"goRoutinesCount"`
	Items           []healthCheckItem `json:"items"`
}

type healthCheckItem struct {
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	Error    string   `json:"error"`
	Messages []string `json:"messages"`
}

func New(checkers []HealthChecker, v version.Versioner) *http.ServeMux {
	service := &healthService{healthCheckers: checkers, version: v.GetVersion(), checksum: version.GetChecksum()}
	return service.registerRoute()
}

func (s *healthService) registerRoute() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/health", version.CorsHandler(s))

	return mux
}

func (s *healthService) getHealthReport() healthCheckReponse {
	var checkItems []healthCheckItem
	overall := true

	var errmsg string
	for _, checker := range s.healthCheckers {
		status, err, messages := checker.IsHealthy()
		if err != nil {
			errmsg = err.Error()
		}

		item := healthCheckItem{
			Name:     reflect.TypeOf(checker).String(),
			Status:   getStatus(status),
			Error:    errmsg,
			Messages: messages,
		}

		checkItems = append(checkItems, item)
		if status == false {
			overall = false
			fmt.Printf("Service not healthy: %+v\n", item)
		}
	}

	hostname, _ := os.Hostname()
	return healthCheckReponse{
		Items:           checkItems,
		isHealthy:       overall,
		OverallHealth:   getStatus(overall),
		Version:         s.version,
		Checksum:        s.checksum,
		GoRoutinesCount: runtime.NumGoroutine(),
		Hostname:        hostname,
	}
}

func (s *healthService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	report := s.getHealthReport()
	statusCode := http.StatusOK
	if !report.isHealthy {
		statusCode = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(report)
}

func getStatus(status bool) string {
	if status {
		return "Healthy"
	}

	return "Unhealthy"
}
