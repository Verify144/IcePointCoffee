package builder

import "testing"

func TestBuilder(t *testing.T) {
	b := New()

	tests := []struct {
		typ string
		size int
	}{
		{"house", 10},
		{"tower", 5},
		{"circle", 5},
		{"sphere", 3},
		{"wall", 8},
		{"floor", 10},
		{"rect", 6},
	}

	for _, tt := range tests {
		b2 := New()
		_, err := b2.Build(tt.typ, map[string]interface{}{"size": tt.size})
		if err != nil {
			t.Errorf("Build %s failed: %v", tt.typ, err)
		}
	}

	_ = b
}

func TestUnknownType(t *testing.T) {
	b := New()
	_, err := b.Build("unknown_type", nil)
	if err == nil {
		t.Error("Should fail for unknown type")
	}
}
