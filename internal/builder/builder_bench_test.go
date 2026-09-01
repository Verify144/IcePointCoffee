package builder

import "testing"

func BenchmarkBuildHouse(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := New()
		_, _ = b.Build("house", map[string]interface{}{"size": 10})
	}
}

func BenchmarkBuildTower(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := New()
		_, _ = b.Build("tower", map[string]interface{}{"size": 10})
	}
}

func BenchmarkBuildCircle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := New()
		_, _ = b.Build("circle", map[string]interface{}{"size": 10})
	}
}

func BenchmarkBuildSphere(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := New()
		_, _ = b.Build("sphere", map[string]interface{}{"size": 5})
	}
}

func BenchmarkBuildWall(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := New()
		_, _ = b.Build("wall", map[string]interface{}{"size": 10})
	}
}

func BenchmarkBuildFloor(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := New()
		_, _ = b.Build("floor", map[string]interface{}{"size": 10})
	}
}

func BenchmarkBuildRect(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := New()
		_, _ = b.Build("rect", map[string]interface{}{"size": 10})
	}
}
