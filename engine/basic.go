package engine

import (
	"fmt"
	"math"

	"github.com/hm-choi/pp-stat/utils"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

// Add performs element-wise homomorphic addition on two HEData inputs.
// It returns a new HEData object with the result.
// The function supports ciphertext slices of different lengths by padding the shorter one.
func (e *HEEngine) Add(ct1, ct2 *HEData) (*HEData, error) {
	// Determine output metadata: size, level, scale
	size := max(ct1.Size(), ct2.Size())
	level := min(ct1.Level(), ct2.Level())

	// Ensure both ciphertexts have the same scale
	if ct1.Scale() != ct2.Scale() {
		return nil, fmt.Errorf("scale mismatch: %f vs %f", ct1.Scale(), ct2.Scale())
	}
	scale := math.Min(ct1.Scale(), ct2.Scale())

	// Get ciphertext slices
	ctxts1 := ct1.Ciphertexts()
	ctxts2 := ct2.Ciphertexts()
	ctLen1 := len(ctxts1)
	ctLen2 := len(ctxts2)
	ctNum := max(ctLen1, ctLen2)

	// Prepare output slice
	result := make([]*rlwe.Ciphertext, ctNum)

	// Perform element-wise addition
	for i := 0; i < ctNum; i++ {
		switch {
		case i < ctLen1 && i < ctLen2:
			// Both slices have ciphertext at index i
			ct, err := e.Evaluator().AddNew(ctxts1[i], ctxts2[i])
			if err != nil {
				return nil, fmt.Errorf("AddNew failed at index %d: %w", i, err)
			}
			result[i] = ct

		case i >= ctLen1:
			// Only ct2 has a ciphertext
			result[i] = ctxts2[i]

		case i >= ctLen2:
			// Only ct1 has a ciphertext
			result[i] = ctxts1[i]
		}
	}

	return NewHEData(result, size, level, scale), nil
}

func (e *HEEngine) AddConst(ct *HEData, con float64) (*HEData, error) {
	// Determine output metadata: size, level, scale
	size := ct.Size()
	level := ct.Level()
	scale := ct.Scale()
	ctNum := len(ct.Ciphertexts())

	// Prepare output slice
	ctxts := make([]*rlwe.Ciphertext, ctNum)

	// Perform element-wise addition
	for i := 0; i < ctNum; i++ {
		ct, err := e.Evaluator().AddNew(ct.Ciphertexts()[i], con)
		if err != nil {
			return nil, fmt.Errorf("substraction failed at index %d: %w", i, err)
		}
		ctxts[i] = ct
	}

	return NewHEData(ctxts, size, level, scale), nil
}

// Sub performs element-wise homomorphic subtraction on two HEData inputs.
// It returns a new HEData object with the result.
// The function supports ciphertext slices of different lengths by padding the shorter one.
func (e *HEEngine) Sub(ct1, ct2 *HEData) (*HEData, error) {
	// Determine output metadata: size, level, scale
	size := max(ct1.Size(), ct2.Size())
	level := min(ct1.Level(), ct2.Level())

	// Ensure both ciphertexts have the same scale
	if ct1.Scale() != ct2.Scale() {
		return nil, fmt.Errorf("scale mismatch: %f vs %f", ct1.Scale(), ct2.Scale())
	}
	scale := math.Min(ct1.Scale(), ct2.Scale())

	// Get ciphertext slices
	ctxts1 := ct1.Ciphertexts()
	ctxts2 := ct2.Ciphertexts()
	ctLen1 := len(ctxts1)
	ctLen2 := len(ctxts2)
	ctNum := max(ctLen1, ctLen2)

	// Prepare output slice
	result := make([]*rlwe.Ciphertext, ctNum)

	// Perform element-wise addition
	for i := 0; i < ctNum; i++ {
		switch {
		case i < ctLen1 && i < ctLen2:
			// Both slices have ciphertext at index i
			ct, err := e.Evaluator().SubNew(ctxts1[i], ctxts2[i])
			if err != nil {
				return nil, fmt.Errorf("SubNew failed at index %d: %w", i, err)
			}
			result[i] = ct

		case i >= ctLen1:
			// Only ct2 has a ciphertext
			result[i] = ctxts2[i]

		case i >= ctLen2:
			// Only ct1 has a ciphertext
			result[i] = ctxts1[i]
		}
	}

	return NewHEData(result, size, level, scale), nil
}

func (e *HEEngine) SubConst(ct *HEData, con float64) (*HEData, error) {
	// Determine output metadata: size, level, scale
	size := ct.Size()
	level := ct.Level()
	scale := ct.Scale()
	ctNum := len(ct.Ciphertexts())

	// Prepare output slice
	ctxts := make([]*rlwe.Ciphertext, ctNum)

	// Perform element-wise addition
	for i := 0; i < ctNum; i++ {
		ct, err := e.Evaluator().SubNew(ct.Ciphertexts()[i], con)
		if err != nil {
			return nil, fmt.Errorf("substraction failed at index %d: %w", i, err)
		}
		ctxts[i] = ct
	}

	return NewHEData(ctxts, size, level, scale), nil
}

// Mult performs element-wise homomorphic multiplication with relinearization and rescaling.
func (e *HEEngine) Mult(ct1, ct2 *HEData) (*HEData, error) {
	// Determine output metadata: size, level, scale
	size := min(ct1.Size(), ct2.Size())
	level := min(ct1.Level(), ct2.Level())

	// Ensure both ciphertexts have the same scale
	if ct1.Scale() != ct2.Scale() {
		return nil, fmt.Errorf("scale mismatch: %f vs %f", ct1.Scale(), ct2.Scale())
	}
	scale := math.Min(ct1.Scale(), ct2.Scale())

	if level < 1 {
		return nil, fmt.Errorf("level cannot be smaller than 1")
	}
	// Get ciphertext slices
	ctxts1 := ct1.Ciphertexts()
	ctxts2 := ct2.Ciphertexts()
	ctLen1 := len(ctxts1)
	ctLen2 := len(ctxts2)
	ctNum := min(ctLen1, ctLen2)

	// Prepare output slice
	result := make([]*rlwe.Ciphertext, ctNum)

	// Perform element-wise addition
	for i := 0; i < ctNum; i++ {
		ct1 := ctxts1[i].CopyNew()
		ct2 := ctxts2[i].CopyNew()
		ct, err := e.Evaluator().MulRelinNew(ct1, ct2)
		if err != nil {
			return nil, fmt.Errorf("MulRelinNew failed at index %d: %w", i, err)
		}

		// Rescale to default scale
		if err = e.Evaluator().Rescale(ct, ct); err != nil {
			return nil, fmt.Errorf("Rescale failed at index %d: %w", i, err)
		}
		result[i] = ct
	}

	return NewHEData(result, size, level-1, scale), nil
}

// Mult performs element-wise homomorphic multiplication with relinearization and rescaling.
func (e *HEEngine) MultConst(ct *HEData, con float64) (*HEData, error) {
	// Determine output metadata: size, level, scale
	size := ct.Size()
	level := ct.Level()
	scale := ct.Scale()
	ctNum := len(ct.Ciphertexts())

	// Prepare output slice
	ctxts := make([]*rlwe.Ciphertext, ctNum)
	skipRescale := utils.IsPowerOfTwo(con)

	// Perform element-wise addition
	SIZE := ct.Size()
	// fmt.Println("SIZE", SIZE, e.params.MaxSlots())
	for i := 0; i < ctNum; i++ {
		consts := make([]float64, e.params.MaxSlots())
		if SIZE >= e.params.MaxSlots() {
			for j := 0; j < e.params.MaxSlots(); j++ {
				consts[j] = con
			}
			SIZE -= e.params.MaxSlots()
		} else {
			for j := 0; j < SIZE; j++ {
				consts[j] = con
			}
		}

		ctNew, err := e.Evaluator().MulNew(ct.Ciphertexts()[i], consts)
		if err != nil {
			return nil, fmt.Errorf("MulNew failed at index %d: %w", i, err)
		}
		// Rescale to default scale
		if !skipRescale {
			if err = e.Evaluator().Rescale(ctNew, ctNew); err != nil {
				return nil, fmt.Errorf("Rescale failed at index %d: %w", i, err)
			}
		}
		ctxts[i] = ctNew
	}
	if !skipRescale {
		level -= 1
	}

	return NewHEData(ctxts, size, level, scale), nil
}

// Sum performs a sum of all elements of input HEData.
func (e *HEEngine) Sum(ct *HEData) (result *HEData, err error) {
	ctxt := ct.Ciphertexts()[0].CopyNew()
	if len(ct.Ciphertexts()) > 1 {
		for i := 1; i < len(ct.Ciphertexts()); i++ {
			e.evaluator.Add(ctxt, ct.Ciphertexts()[i], ctxt)
		}
	}

	for i := 0; i < e.params.LogMaxSlots(); i++ {
		rot := 1 << i
		tmp, err := e.evaluator.RotateNew(ctxt, rot)
		if err != nil {
			return nil, fmt.Errorf("rotation failed at %d: %w", rot, err)
		}
		if err = e.evaluator.Add(ctxt, tmp, ctxt); err != nil {
			return nil, fmt.Errorf("addition failed: %w", err)
		}
	}

	ctxts := make([]*rlwe.Ciphertext, len(ct.Ciphertexts()))
	for i := 0; i < len(ct.Ciphertexts()); i++ {
		ctxts[i] = ctxt.CopyNew()
	}

	result = NewHEData(ctxts, ct.Size(), ct.Level(), ct.Scale())
	return result, nil
}

func (e *HEEngine) Mean(ct *HEData) (result *HEData, err error) {
	sumCtxt, err := e.Sum(ct)
	if err != nil {
		return nil, fmt.Errorf("summation failed: %w", err)
	}
	return e.MultConst(sumCtxt, 1.0/float64(ct.Size()))
}

func (e *HEEngine) Variance(ct *HEData) (result *HEData, err error) {
	// Step 1: Compute x²
	ctSquared, err := e.Mult(ct, ct)
	if err != nil {
		return nil, fmt.Errorf("failed to compute x²: %w", err)
	}

	// Step 2: Compute E[x²]
	meanXSquared, err := e.Mean(ctSquared)
	if err != nil {
		return nil, fmt.Errorf("failed to compute mean of x²: %w", err)
	}

	// Step 3: Compute E[x]
	meanX, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("failed to compute mean of x: %w", err)
	}

	// Step 4: Compute (E[x])²
	meanXSquared2, err := e.Mult(meanX, meanX)
	if err != nil {
		return nil, fmt.Errorf("failed to compute (E[x])²: %w", err)
	}

	return e.Sub(meanXSquared, meanXSquared2)
}
