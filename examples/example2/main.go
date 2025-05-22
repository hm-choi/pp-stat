package main

import (
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"time"

	"github.com/hm-choi/pp-stat/config"
	"github.com/hm-choi/pp-stat/engine"
	"github.com/hm-choi/pp-stat/utils"
)

func main() {
	engine := engine.NewHEEngine(config.NewParameters(16, 11, 40, true))

	// Benchmark settings
	DATA_SIZE := 1000000 // Datasize
	RANGE, RANGE2 := 20.0, 100.0
	EVAL_NUM := 10

	// Result arrays (MRE and TIME only)
	ZSCORE_MRE, ZSCORE_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	KURT_MRE, KURT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	SKEW_MRE, SKEW_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CV_MRE, CV_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CORR_MRE, CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	file, err := os.Create("output.txt")
	if err != nil {
		fmt.Println("Failed to create file:", err)
		return
	}
	defer file.Close()

	for i := 0; i < EVAL_NUM; i++ {
		// Generate random plaintext data
		values1 := make([]float64, DATA_SIZE)
		values2 := make([]float64, DATA_SIZE)
		values3 := make([]float64, DATA_SIZE)
		for j := 0; j < DATA_SIZE; j++ {
			values1[j] = RANGE * rand.Float64()
			values2[j] = RANGE * rand.Float64()
			values3[j] = RANGE2 * rand.Float64()
		}

		// Encrypt inputs
		ctxt1, _ := engine.Encrypt(values1)
		ctxt2, _ := engine.Encrypt(values2)
		ctxt3, _ := engine.Encrypt(values3)

		// [Z-Score Normalization]
		fmt.Println("[Z-Score Normalization]")
		start := time.Now()
		zNormReal := utils.ZScoreNorm(values3)
		zNorm, _ := engine.ZScoreNorm(ctxt3)
		duration := time.Since(start)
		zNormResult, _ := engine.Decrypt(zNorm)
		_, mre := utils.CheckMRE(zNormResult, zNormReal, zNormReal, len(zNormReal))
		fmt.Println("Z-Score MRE:", mre, duration)
		ZSCORE_MRE[i] = mre
		ZSCORE_TIME[i] = duration.Seconds()

		// [Skewness]
		fmt.Println("[Skewness]")
		_, _, skewReal := utils.Skewness(values1)
		start = time.Now()
		skew, _ := engine.Skewness(ctxt1)
		duration = time.Since(start)
		skewResult, _ := engine.Decrypt(skew)
		mre = math.Abs(skewResult[0]-skewReal) / math.Abs(skewReal)
		fmt.Println("Skewness MRE:", mre, duration)
		SKEW_MRE[i] = mre
		SKEW_TIME[i] = duration.Seconds()

		// [Kurtosis]
		fmt.Println("[Kurtosis]")
		_, _, kurtReal := utils.Kurtosis(values1)
		start = time.Now()
		kurt, _ := engine.Kurtosis(ctxt1)
		duration = time.Since(start)
		kurtResult, _ := engine.Decrypt(kurt)
		mre = math.Abs(kurtResult[0]-kurtReal) / math.Abs(kurtReal)
		fmt.Println("Kurtosis MRE:", mre, duration)
		KURT_MRE[i] = mre
		KURT_TIME[i] = duration.Seconds()

		// [Coefficient of Variation]
		fmt.Println("[CoeffVar]")
		_, _, cvReal := utils.CoeffVar(values1)
		start = time.Now()
		cv, _ := engine.CoeffVar(ctxt1)
		duration = time.Since(start)
		cvResult, _ := engine.Decrypt(cv)
		mre = math.Abs(cvResult[0]-cvReal) / math.Abs(cvReal)
		fmt.Println("CoeffVar MRE:", mre, duration)
		CV_MRE[i] = mre
		CV_TIME[i] = duration.Seconds()

		// [Correlation]
		fmt.Println("[Correlation]")
		_, corrReal, _ := utils.Correlation(values1, values2)
		start = time.Now()
		corr, _ := engine.PCorrCoeff(ctxt1, ctxt2)
		duration = time.Since(start)
		corrResult, _ := engine.Decrypt(corr)
		mre = math.Abs(corrResult[0]-corrReal) / math.Abs(corrReal)
		fmt.Println("Correlation MRE:", mre, duration)
		CORR_MRE[i] = mre
		CORR_TIME[i] = duration.Seconds()
	}

	// Write summary results to output file
	result := fmt.Sprintf("[ZSCORE] MRE %e (%e), TIME %f (%f)\n", utils.Mean(ZSCORE_MRE), utils.StdDev(ZSCORE_MRE), utils.Mean(ZSCORE_TIME), utils.StdDev(ZSCORE_TIME))
	result += fmt.Sprintf("[SKEWNESS] MRE %e (%e), TIME %f (%f)\n", utils.Mean(SKEW_MRE), utils.StdDev(SKEW_MRE), utils.Mean(SKEW_TIME), utils.StdDev(SKEW_TIME))
	result += fmt.Sprintf("[KURTOSIS] MRE %e (%e), TIME %f (%f)\n", utils.Mean(KURT_MRE), utils.StdDev(KURT_MRE), utils.Mean(KURT_TIME), utils.StdDev(KURT_TIME))
	result += fmt.Sprintf("[COEFFVAR] MRE %e (%e), TIME %f (%f)\n", utils.Mean(CV_MRE), utils.StdDev(CV_MRE), utils.Mean(CV_TIME), utils.StdDev(CV_TIME))
	result += fmt.Sprintf("[CORREL] MRE %e (%e), TIME %f (%f)\n", utils.Mean(CORR_MRE), utils.StdDev(CORR_MRE), utils.Mean(CORR_TIME), utils.StdDev(CORR_TIME))
	io.WriteString(file, result)
}
