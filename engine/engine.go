package engine

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/ring"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils"
)

type HEEngine struct {
	params    ckks.Parameters
	Sk        *rlwe.SecretKey
	Pk        *rlwe.PublicKey
	Rlk       *rlwe.RelinearizationKey
	Evk       *bootstrapping.EvaluationKeys
	Encryptor *rlwe.Encryptor
	Decryptor *rlwe.Decryptor
	evaluator *ckks.Evaluator
	Encoder   *ckks.Encoder
	BTS       *bootstrapping.Evaluator
	Slots     int
	IsBTS     bool
}

func (e *HEEngine) Evaluator() *ckks.Evaluator { return e.evaluator }
func (e *HEEngine) Params() ckks.Parameters    { return e.params }

func GetParam(LogN int, LEVEL int, SCALE int) ckks.Parameters {
	LogQ := make([]int, LEVEL+1)
	LogQ[0] = 60
	for i := range LEVEL {
		LogQ[i+1] = SCALE
	}

	params, _ := ckks.NewParametersFromLiteral(
		ckks.ParametersLiteral{
			LogN:            LogN,                  // log2(ring degree)
			LogQ:            LogQ,                  // log2(primes Q) (ciphertext modulus)
			LogP:            []int{61, 61, 61, 61}, // log2(primes P) (auxiliary modulus)
			LogDefaultScale: SCALE,                 // log2(scale)
		})
	return params
}

func GetBSParam(LogN int, LEVEL int, SCALE int) (ckks.Parameters, bootstrapping.Parameters) {
	logQ := make([]int, LEVEL+1)
	logQ[0] = 60
	for idx := range LEVEL {
		logQ[idx+1] = SCALE
	}
	params, err := ckks.NewParametersFromLiteral(ckks.ParametersLiteral{
		LogN:            LogN,          // Log2 of the ring degree
		LogQ:            logQ,          // Log2 of the ciphertext prime moduli
		LogP:            []int{61, 61}, // Log2 of the key-switch auxiliary prime moduli
		LogDefaultScale: 40,            // Log2 of the scale
		Xs:              ring.Ternary{H: 192},
	})

	if err != nil {
		panic(err)
	}

	btpParametersLit := bootstrapping.ParametersLiteral{
		LogN: utils.Pointy(LogN),
		LogP: []int{61, 61},
		Xs:   params.Xs(),
	}

	btpParams, err := bootstrapping.NewParametersFromLiteral(params, btpParametersLit)
	if err != nil {
		panic(err)
	}

	return params, btpParams
}

func NewHEEngine(isBTS bool, params ckks.Parameters, btpParams bootstrapping.Parameters) *HEEngine {
	kgen := rlwe.NewKeyGenerator(params)
	sk, pk := kgen.GenKeyPairNew()
	rlk := kgen.GenRelinearizationKeyNew(sk)

	galEls := []uint64{params.GaloisElementForComplexConjugation()}
	for rot := 1; rot < params.MaxSlots(); rot *= 2 {
		galEls = append(galEls, params.GaloisElement(rot))
	}

	ecd := ckks.NewEncoder(params)
	dec := rlwe.NewDecryptor(params, sk)
	enc := rlwe.NewEncryptor(params, pk)

	var eval *ckks.Evaluator

	var bts *bootstrapping.Evaluator

	if isBTS {
		btsEvk, _, _ := btpParams.GenEvaluationKeys(sk)
		eval = ckks.NewEvaluator(params, btsEvk)
		eval = eval.WithKey(rlwe.NewMemEvaluationKeySet(rlk, kgen.GenGaloisKeysNew(galEls, sk)...))
		bts, _ = bootstrapping.NewEvaluator(btpParams, btsEvk)
	} else {
		evk := rlwe.NewMemEvaluationKeySet(rlk)
		eval = ckks.NewEvaluator(params, evk)
		eval = eval.WithKey(rlwe.NewMemEvaluationKeySet(rlk, kgen.GenGaloisKeysNew(galEls, sk)...))
	}

	return &HEEngine{
		params:    params,
		Sk:        sk,
		Pk:        pk,
		Rlk:       rlk,
		Evk:       nil,
		Encryptor: enc,
		Decryptor: dec,
		evaluator: eval,
		Encoder:   ecd,
		BTS:       bts,
		Slots:     params.MaxSlots(),
		IsBTS:     isBTS,
	}
}

func (e *HEEngine) Encrypt(input []float64) (ctxt *HEData, err error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("input data size is zero: %w", err)
	}

	dataSize := len(input)

	ctxtNum := ((len(input) + e.Slots - 1) / e.Slots)

	ciphertexts := make([]*rlwe.Ciphertext, ctxtNum)
	for i := range ctxtNum {
		start := i * e.Slots
		end := start + e.Slots
		if end > len(input) {
			end = len(input)
		}

		pt := ckks.NewPlaintext(e.params, e.params.MaxLevel())
		if err = e.Encoder.Encode(input[start:end], pt); err != nil {
			return nil, fmt.Errorf("encoding failed: %w", err)
		}
		ctxt, err := e.Encryptor.EncryptNew(pt)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		ciphertexts[i] = ctxt
	}
	heData := NewHEData(ciphertexts, dataSize, e.params.MaxLevel(), 60.0)
	return heData, nil
}

func (e *HEEngine) Decrypt(ctxt *HEData) (output []float64, err error) {
	output = []float64{}
	ctxts := ctxt.Ciphertexts()
	for i := range len(ctxts) {
		tmpSlice := make([]float64, e.params.MaxSlots())
		if err = e.Encoder.Decode(e.Decryptor.DecryptNew(ctxts[i]), tmpSlice); err != nil {
			return nil, fmt.Errorf("decoding failed: %w", err)
		}
		output = append(output, tmpSlice...)
	}
	return output[:ctxt.Size()], nil
}

func (e *HEEngine) DecryptComplex(ctxt *HEData) (output []complex128, err error) {
	output = []complex128{}
	ctxts := ctxt.Ciphertexts()
	for i := range len(ctxts) {
		tmpSlice := make([]complex128, e.params.MaxSlots())
		if err = e.Encoder.Decode(e.Decryptor.DecryptNew(ctxts[i]), tmpSlice); err != nil {
			return nil, fmt.Errorf("decoding failed: %w", err)
		}
		output = append(output, tmpSlice...)
	}
	return output, nil
}
func (e *HEEngine) DoBootstrap(ctxt *HEData, level int) (*HEData, error) {
	if !e.IsBTS {
		return nil, fmt.Errorf("The parameter does not support bootstrapping.")
	}
	if ctxt.Ciphertexts()[0].Level() < level {
		ctxtNum := len(ctxt.Ciphertexts())
		btsCtxts := make([]*rlwe.Ciphertext, ctxtNum)
		for i := 0; i < ctxtNum; i++ {
			ct := ctxt.Ciphertexts()[i].CopyNew()
			ct.Scale = e.params.DefaultScale().Mul(rlwe.NewScale(2))
			conj, _ := e.evaluator.ConjugateNew(ct)
			ct, _ = e.evaluator.AddNew(conj, ct)
			ct, _ = e.BTS.Bootstrap(ct)
			btsCtxts[i] = ct
		}
		return NewHEData(btsCtxts, ctxt.Size(), btsCtxts[0].Level(), ctxt.Scale()), nil
	} else {
		return ctxt, nil
	}
}
