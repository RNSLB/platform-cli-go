package namespace

import (
	"testing"
)

func TestParseMemoryQuota(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantGB    float64
		wantError bool
	}{
		// Valid cases
		{"gigabytes", "20Gi", 20.0, false},
		{"megabytes", "512Mi", 0.5, false},
		{"kilobytes", "1024Ki", 0.001, false},
		{"terabytes", "1Ti", 1024.0, false},
		{"lowercase gi", "20gi", 20.0, false},
		{"with spaces", "  20Gi  ", 20.0, false},
		{"decimal", "1.5Gi", 1.5, false},
		
		// Edge cases
		{"no unit assumes GB", "20", 20.0, false},
		
		// Invalid cases
		{"empty string", "", 0, true},
		{"invalid unit", "20Xi", 0, true},
		{"negative value", "-10Gi", 0, true},
		{"invalid number", "abcGi", 0, true},
		{"just unit", "Gi", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMemoryQuota(tt.input)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseMemoryQuota(%q) expected error, got nil", tt.input)
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseMemoryQuota(%q) unexpected error: %v", tt.input, err)
				return
			}
			
			// Use approximate comparison for floats
			if diff := got - tt.wantGB; diff > 0.001 || diff < -0.001 {
				t.Errorf("ParseMemoryQuota(%q) = %v, want %v", tt.input, got, tt.wantGB)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	tests := []struct {
		name          string
		ns            Namespace
		wantCPUCost   float64
		wantMemCost   float64
		wantTotalCost float64
		wantError     bool
	}{
		{
			name: "standard namespace",
			ns: Namespace{
				Name:        "test",
				Team:        "engineering",
				CPUQuota:    10,
				MemoryQuota: "20Gi",
				Environment: "prod",
			},
			wantCPUCost:   730.0,  // 10 * 73
			wantMemCost:   200.0,  // 20 * 10
			wantTotalCost: 930.0,
			wantError:     false,
		},
		{
			name: "small namespace with Mi",
			ns: Namespace{
				Name:        "test-small",
				Team:        "data",
				CPUQuota:    1,
				MemoryQuota: "512Mi",
				Environment: "dev",
			},
			wantCPUCost:   73.0,   // 1 * 73
			wantMemCost:   5.0,    // 0.5 * 10
			wantTotalCost: 78.0,
			wantError:     false,
		},
		{
			name: "large namespace with Ti",
			ns: Namespace{
				Name:        "test-large",
				Team:        "ml",
				CPUQuota:    50,
				MemoryQuota: "1Ti",
				Environment: "prod",
			},
			wantCPUCost:   3650.0,   // 50 * 73
			wantMemCost:   10240.0,  // 1024 * 10
			wantTotalCost: 13890.0,
			wantError:     false,
		},
		{
			name: "invalid memory format",
			ns: Namespace{
				Name:        "test-invalid",
				Team:        "platform",
				CPUQuota:    10,
				MemoryQuota: "invalid",
				Environment: "dev",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalculateCost(tt.ns)
			
			if tt.wantError {
				if err == nil {
					t.Error("CalculateCost() expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Errorf("CalculateCost() unexpected error: %v", err)
				return
			}
			
			if got.CPUCost != tt.wantCPUCost {
				t.Errorf("CPUCost = %v, want %v", got.CPUCost, tt.wantCPUCost)
			}
			if got.MemoryCost != tt.wantMemCost {
				t.Errorf("MemoryCost = %v, want %v", got.MemoryCost, tt.wantMemCost)
			}
			if got.TotalCost != tt.wantTotalCost {
				t.Errorf("TotalCost = %v, want %v", got.TotalCost, tt.wantTotalCost)
			}
		})
	}
}

func TestCalculateAllCosts(t *testing.T) {
	namespaces := []Namespace{
		{
			Name:        "ns1",
			Team:        "engineering",
			CPUQuota:    10,
			MemoryQuota: "20Gi",
			Environment: "prod",
		},
		{
			Name:        "ns2",
			Team:        "data",
			CPUQuota:    5,
			MemoryQuota: "10Gi",
			Environment: "dev",
		},
	}

	costs, grandTotal, err := CalculateAllCosts(namespaces)
	
	if err != nil {
		t.Fatalf("CalculateAllCosts() unexpected error: %v", err)
	}

	if len(costs) != 2 {
		t.Errorf("got %d costs, want 2", len(costs))
	}

	// ns1: (10*73) + (20*10) = 730 + 200 = 930
	// ns2: (5*73) + (10*10) = 365 + 100 = 465
	// total: 1395
	wantTotal := 1395.0
	if grandTotal != wantTotal {
		t.Errorf("grandTotal = %v, want %v", grandTotal, wantTotal)
	}
}

// Benchmark memory parsing
func BenchmarkParseMemoryQuota(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseMemoryQuota("20Gi")
	}
}

// Benchmark cost calculation
func BenchmarkCalculateCost(b *testing.B) {
	ns := Namespace{
		Name:        "test",
		Team:        "engineering",
		CPUQuota:    10,
		MemoryQuota: "20Gi",
		Environment: "prod",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateCost(ns)
	}
}
