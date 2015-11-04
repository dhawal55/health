# health

Package 'health' returns a HTTP request multiplexer with a "/health" endpoint that returns the application's health.

## Usage

Implement the HealthChecker interface and pass it to the New method. Add the returned handler to your request multiplexer.

```
package main

import (
    "errors"
    "log"
    "net/http"

    "github.com/dhawal55/health"
    "github.com/dhawal55/version"
)

type Dependency1 struct{}

func (d *Dependency1) IsHealthy() (bool, error, []string) {
    //Perform health check logic
    return true, nil, []string{"RequestTime: 10ms"}
}

type Dependency2 struct{}

func (d *Dependency2) IsHealthy() (bool, error, []string) {
    //Perform health check logic
    return false, errors.New("Cannot connect"), nil
}

type Version struct{}

func (v *Version) GetVersion() string {
    return "1.0"
}

func main() {
    d1 := &Dependency1{}
    d2 := &Dependency2{}
    checkers := []health.HealthChecker{d1, d2}
    v := &Version{}

    mux := http.NewServeMux()
    //Add application handlers
    //mux.Handle("/users", userHandler)

    versionMux := version.New(v)
    healthMux := health.New(checkers, v)
    //Add versionMux to your application mux
    mux.Handle("/", versionMux)
    //Add healthMux to versionMux to avoid duplicate routes in application mux
    versionMux.Handle("/", healthMux)

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```
