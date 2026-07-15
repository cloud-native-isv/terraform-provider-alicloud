package flinkcapacity

import "testing"

func TestParseCU(t *testing.T) {
	tests := []struct {
		name    string
		in      float64
		want    CU
		wantErr bool
	}{
		{name: "zero", in: 0, want: 0},
		{name: "half", in: 0.5, want: 1},
		{name: "one", in: 1, want: 2},
		{name: "one and half", in: 1.5, want: 3},
		{name: "unsupported fraction", in: 0.1, wantErr: true},
		{name: "negative", in: -0.5, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseCU(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ParseCU(%v) error = %v, wantErr %v", tc.in, err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("ParseCU(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestCUFloat64(t *testing.T) {
	if got := CU(3).Float64(); got != 1.5 {
		t.Fatalf("CU(3).Float64() = %v, want 1.5", got)
	}
}

func TestNewCapacityAliases(t *testing.T) {
	fixed := CU(4)
	elastic := CU(2)
	max := CU(6)

	byElastic, err := NewCapacity(fixed, &elastic, nil)
	if err != nil {
		t.Fatal(err)
	}
	byMax, err := NewCapacity(fixed, nil, &max)
	if err != nil {
		t.Fatal(err)
	}
	if byElastic != byMax {
		t.Fatalf("elastic alias result %+v != max alias result %+v", byElastic, byMax)
	}
	if byElastic != (Capacity{Fixed: 4, Limit: 6}) {
		t.Fatalf("unexpected capacity: %+v", byElastic)
	}

	if _, err := NewCapacity(fixed, &elastic, &max); err == nil {
		t.Fatal("expected elastic_cu_limit and max_cu_limit conflict")
	}
}

func TestNewCapacityValidation(t *testing.T) {
	negative := CU(-1)
	belowFixed := CU(1)

	if _, err := NewCapacity(-1, nil, nil); err == nil {
		t.Fatal("expected negative fixed CU to fail")
	}
	if _, err := NewCapacity(0, &negative, nil); err == nil {
		t.Fatal("expected negative elastic CU to fail")
	}
	if _, err := NewCapacity(2, nil, &belowFixed); err == nil {
		t.Fatal("expected max below fixed to fail")
	}
	got, err := NewCapacity(2, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != (Capacity{Fixed: 2, Limit: 2}) {
		t.Fatalf("zero elasticity default = %+v", got)
	}
}

func TestWorkspaceCapacity(t *testing.T) {
	w := WorkspaceCapacity{FixedCU: 0, CrossZoneFixedCU: 32, Limit: 64}
	if got := w.TotalFixed(); got != 32 {
		t.Fatalf("TotalFixed = %v, want 32", got)
	}
	if got := w.AsCapacity(); got != (Capacity{Fixed: 32, Limit: 64}) {
		t.Fatalf("AsCapacity = %+v", got)
	}
}
