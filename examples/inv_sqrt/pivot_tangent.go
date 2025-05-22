package invsqrt

import (
	"fmt"
	"math"

	"github.com/hm-choi/pp-stat/engine"
	"github.com/tuneinsight/lattigo/v6/circuits/ckks/minimax"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

type TwoLineParams struct {
	K1, K2, X1, X2, Pivot float64
}

func GetParams(d int) (*TwoLineParams, error) {
	switch d {
	case 7:
		return &TwoLineParams{0.128, 1.6645, 0.3111, 343.6645, 0.9053}, nil
	case 8:
		return &TwoLineParams{0.0855, 1.6876, 0.1322, 340.035, 0.3887}, nil
	case 9:
		return &TwoLineParams{0.0702, 1.6958, 0.0876, 338.781, 0.2587}, nil
	default:
		return nil, fmt.Errorf("invalid depth d=%d; must be one of 7, 8, or 9", d)
	}
}

func TwoLineApprox(e *engine.HEEngine, ct *engine.HEData, d int) (*engine.HEData, error) {
	params, err := GetParams(d)
	if err != nil {
		return nil, err
	}

	a, b := 0.001, 1000.0
	scale := 1.0 / (b - a)

	// Compute two tangents L1 and L2
	L1, err := e.MultConst(ct, -0.5*params.K2*math.Pow(params.X1, -1.5))
	if err != nil {
		return nil, fmt.Errorf("compute L1 slope: %w", err)
	}
	L1, err = e.AddConst(L1, 1.5*params.K2/math.Sqrt(params.X1))
	if err != nil {
		return nil, fmt.Errorf("add L1 bias: %w", err)
	}

	L2, err := e.MultConst(ct, -0.5*params.K2*math.Pow(params.X2, -1.5))
	if err != nil {
		return nil, fmt.Errorf("compute L2 slope: %w", err)
	}
	L2, err = e.AddConst(L2, 1.5*params.K2/math.Sqrt(params.X2))
	if err != nil {
		return nil, fmt.Errorf("add L2 bias: %w", err)
	}

	// Encrypt ones
	ctOnes, err := EncryptOneVector(e, ct.Size())
	if err != nil {
		return nil, fmt.Errorf("encrypt ones: %w", err)
	}

	// Step function input: normalized and centered
	stepInput, err := e.MultConst(ct, scale)
	if err != nil {
		return nil, fmt.Errorf("scale step input: %w", err)
	}
	stepInput, err = e.SubConst(stepInput, params.Pivot*scale)
	if err != nil {
		return nil, fmt.Errorf("center step input: %w", err)
	}

	// Step function evaluation using sign poly
	poly := minimax.NewPolynomial(engine.DefaultPolynomialForSign)
	eval := minimax.NewEvaluator(e.Params(), e.Evaluator(), e.BTS)
	stepEval := engine.NewEvaluator(e.Params(), eval, poly)

	stepCtxts := make([]*rlwe.Ciphertext, len(stepInput.Ciphertexts()))
	for i, c := range stepInput.Ciphertexts() {
		stepCtxts[i], err = stepEval.Step(c)
		if err != nil {
			return nil, fmt.Errorf("step evaluation failed at slot %d: %w", i, err)
		}
	}
	beta := engine.NewHEData(stepCtxts, ct.Size(), stepCtxts[0].Level(), ct.Scale())

	// Final interpolation: result = L1*(1-β) + L2*β
	oneMinusBeta, err := e.Sub(ctOnes, beta)
	if err != nil {
		return nil, fmt.Errorf("1 - beta: %w", err)
	}

	r1, err := e.Mult(L1, oneMinusBeta)
	if err != nil {
		return nil, fmt.Errorf("L1 × (1-β): %w", err)
	}
	r2, err := e.Mult(L2, beta)
	if err != nil {
		return nil, fmt.Errorf("L2 × β: %w", err)
	}

	result, err := e.Add(r1, r2)
	if err != nil {
		return nil, fmt.Errorf("L1(1-β) + L2β: %w", err)
	}

	return result, nil
}

func PivotTangent(e *engine.HEEngine, ct *engine.HEData, iter, d int) (*engine.HEData, error) {
	const (
		B              = 100.0
		bootstrapDepth = 10
	)

	y0, err := TwoLineApprox(e, ct, d)
	if err != nil {
		return nil, fmt.Errorf("TwoLineApprox failed: %w", err)
	}

	if e.IsBTS {
		y0, err = e.DoBootstrap(y0, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap(y0): %w", err)
		}
	}

	result, err := e.HENewtonInv(ct, y0, B, iter, 2)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv failed: %w", err)
	}

	return result, nil
}

func EncryptOneVector(e *engine.HEEngine, size int) (*engine.HEData, error) {
	vec := make([]float64, size)
	for i := range vec {
		vec[i] = 1.0
	}
	return e.Encrypt(vec)
}
