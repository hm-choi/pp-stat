package engine

import (
	"fmt"
	"math"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
)

func (e *HEEngine) ChebyshevInvSqrt(ct *HEData, mode int, B float64) (*HEData, error) {
	cpData := ct.CopyData()
	d := 9.0

	var F func(float64) float64

	switch mode {
	case 1:
		cpData, _ = e.MultConst(cpData, 2.0/B)
		F = func(x float64) (y float64) {
			if x > -1.0 {
				return 1 / math.Sqrt(B/2) / (math.Sqrt(x + 1.0))
			} else {
				return 0
			}
		}
	case 2:
		F = func(x float64) (y float64) {
			if x > -1.0 {
				return 1 / math.Sqrt(B) / (math.Sqrt(x + 1.0))
			} else {
				return 0
			}
		}
	case 0:
		F = func(x float64) float64 {
			if x > -1.0 {
				return 1 / math.Sqrt(x+1.0)
			}
			return 0
		}
	default:
		return nil, fmt.Errorf("invalid InvSqrt mode: %d", mode)

	}

	scaled_ct, err := e.SubConst(cpData, 1)
	if err != nil {
		return nil, err
	}
	gcbsp := GetChebyshevPoly(1.0, int(math.Pow(2, float64(d))-2), F)
	poly := polynomial.NewPolynomial(gcbsp)
	polyEval := polynomial.NewEvaluator(e.params, e.Evaluator())

	scaledCtxts := scaled_ct.Ciphertexts()
	invCtxts := []*rlwe.Ciphertext{}
	targetScale := e.params.DefaultScale().Div(rlwe.NewScale(2))
	for i := 0; i < len(scaledCtxts); i++ {
		// tmpList1 := make([]float64, e.params.MaxSlots())
		// tmpList2 := make([]float64, e.params.MaxSlots())
		// if scaled_ct.Size() < e.params.MaxSlots() {
		// 	for j := scaled_ct.Size(); j < e.params.MaxSlots(); j++ {
		// 		tmpList1[j] = 1
		// 		tmpList2[j] = 1 / math.Sqrt(B)
		// 	}
		// 	e.evaluator.Add(scaledCtxts[i], tmpList1, scaledCtxts[i])
		// }
		p2, _ := polyEval.Evaluate(scaledCtxts[i], poly, targetScale)
		p2.Scale = p2.Scale.Mul(rlwe.NewScale(2))
		conj, _ := e.evaluator.ConjugateNew(p2)
		e.evaluator.Add(p2, conj, p2)
		p2.Scale = scaledCtxts[i].Scale
		invCtxts = append(invCtxts, p2)
		// if scaled_ct.Size() < e.params.MaxSlots() {
		// 	e.evaluator.Sub(scaledCtxts[i], tmpList1, scaledCtxts[i])
		// 	e.evaluator.Sub(invCtxts[i], tmpList2, invCtxts[i])
		// }
	}
	return NewHEData(invCtxts, ct.Size(), invCtxts[0].Level(), ct.Scale()), nil
}

func (e *HEEngine) HENewtonInv(ct, init *HEData, B float64, iter, mode int) (*HEData, error) {
	N := 1.0
	x, y := ct.CopyData(), init.CopyData()
	switch mode {
	case 1:
		x, _ = e.MultConst(x, B)
	case 2:
		N = 2
		x, _ = e.MultConst(x, 1.0/N)
	case 3:
		N = 2
		x, _ = e.MultConst(x, B/N)
	}
	for _ = range iter {
		if e.IsBTS {
			y, _ = e.DoBootstrap(y, 3)
		}

		tmp_a, _ := e.MultConst(y, float64((N+1))/float64(N))
		tmp_b, _ := e.Mult(x, y)

		if N == 2.0 {
			y, _ = e.Mult(y, y)
		}
		tmp_b, _ = e.Mult(tmp_b, y)
		y, _ = e.Sub(tmp_a, tmp_b)
	}
	return y, nil
}

func (e *HEEngine) CryptoInvSqrt(ct *HEData) (*HEData, error) {
	//	B := math.Pow(2.0, 4)
	y, err := e.ChebyshevInvSqrt(ct, 1, B)
	if err != nil {
		return y, err
	}
	if e.IsBTS {
		y, err = e.DoBootstrap(y, 3)
	}
	return e.HENewtonInv(ct, y, B, 6, 2)
}

func (e *HEEngine) CryptoInv(ct *HEData) (*HEData, error) {
	y0, err := e.ChebyshevInvSqrt(ct, 2, 1.0)
	if err != nil {
		return y0, err
	}

	y0, err = e.Mult(y0, y0)
	return e.HENewtonInv(ct, y0, 1.0, 4, 1)

}

func (e *HEEngine) CryptoSqrt(ct *HEData, types int, B float64) (*HEData, error) {
	d := 9
	F3 := func(x float64) (y float64) {
		if x > -1.0 {
			return math.Sqrt(x + 1) //* math.Sqrt(B/2.0)
		} else {
			return 0
		}
	}

	x := ct.CopyData()
	if types == 1 {
		x, _ = e.MultConst(x, 2.0/B)
	} else if types == 2 {
		F3 = func(x float64) (y float64) {
			if x > -1.0 {
				return math.Sqrt(B) * math.Sqrt(x+1) //* math.Sqrt(B/2.0)
			} else {
				return 0
			}
		}
	}

	scaled_ct, _ := e.SubConst(x, 1)
	gcbsp := GetChebyshevPoly(1.0, int(math.Pow(2, float64(d))-2), F3)
	poly := polynomial.NewPolynomial(gcbsp)
	polyEval := polynomial.NewEvaluator(e.params, e.evaluator)

	ctxts := make([]*rlwe.Ciphertext, len(scaled_ct.Ciphertexts()))
	for i := 0; i < len(ctxts); i++ {
		p2, _ := polyEval.Evaluate(scaled_ct.Ciphertexts()[i], poly, e.params.DefaultScale().Div(rlwe.NewScale(2)))
		p2.Scale = p2.Scale.Mul(rlwe.NewScale(2))
		conj, _ := e.evaluator.ConjugateNew(p2)
		e.evaluator.Add(p2, conj, p2)
		p2.Scale = scaled_ct.Ciphertexts()[i].Scale
		ctxts[i] = p2
	}

	return NewHEData(ctxts, ct.Size(), ctxts[0].Level(), ct.Scale()), nil
}
