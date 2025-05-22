package config

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/circuits/ckks/bootstrapping"
	"github.com/tuneinsight/lattigo/v6/schemes/ckks"
	"github.com/tuneinsight/lattigo/v6/utils"
)

type Parameters struct {
	LogN  int
	Level int
	Scale float64
}

func NewParameters(LogN, Level, Scale int, isBTS bool) (bool, ckks.Parameters, bootstrapping.Parameters) {
	LogQ := make([]int, Level+1)
	LogQ[0] = 60
	for i := range Level {
		LogQ[i+1] = Scale
	}
	params, err := ckks.NewParametersFromLiteral(
		ckks.ParametersLiteral{
			LogN:            LogN,                  // log2(ring degree)
			LogQ:            LogQ,                  // log2(primes Q) (ciphertext modulus)
			LogP:            []int{61, 61, 61, 61}, // log2(primes P) (auxiliary modulus)
			LogDefaultScale: Scale,                 // log2(scale)
		})
	if err != nil {
		panic(err)
	}

	var btpParams bootstrapping.Parameters

	if isBTS {
		btpParametersLit := bootstrapping.ParametersLiteral{
			LogN: utils.Pointy(LogN),
			LogP: []int{61, 61},
			Xs:   params.Xs(),
		}
		btpParams, err = bootstrapping.NewParametersFromLiteral(params, btpParametersLit)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Residual parameters: logN=%d, logSlots=%d, H=%d, sigma=%f, logQP=%f, levels=%d, scale=2^%d\n",
			btpParams.ResidualParameters.LogN(),
			btpParams.ResidualParameters.LogMaxSlots(),
			btpParams.ResidualParameters.XsHammingWeight(),
			btpParams.ResidualParameters.Xe(), params.LogQP(),
			btpParams.ResidualParameters.MaxLevel(),
			btpParams.ResidualParameters.LogDefaultScale())

		fmt.Printf("Bootstrapping parameters: logN=%d, logSlots=%d, H(%d; %d), sigma=%f, logQP=%f, levels=%d, scale=2^%d\n",
			btpParams.BootstrappingParameters.LogN(),
			btpParams.BootstrappingParameters.LogMaxSlots(),
			btpParams.BootstrappingParameters.XsHammingWeight(),
			btpParams.EphemeralSecretWeight,
			btpParams.BootstrappingParameters.Xe(),
			btpParams.BootstrappingParameters.LogQP(),
			btpParams.BootstrappingParameters.QCount(),
			btpParams.BootstrappingParameters.LogDefaultScale())

		return isBTS, params, btpParams
	} else {
		return isBTS, params, btpParams
	}
}
