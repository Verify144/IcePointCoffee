package metrics

import (
	"strings"
	"testing"
	"time"
)

func TestRegistryBasic(t *testing.T) {
	r := NewRegistry()

	r.Register("test_counter", "A test counter", TypeCounter)
	r.Inc("test_counter", 1)
	r.Inc("test_counter", 2)
	if r.metrics["test_counter"].Value != 3 {
		t.Errorf("Counter should be 3, got %f", r.metrics["test_counter"].Value)
	}

	r.Register("test_gauge", "A test gauge", TypeGauge)
	r.Set("test_gauge", 42.5)
	if r.metrics["test_gauge"].Value != 42.5 {
		t.Errorf("Gauge should be 42.5, got %f", r.metrics["test_gauge"].Value)
	}
}

func TestRegistryLabels(t *testing.T) {
	r := NewRegistry()
	r.Register("test_labeled", "A labeled counter", TypeCounter)

	labels := map[string]string{"path": "/api/test", "method": "GET"}
	r.AddLabel("test_labeled", labels, 1)

	key := labelKey("test_labeled", labels)
	if _, ok := r.metrics[key]; !ok {
		t.Error("Labeled metric should be registered")
	}

	r.AddLabel("test_labeled", labels, 1)
	if r.metrics[key].Value != 2 {
		t.Errorf("Should be 2, got %f", r.metrics[key].Value)
	}
}

func TestHistogram(t *testing.T) {
	r := NewRegistry()
	r.Register("test_histogram", "A test histogram", TypeHistogram)

	buckets := []float64{0.1, 0.5, 1.0}
	r.Observe("test_histogram", 0.05, buckets)
	r.Observe("test_histogram", 0.3, buckets)
	r.Observe("test_histogram", 0.8, buckets)
	r.Observe("test_histogram", 1.5, buckets)

	m := r.metrics["test_histogram"]
	if m.Count != 4 {
		t.Errorf("Count should be 4, got %d", m.Count)
	}
	if m.Sum != 2.65 {
		t.Errorf("Sum should be 2.65, got %f", m.Sum)
	}
	if m.Buckets["0.1"] != 1 {
		t.Errorf("Bucket 0.1 should be 1, got %f", m.Buckets["0.1"])
	}
	if m.Buckets["+Inf"] != 4 {
		t.Errorf("+Inf bucket should be 4, got %f", m.Buckets["+Inf"])
	}
}

func TestWritePrometheus(t *testing.T) {
	r := NewRegistry()
	r.Register("http_requests_total", "Total HTTP requests", TypeCounter)
	r.AddLabel("http_requests_total", map[string]string{"path": "/api/test", "method": "GET"}, 100)

	var sb strings.Builder
	err := r.WriteTo(&sb)
	if err != nil {
		t.Errorf("WriteTo should not error: %v", err)
	}

	output := sb.String()

	if !strings.Contains(output, "# HELP http_requests_total") {
		t.Error("Should contain HELP line")
	}
	if !strings.Contains(output, "# TYPE http_requests_total counter") {
		t.Error("Should contain TYPE line")
	}
	if !strings.Contains(output, "icepoint_uptime_seconds") {
		t.Error("Should contain uptime metric")
	}
	if !strings.Contains(output, "icepoint_go_goroutines") {
		t.Error("Should contain goroutines metric")
	}
	if !strings.Contains(output, "icepoint_go_memstats") {
		t.Error("Should contain memstats metrics")
	}
}

func TestBusinessMetrics(t *testing.T) {
	r := NewRegistry()
	b := NewBusinessMetrics(r)

	b.IncHTTPRequest("/api/test", "GET", 200)
	b.ObserveHTTPLatency("/api/test", 50*time.Millisecond)

	b.IncAIChat(true, 150)
	b.IncToolCall("echo")
	b.IncToolCall("calculate")

	b.IncCommand(true)

	b.IncBuild("house", 250)
	b.IncBuild("tower", 500)

	b.SetMCConnected(true)
	b.SetPluginCount(5)
	b.SetMemorySize(20)

	b.IncEvent("ai_chat")
	b.IncRateLimited()

	if r.metrics["icepoint_mc_connected"].Value != 1.0 {
		t.Errorf("MC connected should be 1.0, got %f", r.metrics["icepoint_mc_connected"].Value)
	}
	if r.metrics["icepoint_plugins_count"].Value != 5.0 {
		t.Errorf("Plugin count should be 5.0, got %f", r.metrics["icepoint_plugins_count"].Value)
	}
}

func TestGlobalRegistry(t *testing.T) {
	InitDefault()
	if DefaultRegistry == nil {
		t.Error("DefaultRegistry should not be nil")
	}
	if DefaultBusiness == nil {
		t.Error("DefaultBusiness should not be nil")
	}
}

func TestLabelKey(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
	}{
		{"empty", nil},
		{"single", map[string]string{"k": "v"}},
		{"multi", map[string]string{"b": "2", "a": "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labelKey("test", tt.labels)
			if got == "" || got == "test" && tt.labels != nil {
				t.Errorf("labelKey() should not be empty for %s", tt.name)
			}
		})
	}
}
