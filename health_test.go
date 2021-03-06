package health

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhawal55/health/mocks"
	"github.com/stretchr/testify/assert"
)

func empty() []string {
	return []string{}
}

func getServer(service *healthService) *httptest.Server {
	mux := service.registerRoute()
	return httptest.NewServer(mux)
}

func Test_Get(t *testing.T) {
	checker := mocks.HealthChecker{}
	messages := []string{"foo: bar"}
	checker.Mock.On("IsHealthy").Return(true, nil, messages)

	version := "1.0"
	checksum := "checksum"

	server := getServer(&healthService{
		healthCheckers: []HealthChecker{&checker},
		version:        version,
		checksum:       checksum,
	})
	res, _ := http.Get(server.URL + "/health")
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	expected := healthCheckResponse{
		OverallHealth: "Healthy",
		Version:       version,
		Checksum:      checksum,
		Items:         []healthCheckItem{healthCheckItem{Name: "*mocks.HealthChecker", Status: "Healthy", Error: "", Messages: messages}},
	}
	var actual healthCheckResponse
	json.Unmarshal(body, &actual)

	assert.Equal(t, expected.OverallHealth, "Healthy")
	assert.Equal(t, expected.Version, version)
	assert.Equal(t, expected.Items, actual.Items)
}

func Test_Get_WithError(t *testing.T) {
	checker := mocks.HealthChecker{}
	errMessage := "I failed"
	checker.Mock.On("IsHealthy").Return(false, errors.New(errMessage), nil)

	server := getServer(&healthService{healthCheckers: []HealthChecker{&checker}})
	res, _ := http.Get(server.URL + "/health")
	body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	expected := healthCheckResponse{
		OverallHealth: "Unhealthy",
		Items:         []healthCheckItem{healthCheckItem{Name: "*mocks.HealthChecker", Status: "Unhealthy", Error: errMessage, Messages: nil}},
	}
	var actual healthCheckResponse
	json.Unmarshal(body, &actual)

	assert.Equal(t, expected.OverallHealth, "Unhealthy")
	assert.Equal(t, expected.Items, actual.Items)
}

func Test_Get_Cors(t *testing.T) {
	server := getServer(&healthService{})
	defer server.Close()

	client := &http.Client{}

	req, _ := http.NewRequest("OPTIONS", server.URL+"/health", nil)
	req.Header.Add("Origin", "bar.com")

	res, err := client.Do(req)
	if err != nil {
		assert.Fail(t, "No error expected. Got error: %s", err.Error())
	}

	assertHeaders(t, res.Header, map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET",
	})
}

func assertHeaders(t *testing.T, resHeaders http.Header, reqHeaders map[string]string) {
	for name, value := range reqHeaders {
		if actual := strings.Join(resHeaders[name], ", "); actual != value {
			t.Errorf("Invalid header `%s', wanted `%s', got `%s'", name, value, actual)
		}
	}
}
