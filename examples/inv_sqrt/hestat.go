package invsqrt

import (
	"fmt"
	"math"

	"github.com/hm-choi/pp-stat/engine"
)

func HEStat(e *engine.HEEngine, ct *engine.HEData, iter int, B float64) (*engine.HEData, error) {
	// Step 1: Normalize ct = ct / B
	normCt, err := e.MultConst(ct, 1.0/B)
	if err != nil {
		return nil, fmt.Errorf("normalize input: %w", err)
	}

	// Step 2: Encrypt vector of 1s
	oneVec := make([]float64, normCt.Size())
	for i := range oneVec {
		oneVec[i] = 1.0
	}
	ctxtOne, err := e.Encrypt(oneVec)
	if err != nil {
		return nil, fmt.Errorf("encrypt one vector: %w", err)
	}

	// Step 3: Perform Newton's iteration for inverse square root
	invSqrt, err := e.HENewtonInv(normCt, ctxtOne, B, iter, 2)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 4: Final scaling adjustment
	scaledResult, err := e.MultConst(invSqrt, 1.0/math.Sqrt(B))
	if err != nil {
		return nil, fmt.Errorf("final scaling: %w", err)
	}

	return scaledResult, nil
}
