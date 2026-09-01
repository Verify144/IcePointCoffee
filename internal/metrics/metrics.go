// Package metrics provides lightweight Prometheus metrics export.
// Zero external dependencies, Prometheus text format v0.0.4 compliant.
package metrics

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// MetricType represents the Prometheus metric type.
type MetricType string

const (
	TypeCounter   MetricType = "counter"
	TypeGauge     MetricType = "gauge"
	TypeHistogram MetricType = "histogram"
)

// Metric holds a single metric's metadata and value.
type Metric struct {
	Name    string
	Help    string
	Type    MetricType
	Labels  map[string]string
	Value   float64
	Buckets map[string]float64
	Sum     float64
	Count   uint64
}

// Registry is the metrics registry.
type Registry struct {
	mu        sync.RWMutex
	metrics   map[string]*Metric
	startTime time.Time
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{
		metrics:   make(map[string]*Metric),
		startTime: time.Now(),
	}
}

// Register registers a new metric.
func (r *Registry) Register(name, help string, mtype MetricType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.metrics[name]; !exists {
		r.metrics[name] = &Metric{
			Name:    name,
			Help:    help,
			Type:    mtype,
			Labels:  make(map[string]string),
			Buckets: make(map[string]float64),
		}
	}
}

// Set sets a gauge value.
func (r *Registry) Set(name string, value float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.metrics[name]; ok {
		m.Value = value
	}
}

// Inc increments a counter.
func (r *Registry) Inc(name string, delta float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m, ok := r.metrics[name]; ok {
		m.Value += delta
	}
}

// AddLabel adds a labeled counter value.
func (r *Registry) AddLabel(name string, labels map[string]string, value float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := labelKey(name, labels)
	if _, ok := r.metrics[key]; !ok {
		orig, exists := r.metrics[name]
		if !exists {
			return
		}
		m := &Metric{
			Name:    name,
			Help:    orig.Help,
			Type:    orig.Type,
			Labels:  labels,
			Buckets: make(map[string]float64),
		}
		r.metrics[key] = m
	}
	r.metrics[key].Value += value
}

// SetLabel sets a labeled gauge value.
func (r *Registry) SetLabel(name string, labels map[string]string, value float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := labelKey(name, labels)
	if _, ok := r.metrics[key]; !ok {
		orig, exists := r.metrics[name]
		if !exists {
			return
		}
		m := &Metric{
			Name:    name,
			Help:    orig.Help,
			Type:    orig.Type,
			Labels:  labels,
			Buckets: make(map[string]float64),
		}
		r.metrics[key] = m
	}
	r.metrics[key].Value = value
}

// Observe records a histogram observation.
func (r *Registry) Observe(name string, value float64, buckets []float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.metrics[name]
	if !ok {
		return
	}
	m.Sum += value
	m.Count++
	for _, b := range buckets {
		if value <= b {
			m.Buckets[fmt.Sprintf("%g", b)]++
		}
	}
	m.Buckets["+Inf"]++
}

// labelKey generates a unique key for a labeled metric.
func labelKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", k, labels[k]))
	}
	return fmt.Sprintf("%s{%s}", name, strings.Join(parts, ","))
}

// WriteTo writes metrics in Prometheus text format.
func (r *Registry) WriteTo(w io.Writer) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Group by metric name
	grouped := make(map[string][]*Metric)
	for _, m := range r.metrics {
		grouped[m.Name] = append(grouped[m.Name], m)
	}
	names := make([]string, 0, len(grouped))
	for n := range grouped {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		metrics := grouped[name]
		if len(metrics) == 0 {
			continue
		}
		first := metrics[0]
		fmt.Fprintf(w, "# HELP %s %s\n", first.Name, first.Help)
		fmt.Fprintf(w, "# TYPE %s %s\n", first.Name, first.Type)

		switch first.Type {
		case TypeCounter, TypeGauge:
			for _, m := range metrics {
				fmt.Fprintf(w, "%s%s %g\n", m.Name, formatLabels(m.Labels), m.Value)
			}
		case TypeHistogram:
			for _, m := range metrics {
				keys := make([]string, 0, len(m.Buckets))
				for k := range m.Buckets {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					bLabel := make(map[string]string)
					for k2, v := range m.Labels {
						bLabel[k2] = v
					}
					bLabel["le"] = k
					fmt.Fprintf(w, "%s_bucket%s %g\n", m.Name, formatLabels(bLabel), m.Buckets[k])
				}
				fmt.Fprintf(w, "%s_sum%s %g\n", m.Name, formatLabels(m.Labels), m.Sum)
				fmt.Fprintf(w, "%s_count%s %d\n", m.Name, formatLabels(m.Labels), m.Count)
			}
		}
	}

	// Go runtime metrics
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	fmt.Fprintf(w, "# HELP icepoint_uptime_seconds IcePoint uptime in seconds\n")
	fmt.Fprintf(w, "# TYPE icepoint_uptime_seconds counter\n")
	fmt.Fprintf(w, "icepoint_uptime_seconds %g\n", time.Since(r.startTime).Seconds())

	fmt.Fprintf(w, "# HELP icepoint_go_memstats_alloc_bytes Go allocated memory\n")
	fmt.Fprintf(w, "# TYPE icepoint_go_memstats_alloc_bytes gauge\n")
	fmt.Fprintf(w, "icepoint_go_memstats_alloc_bytes %d\n", mem.Alloc)

	fmt.Fprintf(w, "# HELP icepoint_go_memstats_heap_bytes Go heap memory\n")
	fmt.Fprintf(w, "# TYPE icepoint_go_memstats_heap_bytes gauge\n")
	fmt.Fprintf(w, "icepoint_go_memstats_heap_bytes %d\n", mem.HeapAlloc)

	fmt.Fprintf(w, "# HELP icepoint_go_memstats_sys_bytes Go system memory\n")
	fmt.Fprintf(w, "# TYPE icepoint_go_memstats_sys_bytes gauge\n")
	fmt.Fprintf(w, "icepoint_go_memstats_sys_bytes %d\n", mem.Sys)

	fmt.Fprintf(w, "# HELP icepoint_go_goroutines Current goroutines\n")
	fmt.Fprintf(w, "# TYPE icepoint_go_goroutines gauge\n")
	fmt.Fprintf(w, "icepoint_go_goroutines %d\n", runtime.NumGoroutine())

	fmt.Fprintf(w, "# HELP icepoint_go_gc_count Garbage collection count\n")
	fmt.Fprintf(w, "# TYPE icepoint_go_gc_count counter\n")
	fmt.Fprintf(w, "icepoint_go_gc_count %d\n", mem.NumGC)

	fmt.Fprintf(w, "# HELP icepoint_go_gc_pause_total_ns Total GC pause nanoseconds\n")
	fmt.Fprintf(w, "# TYPE icepoint_go_gc_pause_total_ns counter\n")
	fmt.Fprintf(w, "icepoint_go_gc_pause_total_ns %d\n", mem.PauseTotalNs)

	return nil
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%q", k, labels[k]))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ","))
}

// StandardBuckets are the default histogram buckets.
var StandardBuckets = []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

// BusinessMetrics wraps business-level metrics.
type BusinessMetrics struct {
	registry *Registry
}

// NewBusinessMetrics creates business metrics.
func NewBusinessMetrics(r *Registry) *BusinessMetrics {
	r.Register("icepoint_http_requests_total", "Total HTTP requests", TypeCounter)
	r.Register("icepoint_http_request_duration_seconds", "HTTP request latency", TypeHistogram)
	r.Register("icepoint_http_errors_total", "Total HTTP errors", TypeCounter)
	r.Register("icepoint_ai_chat_total", "Total AI chat calls", TypeCounter)
	r.Register("icepoint_ai_tokens_used", "AI tokens consumed", TypeCounter)
	r.Register("icepoint_ai_tool_calls_total", "AI tool invocations", TypeCounter)
	r.Register("icepoint_commands_total", "Total MC commands executed", TypeCounter)
	r.Register("icepoint_build_structures_total", "Total structures built", TypeCounter)
	r.Register("icepoint_build_blocks_total", "Total blocks placed", TypeCounter)
	r.Register("icepoint_mc_connected", "MC connection status (1=connected)", TypeGauge)
	r.Register("icepoint_plugins_count", "Number of registered plugins", TypeGauge)
	r.Register("icepoint_memory_size", "AI memory messages count", TypeGauge)
	r.Register("icepoint_events_total", "Total events processed", TypeCounter)
	r.Register("icepoint_rate_limited_total", "Total rate limited requests", TypeCounter)
	return &BusinessMetrics{registry: r}
}

// IncHTTPRequest records an HTTP request.
func (b *BusinessMetrics) IncHTTPRequest(path, method string, status int) {
	b.registry.AddLabel("icepoint_http_requests_total", map[string]string{
		"path":   path,
		"method": method,
		"status": fmt.Sprintf("%d", status),
	}, 1)
	if status >= 400 {
		b.registry.AddLabel("icepoint_http_errors_total", map[string]string{
			"path":   path,
			"status": fmt.Sprintf("%d", status),
		}, 1)
	}
}

// ObserveHTTPLatency records HTTP latency.
func (b *BusinessMetrics) ObserveHTTPLatency(path string, duration time.Duration) {
	b.registry.Observe("icepoint_http_request_duration_seconds", duration.Seconds(), StandardBuckets)
}

// IncAIChat records an AI chat call.
func (b *BusinessMetrics) IncAIChat(success bool, tokens int) {
	status := "success"
	if !success {
		status = "error"
	}
	b.registry.AddLabel("icepoint_ai_chat_total", map[string]string{"status": status}, 1)
	if tokens > 0 {
		b.registry.AddLabel("icepoint_ai_tokens_used", map[string]string{"status": status}, float64(tokens))
	}
}

// IncToolCall records a tool invocation.
func (b *BusinessMetrics) IncToolCall(toolName string) {
	b.registry.AddLabel("icepoint_ai_tool_calls_total", map[string]string{"tool": toolName}, 1)
}

// IncCommand records a command execution.
func (b *BusinessMetrics) IncCommand(success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	b.registry.AddLabel("icepoint_commands_total", map[string]string{"status": status}, 1)
}

// IncBuild records a structure build.
func (b *BusinessMetrics) IncBuild(structureType string, blocks int) {
	b.registry.AddLabel("icepoint_build_structures_total", map[string]string{"type": structureType}, 1)
	b.registry.AddLabel("icepoint_build_blocks_total", map[string]string{"type": structureType}, float64(blocks))
}

// SetMCConnected sets the MC connection status.
func (b *BusinessMetrics) SetMCConnected(connected bool) {
	val := 0.0
	if connected {
		val = 1.0
	}
	b.registry.Set("icepoint_mc_connected", val)
}

// SetPluginCount sets the plugin count.
func (b *BusinessMetrics) SetPluginCount(n int) {
	b.registry.Set("icepoint_plugins_count", float64(n))
}

// SetMemorySize sets the memory size.
func (b *BusinessMetrics) SetMemorySize(n int) {
	b.registry.Set("icepoint_memory_size", float64(n))
}

// IncEvent records an event.
func (b *BusinessMetrics) IncEvent(eventType string) {
	b.registry.AddLabel("icepoint_events_total", map[string]string{"type": eventType}, 1)
}

// IncRateLimited records a rate limited request.
func (b *BusinessMetrics) IncRateLimited() {
	b.registry.Inc("icepoint_rate_limited_total", 1)
}

// Default registry singletons.
var (
	DefaultRegistry *Registry
	DefaultBusiness *BusinessMetrics
	metricsOnce    sync.Once
)

// InitDefault initializes the default registry.
func InitDefault() {
	metricsOnce.Do(func() {
		DefaultRegistry = NewRegistry()
		DefaultBusiness = NewBusinessMetrics(DefaultRegistry)
	})
}
