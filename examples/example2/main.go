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
	DATA_SIZE := 1000000
	RANGE, RANGE2 := 20.0, 100.0
	EVAL_NUM := 10
	ZSCORE_MRE, ZSCORE_MAE, ZSCORE_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	KURT_MRE, KURT_MAE, KURT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	SKEW_MRE, SKEW_MAE, SKEW_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CV_MRE, CV_MAE, CV_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CORR_MRE, CORR_MAE, CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	file, err := os.Create("output.txt")
	if err != nil {
		return
	}
	defer file.Close()
	for i := 0; i < int(EVAL_NUM); i++ {

		values1 := make([]float64, DATA_SIZE)
		values2 := make([]float64, DATA_SIZE)
		values3 := make([]float64, DATA_SIZE)
		for i := 0; i < DATA_SIZE; i++ {
			values1[i] = RANGE * (rand.Float64())
			values2[i] = RANGE * (rand.Float64())
			values3[i] = RANGE2 * (rand.Float64())
		}

		fmt.Println("VAR, MEAN", utils.Variance(values1)/(128), utils.Mean(values1))
		ctxt1, _ := engine.Encrypt(values1)
		ctxt2, _ := engine.Encrypt(values2)
		ctxt3, _ := engine.Encrypt(values3)

		fmt.Println("[Z-Score Normalization]")
		TIME := time.Now()
		zNormReal := utils.ZScoreNorm(values3)
		zNorm, _ := engine.ZScoreNorm(ctxt3)
		ZNORM_TIME := time.Since(TIME)
		zNormResult, _ := engine.Decrypt(zNorm)

		_, mre := utils.CheckMRE(zNormResult, zNormReal, zNormReal, len(zNormReal))
		_, mae := utils.CheckMAE(zNormResult, zNormReal, zNormReal, len(zNormReal))

		fmt.Println("ZNormResult (mae, mre)", mae, mre, ZNORM_TIME)
		ZSCORE_MRE[i] = mre
		ZSCORE_MAE[i] = mae
		ZSCORE_TIME[i] = float64(ZNORM_TIME.Seconds())

		fmt.Println("[Skewness]")
		_, _, skewReal := utils.Skewness(values1)
		TIME = time.Now()
		skew, _ := engine.Skewness(ctxt1)
		SK_TIME := time.Since(TIME)
		skewResult, _ := engine.Decrypt(skew)
		fmt.Println("SkewResult", skewResult[0], skewReal, math.Abs(skewResult[0]-skewReal), math.Abs(skewResult[0]-skewReal)/math.Abs(skewReal), SK_TIME)

		SKEW_MRE[i] = math.Abs(skewResult[0]-skewReal) / math.Abs(skewReal)
		SKEW_MAE[i] = math.Abs(skewResult[0] - skewReal)
		SKEW_TIME[i] = float64(SK_TIME.Seconds())

		fmt.Println("[Kutosis]")
		_, _, kurtReal := utils.Kurtosis(values1)
		TIME = time.Now()
		kurt, _ := engine.Kurtosis(ctxt1)
		KT_TIME := time.Since(TIME)
		kurtResult, _ := engine.Decrypt(kurt)
		fmt.Println("KurtResult", kurtResult[0], kurtReal, math.Abs(kurtResult[0]-kurtReal), math.Abs(kurtResult[0]-kurtReal)/math.Abs(kurtReal), KT_TIME)

		KURT_MRE[i] = math.Abs(kurtResult[0]-kurtReal) / math.Abs(kurtReal)
		KURT_MAE[i] = math.Abs(kurtResult[0] - kurtReal)
		KURT_TIME[i] = float64(KT_TIME.Seconds())

		fmt.Println("[CoeffVar]")
		_, _, cvReal := utils.CoeffVar(values1)
		TIME = time.Now()
		cv, _ := engine.CoeffVar(ctxt1)
		CVV_TIME := time.Since(TIME)
		cvResult, _ := engine.Decrypt(cv)
		fmt.Println("CVResult", cvResult[0], cvReal, math.Abs(cvResult[0]-cvReal), math.Abs(cvResult[0]-cvReal)/math.Abs(cvReal), CVV_TIME)
		CV_MRE[i] = math.Abs(cvResult[0]-cvReal) / math.Abs(cvReal)
		CV_MAE[i] = math.Abs(cvResult[0] - cvReal)
		CV_TIME[i] = float64(CVV_TIME.Seconds())

		fmt.Println("[Correlation]")
		_, corrReal, _ := utils.Correlation(values1, values2)
		TIME = time.Now()
		corr, _ := engine.PCorrCoeff(ctxt1, ctxt2)
		COR_TIME := time.Since(TIME)
		corrResult, _ := engine.Decrypt(corr)
		fmt.Println("CorrResult", corrResult[0], corrReal, math.Abs(corrResult[0]-corrReal), math.Abs(corrResult[0]-corrReal)/math.Abs(corrReal), COR_TIME)
		CORR_MRE[i] = math.Abs(corrResult[0]-corrReal) / math.Abs(corrReal)
		CORR_MAE[i] = math.Abs(corrResult[0] - corrReal)
		CORR_TIME[i] = float64(COR_TIME.Seconds())

	}

	result := fmt.Sprintf("[ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(ZSCORE_MAE), utils.StdDev(ZSCORE_MAE), utils.Mean(ZSCORE_MRE), utils.StdDev(ZSCORE_MRE), utils.Mean(ZSCORE_TIME), utils.StdDev(ZSCORE_TIME))
	result += fmt.Sprintf("[SKEWNE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(SKEW_MAE), utils.StdDev(SKEW_MAE), utils.Mean(SKEW_MRE), utils.StdDev(SKEW_MRE), utils.Mean(SKEW_TIME), utils.StdDev(SKEW_TIME))
	result += fmt.Sprintf("[KURTOS] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(KURT_MAE), utils.StdDev(KURT_MAE), utils.Mean(KURT_MRE), utils.StdDev(KURT_MRE), utils.Mean(KURT_TIME), utils.StdDev(KURT_TIME))
	result += fmt.Sprintf("[COEFFV] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CV_MAE), utils.StdDev(CV_MAE), utils.Mean(CV_MRE), utils.StdDev(CV_MRE), utils.Mean(CV_TIME), utils.StdDev(CV_TIME))
	result += fmt.Sprintf("[CORREL] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CORR_MAE), utils.StdDev(CORR_MAE), utils.Mean(CORR_MRE), utils.StdDev(CORR_MRE), utils.Mean(CORR_TIME), utils.StdDev(CORR_TIME))
	io.WriteString(file, result)
}
