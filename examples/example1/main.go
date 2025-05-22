package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/hm-choi/pp-stat/config"
	"github.com/hm-choi/pp-stat/engine"
	invsqrt "github.com/hm-choi/pp-stat/examples/inv_sqrt"
	"github.com/hm-choi/pp-stat/utils"
)

func main() {
	engine := engine.NewHEEngine(config.NewParameters(16, 11, 40, true))

	const (
		DATA_SIZE = 32768 // Slot size
		B         = 100.0
		START     = 0.001
		MIDDLE    = 1.0
		STOP      = B
		EVAL_NUM  = 10
	)

	// Generate input test values and ground truth (1/sqrt(x))
	test1 := utils.Linspace(START, MIDDLE, DATA_SIZE/2)
	test2 := utils.Linspace(MIDDLE, STOP, DATA_SIZE/2)
	test := append(test1, test2...)

	invS := make([]float64, DATA_SIZE)
	for i, v := range test {
		invS[i] = 1.0 / math.Sqrt(v)
	}

	// Create output file
	file, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("Failed to create output file:", err)
		return
	}
	defer file.Close()

	// Allocate result slices
	PIVOT_MRE, PIVOT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HSTAT_MRE, HSTAT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	INVSQRT_MRE, INVSQRT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	for i := 0; i < EVAL_NUM; i++ {
		// Encrypt input
		ct, _ := engine.Encrypt(test)

		// [1] Pivot-Tangent method
		start := time.Now()
		pivotTangent, _ := invsqrt.PivotTangent(engine, ct, 12, 7)
		elapsed := time.Since(start)
		ptResult, _ := engine.Decrypt(pivotTangent)
		_, ptMRE := utils.CheckMRE(invS, invS, ptResult, ct.Size())
		PIVOT_MRE[i], PIVOT_TIME[i] = ptMRE, elapsed.Seconds()
		fmt.Println("Pivot-Tangent", ptResult[:3], ptMRE, elapsed)

		// [2] HEaaN-Stat method
		start = time.Now()
		heStat, _ := invsqrt.HEStat(engine, ct, 21, B)
		elapsed = time.Since(start)
		hsResult, _ := engine.Decrypt(heStat)
		_, hsMRE := utils.CheckMRE(invS, invS, hsResult, ct.Size())
		HSTAT_MRE[i], HSTAT_TIME[i] = hsMRE, elapsed.Seconds()
		fmt.Println("HEaaN-Stat", hsResult[:3], hsMRE, elapsed)

		// [3] CryptoInvSqrt method
		start = time.Now()
		cryptoInvSqrt, _ := engine.CryptoInvSqrt(ct)
		elapsed = time.Since(start)
		cisResult, _ := engine.Decrypt(cryptoInvSqrt)
		_, cisMRE := utils.CheckMRE(invS, invS, cisResult, ct.Size())
		INVSQRT_MRE[i], INVSQRT_TIME[i] = cisMRE, elapsed.Seconds()
		fmt.Println("CryptoInvSqrt", cisResult[:3], cisMRE, elapsed)
	}

	// Summarize results: mean (stddev) for MAE, MRE, and TIME
	result := fmt.Sprintf("[PIVOT] MRE %e (%e), TIME %.3f (%.3f)\n",
		utils.Mean(PIVOT_MRE), utils.StdDev(PIVOT_MRE),
		utils.Mean(PIVOT_TIME), utils.StdDev(PIVOT_TIME),
	)
	result += fmt.Sprintf("[HSTAT] MRE %e (%e), TIME %.3f (%.3f)\n",
		utils.Mean(HSTAT_MRE), utils.StdDev(HSTAT_MRE),
		utils.Mean(HSTAT_TIME), utils.StdDev(HSTAT_TIME),
	)
	result += fmt.Sprintf("[CISRT] MRE %e (%e), TIME %.3f (%.3f)\n",
		utils.Mean(INVSQRT_MRE), utils.StdDev(INVSQRT_MRE),
		utils.Mean(INVSQRT_TIME), utils.StdDev(INVSQRT_TIME),
	)

	// Write result to file
	io.WriteString(file, result)
}
