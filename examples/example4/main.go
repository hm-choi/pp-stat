package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/hm-choi/pp-stat/config"
	"github.com/hm-choi/pp-stat/engine"
	"github.com/hm-choi/pp-stat/utils"
)

func main() {
	engine := engine.NewHEEngine(config.NewParameters(16, 11, 40, true))
	ageSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 0)
	bmiSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 2)
	smokerSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 4)
	chargeSlice, _ := utils.ReadCSV("../../examples/dataset/insurance.csv", 6)

	for i := 0; i < len(chargeSlice); i++ {
		chargeSlice[i] = chargeSlice[i] / 1000.0
	}

	EVAL_NUM := 10
	CG_ZSCORE_MRE, CG_ZSCORE_MAE, CG_ZSCORE_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_KURT_MRE, CG_KURT_MAE, CG_KURT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_SKEW_MRE, CG_SKEW_MAE, CG_SKEW_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	CG_CV_MRE, CG_CV_MAE, CG_CV_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AGE_CG_CORR_MRE, AGE_CG_CORR_MAE, AGE_CG_CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	BMI_CG_CORR_MRE, BMI_CG_CORR_MAE, BMI_CG_CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	SMOKER_CG_CORR_MRE, SMOKER_CG_CORR_MAE, SMOKER_CG_CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	file, err := os.Create("output.txt")
	if err != nil {
		return
	}
	defer file.Close()

	age, _ := engine.Encrypt(ageSlice)
	bmi, _ := engine.Encrypt(bmiSlice)
	smoker, _ := engine.Encrypt(smokerSlice)
	charge, _ := engine.Encrypt(chargeSlice)

	_, _, skew_cg := utils.Skewness(chargeSlice)
	_, _, kurt_cg := utils.Kurtosis(chargeSlice)
	_, _, cv_cg := utils.CoeffVar(chargeSlice)

	fmt.Println("==============================")
	TIME := time.Now()

	for i := 0; i < int(EVAL_NUM); i++ {
		TIME = time.Now()
		zNormCharge, _ := engine.ZScoreNorm(charge)
		CHARGE_ZNORM_TIME := time.Since(TIME)
		zSNcharge, _ := engine.Decrypt(zNormCharge)

		_, zScoreMreCharge := utils.CheckMRE(utils.ZScoreNorm(chargeSlice), zSNcharge, zSNcharge, len(chargeSlice))
		_, zScoreMaeCharge := utils.CheckMAE(utils.ZScoreNorm(chargeSlice), zSNcharge, zSNcharge, len(chargeSlice))

		CG_ZSCORE_MRE[i] = zScoreMreCharge
		CG_ZSCORE_MAE[i] = zScoreMaeCharge
		CG_ZSCORE_TIME[i] = float64(CHARGE_ZNORM_TIME.Seconds())
		fmt.Println("Charge ZNorm", zScoreMaeCharge, zScoreMreCharge, CHARGE_ZNORM_TIME)

		TIME = time.Now()
		skewCharge, _ := engine.Skewness(charge)
		CHARGE_SKEW_TIME := time.Since(TIME)
		skCharge, _ := engine.Decrypt(skewCharge)
		CG_SKEW_MRE[i] = math.Abs(skCharge[0]-skew_cg) / math.Abs(skew_cg)
		CG_SKEW_MAE[i] = math.Abs(skCharge[0] - skew_cg)
		CG_SKEW_TIME[i] = float64(CHARGE_SKEW_TIME.Seconds())
		fmt.Println("Charge skewResult", skCharge[0], math.Abs(skCharge[0]-skew_cg), math.Abs(skCharge[0]-skew_cg)/math.Abs(skew_cg), CHARGE_SKEW_TIME)

		TIME = time.Now()
		kurtCharge, _ := engine.Kurtosis(charge)
		CHARGE_KURT_TIME := time.Since(TIME)
		ktCharge, _ := engine.Decrypt(kurtCharge)
		CG_KURT_MRE[i] = math.Abs(ktCharge[0]-kurt_cg) / math.Abs(kurt_cg)
		CG_KURT_MAE[i] = math.Abs(ktCharge[0] - kurt_cg)
		CG_KURT_TIME[i] = float64(CHARGE_KURT_TIME.Seconds())
		fmt.Println("BCharge kurtResult", ktCharge[0], math.Abs(ktCharge[0]-kurt_cg), math.Abs(ktCharge[0]-kurt_cg)/math.Abs(kurt_cg), CHARGE_KURT_TIME)

		TIME = time.Now()
		cvCharge, _ := engine.CoeffVar(charge)
		CHARGE_CV_TIME := time.Since(TIME)
		cCharge, _ := engine.Decrypt(cvCharge)
		CG_CV_MRE[i] = math.Abs(cCharge[0]-cv_cg) / math.Abs(cv_cg)
		CG_CV_MAE[i] = math.Abs(cCharge[0] - cv_cg)
		CG_CV_TIME[i] = float64(CHARGE_CV_TIME.Seconds())
		fmt.Println("Charge cvResult", cCharge[0], math.Abs(cCharge[0]-cv_cg), math.Abs(cCharge[0]-cv_cg)/math.Abs(cv_cg), CHARGE_CV_TIME)

		_, corrr1, _ := utils.Correlation(ageSlice, chargeSlice)
		TIME = time.Now()
		corr, _ := engine.PCorrCoeff(age, charge)
		AGE_CG_TIME := time.Since(TIME)
		corrResult, _ := engine.Decrypt(corr)
		AGE_CG_CORR_MAE[i] = math.Abs(corrResult[0] - corrr1)
		AGE_CG_CORR_MRE[i] = math.Abs(corrResult[0]-corrr1) / math.Abs(corrr1)
		AGE_CG_CORR_TIME[i] = float64(AGE_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE)", corrResult[0], math.Abs(corrResult[0]-corrr1), math.Abs(corrResult[0]-corrr1)/corrr1)

		_, corrr2, _ := utils.Correlation(bmiSlice, chargeSlice)
		TIME = time.Now()
		corr2, _ := engine.PCorrCoeff(bmi, charge)
		BMI_CG_TIME := time.Since(TIME)
		corrResult2, _ := engine.Decrypt(corr2)
		BMI_CG_CORR_MAE[i] = math.Abs(corrResult2[0] - corrr2)
		BMI_CG_CORR_MRE[i] = math.Abs(corrResult2[0]-corrr2) / math.Abs(corrr2)
		BMI_CG_CORR_TIME[i] = float64(BMI_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE)", corrResult2[0], math.Abs(corrResult2[0]-corrr2), math.Abs(corrResult2[0]-corrr2)/corrr2)

		_, corrr3, _ := utils.Correlation(smokerSlice, chargeSlice)
		TIME = time.Now()
		corr3, _ := engine.PCorrCoeff(smoker, charge)
		SMOKER_CG_TIME := time.Since(TIME)
		corrResult3, _ := engine.Decrypt(corr3)
		SMOKER_CG_CORR_MAE[i] = math.Abs(corrResult3[0] - corrr3)
		SMOKER_CG_CORR_MRE[i] = math.Abs(corrResult3[0]-corrr3) / math.Abs(corrr3)
		SMOKER_CG_CORR_TIME[i] = float64(SMOKER_CG_TIME.Seconds())
		fmt.Println("Correlation (BMI, CHARGE)", corrResult3[0], math.Abs(corrResult3[0]-corrr3), math.Abs(corrResult3[0]-corrr3)/corrr3, SMOKER_CG_TIME)
	}
	result := ""
	result += fmt.Sprintf("[ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_ZSCORE_MAE), utils.StdDev(CG_ZSCORE_MAE), utils.Mean(CG_ZSCORE_MRE), utils.StdDev(CG_ZSCORE_MRE), utils.Mean(CG_ZSCORE_TIME), utils.StdDev(CG_ZSCORE_TIME))
	result += fmt.Sprintf("[KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_KURT_MAE), utils.StdDev(CG_KURT_MAE), utils.Mean(CG_KURT_MRE), utils.StdDev(CG_KURT_MRE), utils.Mean(CG_KURT_TIME), utils.StdDev(CG_KURT_TIME))
	result += fmt.Sprintf("[SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_SKEW_MAE), utils.StdDev(CG_SKEW_MAE), utils.Mean(CG_SKEW_MRE), utils.StdDev(CG_SKEW_MRE), utils.Mean(CG_SKEW_TIME), utils.StdDev(CG_SKEW_TIME))
	result += fmt.Sprintf("[CV] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(CG_CV_MAE), utils.StdDev(CG_CV_MAE), utils.Mean(CG_CV_MRE), utils.StdDev(CG_CV_MRE), utils.Mean(CG_CV_TIME), utils.StdDev(CG_CV_TIME))
	result += fmt.Sprintf("[AGE_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AGE_CG_CORR_MAE), utils.StdDev(AGE_CG_CORR_MAE), utils.Mean(AGE_CG_CORR_MRE), utils.StdDev(AGE_CG_CORR_MRE), utils.Mean(AGE_CG_CORR_TIME), utils.StdDev(AGE_CG_CORR_TIME))
	result += fmt.Sprintf("[BMI_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(BMI_CG_CORR_MAE), utils.StdDev(BMI_CG_CORR_MAE), utils.Mean(BMI_CG_CORR_MRE), utils.StdDev(BMI_CG_CORR_MRE), utils.Mean(BMI_CG_CORR_TIME), utils.StdDev(BMI_CG_CORR_TIME))
	result += fmt.Sprintf("[SMOKER_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(SMOKER_CG_CORR_MAE), utils.StdDev(SMOKER_CG_CORR_MAE), utils.Mean(SMOKER_CG_CORR_MRE), utils.StdDev(SMOKER_CG_CORR_MRE), utils.Mean(SMOKER_CG_CORR_TIME), utils.StdDev(SMOKER_CG_CORR_TIME))
	io.WriteString(file, result)
}
