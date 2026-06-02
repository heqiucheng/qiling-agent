package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type endpoint struct {
	Method string
	Path   string
	Body   func(int64) string
}

func (e endpoint) key() string {
	return e.Method + " " + e.Path
}

type result struct {
	Scenario      string         `json:"scenario"`
	BaseURL       string         `json:"base_url"`
	Duration      string         `json:"duration"`
	Concurrency   int            `json:"concurrency"`
	TotalRequests int64          `json:"total_requests"`
	Success       int64          `json:"success"`
	Errors        int64          `json:"errors"`
	RPS           float64        `json:"rps"`
	Latency       latencySummary `json:"latency_ms"`
	ByEndpoint    []endpointStat `json:"by_endpoint"`
	StatusCodes   map[int]int64  `json:"status_codes"`
	StartedAt     string         `json:"started_at"`
	FinishedAt    string         `json:"finished_at"`
}

type endpointStat struct {
	Endpoint      string         `json:"endpoint"`
	TotalRequests int64          `json:"total_requests"`
	Success       int64          `json:"success"`
	Errors        int64          `json:"errors"`
	Latency       latencySummary `json:"latency_ms"`
}

type endpointAccumulator struct {
	Total     int64
	Success   int64
	Errors    int64
	Latencies []float64
}

type latencySummary struct {
	P50 float64 `json:"p50"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
	Max float64 `json:"max"`
}

func main() {
	baseURL := flag.String("base-url", "http://127.0.0.1:8080", "backend base URL")
	duration := flag.Duration("duration", 30*time.Second, "load test duration")
	concurrency := flag.Int("concurrency", 16, "concurrent workers")
	scenario := flag.String("scenario", "read", "scenario: read, upload, report, report-read, report-export, or report-generate")
	timeout := flag.Duration("timeout", 5*time.Second, "single request timeout")
	userID := flag.String("user-id", "mgr_001", "X-Qiling-User-ID header")
	role := flag.String("role", "manager", "X-Qiling-Role header")
	flag.Parse()

	if *concurrency <= 0 {
		log.Fatal("concurrency must be greater than zero")
	}
	if *duration <= 0 {
		log.Fatal("duration must be greater than zero")
	}

	endpoints, err := scenarioEndpoints(*scenario)
	if err != nil {
		log.Fatal(err)
	}
	if strings.HasPrefix(*scenario, "report") {
		reportID, err := prepareReportScenario(*baseURL, *timeout, *userID, *role)
		if err != nil {
			log.Fatal(err)
		}
		endpoints = reportEndpoints(*scenario, reportID)
	}

	startedAt := time.Now().UTC()
	deadline := time.Now().Add(*duration)
	client := &http.Client{Timeout: *timeout}

	var total int64
	var success int64
	var errors int64
	var sequence int64
	var mu sync.Mutex
	latencies := make([]float64, 0, *concurrency*128)
	statusCodes := map[int]int64{}
	endpointStats := map[string]*endpointAccumulator{}

	var wg sync.WaitGroup
	for workerID := 0; workerID < *concurrency; workerID++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for time.Now().Before(deadline) {
				requestID := atomic.AddInt64(&sequence, 1)
				target := endpoints[int(requestID)%len(endpoints)]
				elapsed, statusCode, err := execute(client, *baseURL, target, requestID, *userID, *role)

				atomic.AddInt64(&total, 1)
				if err != nil || statusCode >= http.StatusBadRequest {
					atomic.AddInt64(&errors, 1)
				} else {
					atomic.AddInt64(&success, 1)
				}

				mu.Lock()
				latencies = append(latencies, float64(elapsed.Microseconds())/1000)
				statusCodes[statusCode]++
				recordEndpoint(endpointStats, target.key(), elapsed, err == nil && statusCode < http.StatusBadRequest)
				mu.Unlock()
			}
		}(workerID)
	}
	wg.Wait()
	finishedAt := time.Now().UTC()

	mu.Lock()
	latency := summarizeLatency(latencies)
	codes := make(map[int]int64, len(statusCodes))
	for code, count := range statusCodes {
		codes[code] = count
	}
	byEndpoint := summarizeEndpoints(endpointStats)
	mu.Unlock()

	report := result{
		Scenario:      *scenario,
		BaseURL:       strings.TrimRight(*baseURL, "/"),
		Duration:      duration.String(),
		Concurrency:   *concurrency,
		TotalRequests: total,
		Success:       success,
		Errors:        errors,
		RPS:           float64(total) / finishedAt.Sub(startedAt).Seconds(),
		Latency:       latency,
		ByEndpoint:    byEndpoint,
		StatusCodes:   codes,
		StartedAt:     startedAt.Format(time.RFC3339),
		FinishedAt:    finishedAt.Format(time.RFC3339),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		log.Fatal(err)
	}
}

func scenarioEndpoints(name string) ([]endpoint, error) {
	switch name {
	case "read":
		return []endpoint{
			{Method: http.MethodGet, Path: "/api/health"},
			{Method: http.MethodGet, Path: "/api/dashboard/summary"},
			{Method: http.MethodGet, Path: "/api/customers?page=1&page_size=20"},
			{Method: http.MethodGet, Path: "/api/customers/cus_001"},
			{Method: http.MethodGet, Path: "/api/customers/cus_001/conversations?page=1&page_size=20"},
			{Method: http.MethodGet, Path: "/api/followup-tasks?status=pending&page=1&page_size=20"},
			{Method: http.MethodGet, Path: "/api/review-reports/summary"},
		}, nil
	case "upload":
		return []endpoint{
			{
				Method: http.MethodPost,
				Path:   "/api/uploads/conversations",
				Body: func(sequence int64) string {
					return fmt.Sprintf(`{"source_type":"pasted_text","content":"压测客户%d 10:20 价格和效果需要再看看","owner_id":"usr_001"}`, sequence)
				},
			},
		}, nil
	case "report", "report-read", "report-export", "report-generate":
		return []endpoint{}, nil
	default:
		return nil, fmt.Errorf("unsupported scenario %q", name)
	}
}

func reportEndpoints(scenario string, reportID string) []endpoint {
	readEndpoints := []endpoint{
		{Method: http.MethodGet, Path: "/api/reports?page=1&page_size=20"},
		{Method: http.MethodGet, Path: "/api/reports/" + reportID},
	}
	exportEndpoints := []endpoint{
		{Method: http.MethodGet, Path: "/api/reports/" + reportID + "/export?format=markdown"},
		{Method: http.MethodGet, Path: "/api/reports/" + reportID + "/export?format=xlsx"},
		{Method: http.MethodGet, Path: "/api/reports/" + reportID + "/export?format=pdf"},
	}
	generateEndpoints := []endpoint{
		{
			Method: http.MethodPost,
			Path:   "/api/reports/customer-intent",
			Body: func(sequence int64) string {
				return `{"range":"last_7_days"}`
			},
		},
	}
	switch scenario {
	case "report-read":
		return readEndpoints
	case "report-export":
		return exportEndpoints
	case "report-generate":
		return generateEndpoints
	default:
		result := make([]endpoint, 0, len(readEndpoints)+len(exportEndpoints)+len(generateEndpoints))
		result = append(result, readEndpoints...)
		result = append(result, exportEndpoints...)
		result = append(result, generateEndpoints...)
		return result
	}
}

func prepareReportScenario(baseURL string, timeout time.Duration, userID string, role string) (string, error) {
	client := &http.Client{Timeout: timeout}
	_, statusCode, body, err := executeWithBody(client, strings.TrimRight(baseURL, "/")+"/api/reports/customer-intent", http.MethodPost, `{"range":"last_7_days"}`, userID, role)
	if err != nil {
		return "", err
	}
	if statusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("prepare report scenario failed with status %d: %s", statusCode, body)
	}

	var response struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &response); err != nil {
		return "", err
	}
	if response.Data.ID == "" {
		return "", fmt.Errorf("prepare report scenario response missing report id")
	}
	return response.Data.ID, nil
}

func execute(client *http.Client, baseURL string, target endpoint, sequence int64, userID string, role string) (time.Duration, int, error) {
	url := strings.TrimRight(baseURL, "/") + target.Path
	var body io.Reader
	if target.Body != nil {
		body = bytes.NewBufferString(target.Body(sequence))
	}

	req, err := http.NewRequest(target.Method, url, body)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Qiling-User-ID", userID)
	req.Header.Set("X-Qiling-Role", role)

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return elapsed, 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return elapsed, resp.StatusCode, nil
}

func executeWithBody(client *http.Client, url string, method string, payload string, userID string, role string) (time.Duration, int, string, error) {
	var body io.Reader
	if payload != "" {
		body = bytes.NewBufferString(payload)
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, 0, "", err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Qiling-User-ID", userID)
	req.Header.Set("X-Qiling-Role", role)

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return elapsed, 0, "", err
	}
	defer resp.Body.Close()
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return elapsed, resp.StatusCode, "", readErr
	}
	return elapsed, resp.StatusCode, string(responseBody), nil
}

func summarizeLatency(values []float64) latencySummary {
	if len(values) == 0 {
		return latencySummary{}
	}

	sort.Float64s(values)
	return latencySummary{
		P50: percentile(values, 0.50),
		P95: percentile(values, 0.95),
		P99: percentile(values, 0.99),
		Max: values[len(values)-1],
	}
}

func recordEndpoint(stats map[string]*endpointAccumulator, key string, elapsed time.Duration, ok bool) {
	current, exists := stats[key]
	if !exists {
		current = &endpointAccumulator{Latencies: make([]float64, 0, 128)}
		stats[key] = current
	}
	current.Total++
	if ok {
		current.Success++
	} else {
		current.Errors++
	}
	current.Latencies = append(current.Latencies, float64(elapsed.Microseconds())/1000)
}

func summarizeEndpoints(stats map[string]*endpointAccumulator) []endpointStat {
	result := make([]endpointStat, 0, len(stats))
	for key, current := range stats {
		result = append(result, endpointStat{
			Endpoint:      key,
			TotalRequests: current.Total,
			Success:       current.Success,
			Errors:        current.Errors,
			Latency:       summarizeLatency(current.Latencies),
		})
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Endpoint < result[right].Endpoint
	})
	return result
}

func percentile(values []float64, ratio float64) float64 {
	if len(values) == 0 {
		return 0
	}
	index := int(float64(len(values)-1) * ratio)
	return values[index]
}
