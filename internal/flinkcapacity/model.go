package flinkcapacity

import (
	"fmt"
	"math"
)

// CU stores capacity in half-CU units. A value of 1 represents 0.5 CU.
type CU int64

func ParseCU(value float64) (CU, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		return 0, fmt.Errorf("CU must be a finite non-negative value, got %v", value)
	}
	scaled := value * 2
	rounded := math.Round(scaled)
	if math.Abs(scaled-rounded) > 1e-9 {
		return 0, fmt.Errorf("CU must be a multiple of 0.5, got %v", value)
	}
	return CU(rounded), nil
}

func (c CU) Float64() float64 {
	return float64(c) / 2
}

type Capacity struct {
	Fixed CU
	Limit CU
}

func (c Capacity) Validate() error {
	if c.Fixed < 0 {
		return fmt.Errorf("fixed CU must be non-negative, got %v", c.Fixed.Float64())
	}
	if c.Limit < 0 {
		return fmt.Errorf("CU limit must be non-negative, got %v", c.Limit.Float64())
	}
	if c.Limit < c.Fixed {
		return fmt.Errorf("CU limit %v must be greater than or equal to fixed CU %v", c.Limit.Float64(), c.Fixed.Float64())
	}
	return nil
}

func (c Capacity) Elastic() CU {
	return c.Limit - c.Fixed
}

func NewCapacity(fixed CU, elastic, max *CU) (Capacity, error) {
	if elastic != nil && max != nil {
		return Capacity{}, fmt.Errorf("elastic_cu_limit and max_cu_limit are mutually exclusive")
	}
	if fixed < 0 {
		return Capacity{}, fmt.Errorf("fixed CU must be non-negative, got %v", fixed.Float64())
	}

	capacity := Capacity{Fixed: fixed, Limit: fixed}
	if elastic != nil {
		if *elastic < 0 {
			return Capacity{}, fmt.Errorf("elastic CU must be non-negative, got %v", elastic.Float64())
		}
		capacity.Limit += *elastic
	}
	if max != nil {
		capacity.Limit = *max
	}
	if err := capacity.Validate(); err != nil {
		return Capacity{}, err
	}
	return capacity, nil
}

type WorkspaceCapacity struct {
	FixedCU          CU
	CrossZoneFixedCU CU
	Limit            CU
	Used             CU
}

type NotReadyError struct {
	Reason string
}

func (e *NotReadyError) Error() string   { return e.Reason }
func (e *NotReadyError) Retryable() bool { return true }

func (c WorkspaceCapacity) TotalFixed() CU {
	return c.FixedCU + c.CrossZoneFixedCU
}

func (c WorkspaceCapacity) AsCapacity() Capacity {
	return Capacity{Fixed: c.TotalFixed(), Limit: c.Limit}
}

type Queue struct {
	Name     string
	Capacity *Capacity
	Used     CU
}

type Namespace struct {
	Name     string
	Capacity *Capacity
	Used     CU
	Queues   []Queue
}

type Tree struct {
	ChargeType string
	Workspace  WorkspaceCapacity
	Namespaces []Namespace
}
