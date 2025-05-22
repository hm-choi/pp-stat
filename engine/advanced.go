package engine

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/minimax"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func (e *HEEngine) ZScoreNorm(ct *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2 // Degree for initial Chebyshev approximation
		newtonIter      = 5 // Iteration count for Newton refinement
		newtonScale     = 2 // Scaling for Newton method
		bootstrapDepth  = 3 // Depth used when bootstrapping initial guess
	)

	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("compute mean: %w", err)
	}

	// Step 2: Approximate variance Var(X)
	denom := float64(ct.Size()) * B
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("compute variance (approx): %w", err)
	}

	// Step 3: Center data => X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center input: %w", err)
	}

	// Select representative ciphertext for approximation
	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance approx): %w", err)
	}

	// Step 4: Initial guess for 1/σ using Chebyshev
	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebyshevDegree, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}

	// Optional: Bootstrap the initial guess for higher precision
	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (invSqrt init): %w", err)
		}
	}

	// Step 5: Refine variance and compute Newton-based 1/σ
	varianceRefined, err := e.Variance(ct)
	if err != nil {
		return nil, fmt.Errorf("compute refined variance: %w", err)
	}
	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance refined): %w", err)
	}

	invSigmaRefined, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 6: Extend 1/σ to all slots for element-wise multiplication
	invSigmaSlots, err := e.extendOneToMulty(invSigmaRefined, len(centered.Ciphertexts()), centered.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	// Step 7: Final Z-score normalization: Z = (X - μ) × (1/σ)
	zscore, err := e.Mult(centered, invSigmaSlots)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return zscore, nil
}

func (e *HEEngine) Kurtosis(ct *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonIter      = 5
		newtonScale     = 2
		bootstrapDepth  = 3
		subtractBias    = 3.0 // for excess kurtosis
	)

	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("mean: %w", err)
	}

	// Step 2: Centered input: x = X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center: %w", err)
	}

	// Step 3: Compute x² and x⁴
	x2, err := e.Mult(centered, centered)
	if err != nil {
		return nil, fmt.Errorf("x^2: %w", err)
	}

	x4, err := e.Mult(x2, x2)
	if err != nil {
		return nil, fmt.Errorf("x^4: %w", err)
	}

	// Step 4: Compute E[x⁴] (numerator)
	numerator, err := e.Mean(x4)
	if err != nil {
		return nil, fmt.Errorf("mean of x^4: %w", err)
	}

	// Step 5: Approximate variance Var(X)
	denom := float64(ct.Size()) * B
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("variance (approx): %w", err)
	}

	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance approx): %w", err)
	}

	// Step 6: Initial approximation of 1/σ using Chebyshev
	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebyshevDegree, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}

	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (Chebyshev init): %w", err)
		}
	}

	// Step 7: Refine 1/σ using Newton method
	varianceRefined, err := e.Variance(ct)
	if err != nil {
		return nil, fmt.Errorf("variance (refined): %w", err)
	}

	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance refined): %w", err)
	}

	invSigma, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 8: Compute 1/σ⁴ = (1/σ)² × (1/σ)²
	invSigma2, err := e.Mult(invSigma, invSigma)
	if err != nil {
		return nil, fmt.Errorf("inv σ²: %w", err)
	}

	invSigma4, err := e.Mult(invSigma2, invSigma2)
	if err != nil {
		return nil, fmt.Errorf("inv σ⁴: %w", err)
	}

	// Step 9: Apply inv(σ⁴) to numerator: E[x⁴] × (1/σ⁴)
	invSigma4Expanded, err := e.extendOneToMulty(invSigma4, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty (inv σ⁴): %w", err)
	}

	kurtosis, err := e.Mult(numerator, invSigma4Expanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	// Step 10: Convert to excess kurtosis: K − 3
	kurtosis, err = e.SubConst(kurtosis, subtractBias)
	if err != nil {
		return nil, fmt.Errorf("subtract bias: %w", err)
	}

	return kurtosis, nil
}

func (e *HEEngine) Skewness(ct *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonIter      = 5
		newtonScale     = 2
		bootstrapDepth  = 3
	)

	// Step 1: Compute mean μ
	mean, err := e.Mean(ct)
	if err != nil {
		return nil, fmt.Errorf("compute mean: %w", err)
	}

	// Step 2: Centered data x = X - μ
	centered, err := e.Sub(ct, mean)
	if err != nil {
		return nil, fmt.Errorf("center data: %w", err)
	}

	// Step 3: Compute x² and x³
	x2, err := e.Mult(centered, centered)
	if err != nil {
		return nil, fmt.Errorf("compute x^2: %w", err)
	}

	x3, err := e.Mult(centered, x2)
	if err != nil {
		return nil, fmt.Errorf("compute x^3: %w", err)
	}

	// Step 4: Compute E[x³]
	numerator, err := e.Mean(x3)
	if err != nil {
		return nil, fmt.Errorf("mean of x^3: %w", err)
	}

	// Step 5: Approximate variance σ²
	denom := float64(ct.Size()) * B
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("variance (approx): %w", err)
	}

	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance approx): %w", err)
	}

	// Step 6: Initial approximation of 1/σ using Chebyshev
	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebyshevDegree, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}

	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap (Chebyshev init): %w", err)
		}
	}

	// Step 7: Refine inverse std dev using Newton
	varianceRefined, err := e.Variance(ct)
	if err != nil {
		return nil, fmt.Errorf("variance (refined): %w", err)
	}

	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (variance refined): %w", err)
	}

	invSigma, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	// Step 8: Compute (1/σ)³ = (1/σ) × (1/σ)²
	invSigma2, err := e.Mult(invSigma, invSigma)
	if err != nil {
		return nil, fmt.Errorf("inv σ²: %w", err)
	}

	invSigma3, err := e.Mult(invSigma, invSigma2)
	if err != nil {
		return nil, fmt.Errorf("inv σ³: %w", err)
	}

	// Step 9: Final result: E[x³] × (1/σ³)
	invSigma3Expanded, err := e.extendOneToMulty(invSigma3, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	skewness, err := e.Mult(numerator, invSigma3Expanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return skewness, nil
}

func (e *HEEngine) CoeffVar(ct *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonIter      = 2
		newtonScale     = 1
		bootstrapDepth  = 3
		bootstrapLevel  = 10
	)

	scaleBase := float64(ct.Size()) * B

	// Step 1: Compute mean
	mean, err := meanWithCustomDenom(e, ct, scaleBase)
	if err != nil {
		return nil, fmt.Errorf("mean: %w", err)
	}

	// Step 2: Compute sign(mean) using minimax polynomial
	polys := minimax.NewPolynomial(DefaultPolynomialForSign)
	minimaxEval := minimax.NewEvaluator(e.params, e.evaluator, e.BTS)
	signEvaluator := NewEvaluator(e.params, minimaxEval, polys)

	signCtxt, err := signEvaluator.Sign(mean.Ciphertexts()[0])
	if err != nil {
		return nil, fmt.Errorf("sign(mean): %w", err)
	}

	// Step 3: Replicate sign across all slots
	if signCtxt.Level() < bootstrapLevel && e.IsBTS {
		signCtxt, err = e.BTS.Bootstrap(signCtxt)
		if err != nil {
			return nil, fmt.Errorf("bootstrap sign(mean): %w", err)
		}
	}
	signSlots := make([]*rlwe.Ciphertext, len(mean.Ciphertexts()))
	for i := range signSlots {
		signSlots[i] = signCtxt.CopyNew()
	}
	signMean := NewHEData(signSlots, mean.Size(), signCtxt.Level(), mean.Scale())

	// Step 4: Compute |mean| = mean × sign(mean)
	absMean, err := e.Mult(mean, signMean)
	if err != nil {
		return nil, fmt.Errorf("signedMean: %w", err)
	}
	absMeanCtxt, err := e.selectOneCtxt(absMean)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt(absMean): %w", err)
	}

	// Step 5: Compute 1 / |mean|
	initInv, err := e.ChebyshevInvSqrt(absMeanCtxt, chebyshevDegree, B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt(absMean): %w", err)
	}
	if e.IsBTS {
		initInv, err = e.DoBootstrap(initInv, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap ChebyshevInvSqrt(mean): %w", err)
		}
	}

	squareInv, err := e.Mult(initInv, initInv)
	if err != nil {
		return nil, fmt.Errorf("square(ChebyshevInvSqrt): %w", err)
	}

	invAbsMean, err := e.HENewtonInv(absMeanCtxt, squareInv, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv(mean): %w", err)
	}

	signOne, err := e.selectOneCtxt(signMean)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt(signMean): %w", err)
	}
	invMeanSigned, err := e.Mult(invAbsMean, signOne)
	if err != nil {
		return nil, fmt.Errorf("restore sign to invMean: %w", err)
	}

	// Step 6: Compute variance
	variance, err := varianceWithCustomDenom(e, ct, scaleBase, scaleBase*B)
	if err != nil {
		return nil, fmt.Errorf("variance: %w", err)
	}
	varianceCtxt, err := e.selectOneCtxt(variance)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt(variance): %w", err)
	}

	// Step 7: sqrt(variance)
	sqrtVar, err := e.CryptoSqrt(varianceCtxt, 0, B)
	if err != nil {
		return nil, fmt.Errorf("CryptoSqrt(variance): %w", err)
	}
	if e.IsBTS {
		sqrtVar, err = e.DoBootstrap(sqrtVar, 2)
		if err != nil {
			return nil, fmt.Errorf("bootstrap sqrt(variance): %w", err)
		}
	}

	// Optional: adjust scale
	sqrtVar, err = e.MultConst(sqrtVar, B)
	if err != nil {
		return nil, fmt.Errorf("scale sqrt(variance): %w", err)
	}

	// Step 8: CV = σ / |μ| = sqrt(variance) × (1 / |mean|)
	cv, err := e.Mult(sqrtVar, invMeanSigned)
	if err != nil {
		return nil, fmt.Errorf("final multiply (CV): %w", err)
	}

	cv, err = e.extendOneToMulty(cv, len(ct.ciphertexts), ct.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	return cv, nil
}

func (e *HEEngine) PCorrCoeff(ct1, ct2 *HEData, B float64) (*HEData, error) {
	const (
		chebyshevDegree = 2
		newtonIter      = 6
		newtonScale     = 2
		bootstrapDepth  = 3
	)

	// Step 1: Compute means
	meanX, err := e.Mean(ct1)
	if err != nil {
		return nil, fmt.Errorf("mean of ct1: %w", err)
	}
	meanY, err := e.Mean(ct2)
	if err != nil {
		return nil, fmt.Errorf("mean of ct2: %w", err)
	}

	// Step 2: Centered data
	xCentered, err := e.Sub(ct1, meanX)
	if err != nil {
		return nil, fmt.Errorf("center ct1: %w", err)
	}
	yCentered, err := e.Sub(ct2, meanY)
	if err != nil {
		return nil, fmt.Errorf("center ct2: %w", err)
	}

	// Step 3: Numerator = E[(X - μx)(Y - μy)]
	mulXY, err := e.Mult(xCentered, yCentered)
	if err != nil {
		return nil, fmt.Errorf("x·y: %w", err)
	}
	numerator, err := e.Mean(mulXY)
	if err != nil {
		return nil, fmt.Errorf("mean of x·y: %w", err)
	}

	// Step 4: Compute inverse std for X
	invStdX, err := e.computeInvStd(ct1, chebyshevDegree, newtonIter, newtonScale, bootstrapDepth, B)
	if err != nil {
		return nil, fmt.Errorf("computeInvStd (X): %w", err)
	}

	// Step 5: Compute inverse std for Y
	invStdY, err := e.computeInvStd(ct2, chebyshevDegree, newtonIter, newtonScale, bootstrapDepth, B)
	if err != nil {
		return nil, fmt.Errorf("computeInvStd (Y): %w", err)
	}

	// Step 6: Compute PCC = numerator × (1/σx) × (1/σy)
	denominator, err := e.Mult(invStdX, invStdY)
	if err != nil {
		return nil, fmt.Errorf("σx·σy inverse: %w", err)
	}

	denominatorExpanded, err := e.extendOneToMulty(denominator, len(numerator.Ciphertexts()), numerator.Size())
	if err != nil {
		return nil, fmt.Errorf("extendOneToMulty: %w", err)
	}

	pcc, err := e.Mult(numerator, denominatorExpanded)
	if err != nil {
		return nil, fmt.Errorf("final multiply: %w", err)
	}

	return pcc, nil
}

func (e *HEEngine) computeInvStd(ct *HEData, chebDeg, newtonIter, newtonScale, bootstrapDepth int, B float64) (*HEData, error) {
	denom := float64(ct.Size()) * B

	// Approximate variance
	varianceApprox, err := varianceWithCustomDenom(e, ct, denom, denom*B)
	if err != nil {
		return nil, fmt.Errorf("variance (approx): %w", err)
	}
	varApproxCtxt, err := e.selectOneCtxt(varianceApprox)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (approx): %w", err)
	}

	// Initial guess for 1/σ
	invSigmaInit, err := e.ChebyshevInvSqrt(varApproxCtxt, chebDeg, B*B)
	if err != nil {
		return nil, fmt.Errorf("ChebyshevInvSqrt: %w", err)
	}
	if e.IsBTS {
		invSigmaInit, err = e.DoBootstrap(invSigmaInit, bootstrapDepth)
		if err != nil {
			return nil, fmt.Errorf("bootstrap: %w", err)
		}
	}

	// Refined variance
	varianceRefined, err := e.Variance(ct)
	if err != nil {
		return nil, fmt.Errorf("variance (refined): %w", err)
	}
	varRefinedCtxt, err := e.selectOneCtxt(varianceRefined)
	if err != nil {
		return nil, fmt.Errorf("selectOneCtxt (refined): %w", err)
	}

	// Newton refinement
	invStd, err := e.HENewtonInv(varRefinedCtxt, invSigmaInit, B, newtonIter, newtonScale)
	if err != nil {
		return nil, fmt.Errorf("HENewtonInv: %w", err)
	}

	return invStd, nil
}

func meanWithCustomDenom(e *HEEngine, ct *HEData, denom float64) (*HEData, error) {
	// Step 1: Divide by denominator
	scaledCt, err := e.MultConst(ct, 1.0/denom)
	if err != nil {
		return nil, fmt.Errorf("division by custom denom failed: %w", err)
	}

	// Step 2: Sum all elements
	mean, err := e.Sum(scaledCt)
	if err != nil {
		return nil, fmt.Errorf("sum failed: %w", err)
	}

	return mean, nil
}

func varianceWithCustomDenom(e *HEEngine, ct *HEData, xDenom, xSquareDenom float64) (*HEData, error) {
	// Step 1: Compute E[X]
	meanXScaled, err := e.MultConst(ct, 1.0/xDenom)
	if err != nil {
		return nil, fmt.Errorf("MultConst(1/xDenom): %w", err)
	}

	meanX, err := e.Sum(meanXScaled)
	if err != nil {
		return nil, fmt.Errorf("Sum(E[x]): %w", err)
	}

	// Step 2: Compute (E[X])^2
	squaredMeanX, err := e.Mult(meanX, meanX)
	if err != nil {
		return nil, fmt.Errorf("square of E[x]: %w", err)
	}

	// Step 3: Compute X^2
	ctSquared, err := e.Mult(ct, ct)
	if err != nil {
		return nil, fmt.Errorf("X^2: %w", err)
	}

	// Step 4: Compute E[X^2]
	meanXSquaredScaled, err := e.MultConst(ctSquared, 1.0/xSquareDenom)
	if err != nil {
		return nil, fmt.Errorf("MultConst(1/xSquareDenom): %w", err)
	}

	meanXSquared, err := e.Sum(meanXSquaredScaled)
	if err != nil {
		return nil, fmt.Errorf("Sum(E[x^2]): %w", err)
	}

	// Step 5: Return variance = E[X^2] - (E[X])^2
	variance, err := e.Sub(meanXSquared, squaredMeanX)
	if err != nil {
		return nil, fmt.Errorf("E[x^2] - E[x]^2: %w", err)
	}

	return variance, nil
}

func (e *HEEngine) selectOneCtxt(ct *HEData) (*HEData, error) {
	size := ct.Size()
	if size > e.params.MaxSlots() {
		size = e.params.MaxSlots()
	}
	ctxt := make([]*rlwe.Ciphertext, 1)
	ctxt[0] = ct.Ciphertexts()[0].CopyNew()
	return NewHEData(ctxt, size, ct.Level(), ct.Scale()), nil
}

func (e *HEEngine) extendOneToMulty(ct *HEData, num, size int) (*HEData, error) {
	ctxts := make([]*rlwe.Ciphertext, num)
	for i := 0; i < num; i++ {
		ctxts[i] = ct.Ciphertexts()[0].CopyNew()
	}
	return NewHEData(ctxts, size, ctxts[0].Level(), ct.Scale()), nil
}
