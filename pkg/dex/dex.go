//Package dex provides all the required middlewares and routines
//to send metrics data to the dex server
package dex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	dexHost = "https://dex.squadcast.com"
)

// Dex struct contains authentication data
// and other fields to store metrics from the middleware
type Dex struct {
	// LogAfter sets the number of metric points to be logged
	// to DEX servers in a single batch. By default 50 is used.
	// The value can be set according to the traffic generated for
	// the application.
	LogAfter int

	mutex *sync.Mutex

	serviceKey  string
	metrics     []Metric
	metricQueue chan Metric
	hostname    string

	sendMetrics bool
	latency     bool
	memory      bool
	statusCode  bool
}

type dexKeyDetail struct {
	Data struct {
		Name    string   `json:"name"`
		Metrics []string `json:"metrics"`
	} `json:"data"`
}

// New function is used to return a Dex instance initialized with the
// passed serviceKey
func New(serviceKey string) *Dex {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "default-" + uuid.New().String()
	}

	d := &Dex{
		LogAfter:    50,
		mutex:       &sync.Mutex{},
		serviceKey:  serviceKey,
		metrics:     make([]Metric, 0),
		metricQueue: make(chan Metric, 50),
		hostname:    hostname,
	}

	go pollKeyDetails(d)
	go d.start()
	return d
}

// pollKeyDetails is used to poll the key details every minute.
func pollKeyDetails(d *Dex) {
	for {
		err := setupDexKeyDetails(d)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			d.sendMetrics = false
		}
		<-time.After(time.Minute)
	}
}

func setupDexKeyDetails(d *Dex) error {
	req, err := http.NewRequest(http.MethodGet, dexHost+"/v1/detail", nil)
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", d.serviceKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("dex::setupDexKeyDetails : expected 200, got " + fmt.Sprint(resp.StatusCode))
	}

	kd := dexKeyDetail{}
	err = json.NewDecoder(resp.Body).Decode(&kd)
	if err != nil {
		return err
	}

	d.mutex.Lock()
	d.latency = stringInSlice("latency", kd.Data.Metrics)
	d.memory = stringInSlice("memory", kd.Data.Metrics)
	d.statusCode = stringInSlice("status_code", kd.Data.Metrics)
	d.sendMetrics = true
	d.mutex.Unlock()

	return nil
}

// Middleware method provides a middleware to be used with the http routers.
// This function records the metrics for a given route handler.
func (d *Dex) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d.handler(w, r, next)
	})
}

// start method starts the queue listener
// and sends a batch of n requests to the dex server
func (d *Dex) start() {
	for {
		d.metrics = append(d.metrics, <-d.metricQueue)

		if len(d.metrics) > d.LogAfter {
			go sendMetric(d.metrics, d.serviceKey)
			d.metrics = make([]Metric, 0)
		}
	}
}

// sendMetric is used to send metrics to the dex metric server
func sendMetric(ms []Metric, key string) {
	bs := bytes.NewBuffer(nil)
	err := json.NewEncoder(bs).Encode(map[string]interface{}{
		"payload": ms,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "dex::sendMetric : json.Encode : %s\n", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, dexHost+"/v1/metric", bs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dex::sendMetric : http.NewRequest : %s\n", err)
		return
	}

	req.Header.Set("X-API-Key", key)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dex::sendMetric : client.Do : %s\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		fmt.Fprintf(os.Stderr, "dex::sendMetric : response : expected 204, got %d\n", resp.StatusCode)
		io.Copy(os.Stderr, resp.Body)
		return
	}
}

func (d *Dex) handler(w http.ResponseWriter, r *http.Request, next http.Handler) {
	start := time.Now()
	mresp := &Response{w: w}

	next.ServeHTTP(mresp, r)

	del := time.Now().Sub(start)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if d.sendMetrics {
		go func(latency time.Duration, memUsage uint64, statusCode int, path, host string) {
			d.mutex.Lock()
			if d.latency {
				d.metricQueue <- NewMetric(start, d.hostname, path, host, "latency", latency.Nanoseconds())
			}

			if d.memory {
				d.metricQueue <- NewMetric(start, d.hostname, path, host, "memory", int64(memUsage))
			}

			if d.statusCode {
				d.metricQueue <- NewMetric(start, d.hostname, path, host, "status_code", int64(statusCode))
			}
			d.mutex.Unlock()
		}(del, m.Alloc, mresp.statusCode, r.URL.Path, strings.Split(r.Host, ":")[0])
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
