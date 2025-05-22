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
	EVAL_NUM, B := 10, 20.0
	ageSlice, _ := utils.ReadCSV("../../examples/dataset/adult_dataset.csv", 0)
	hpwSlice, _ := utils.ReadCSV("../../examples/dataset/adult_dataset.csv", 12)
	eduSlice, _ := utils.ReadCSV("../../examples/dataset/adult_dataset.csv", 4)

	age, _ := engine.Encrypt(ageSlice)
	hpw, _ := engine.Encrypt(hpwSlice)
	edu, _ := engine.Encrypt(eduSlice)

	_, _, skew_age := utils.Skewness(ageSlice)
	_, _, kurt_age := utils.Kurtosis(ageSlice)
	_, _, cfvar_age := utils.CoeffVar(ageSlice)

	_, _, skew_hpw := utils.Skewness(hpwSlice)
	_, _, kurt_hpw := utils.Kurtosis(hpwSlice)
	_, _, cfvar_hpw := utils.CoeffVar(hpwSlice)

	_, _, skew_edu := utils.Skewness(eduSlice)
	_, _, kurt_edu := utils.Kurtosis(eduSlice)
	_, _, cfvar_edu := utils.CoeffVar(eduSlice)

	AgeNorm_MRE, AgeNorm_MAE, AgeNorm_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AgeSkew_MRE, AgeSkew_MAE, AgeSkew_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AgeKurt_MRE, AgeKurt_MAE, AgeKurt_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AgeCV_MRE, AgeCV_MAE, AgeCV_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	HPWNorm_MRE, HPWNorm_MAE, HPWNorm_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HPWSkew_MRE, HPWSkew_MAE, HPWSkew_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HPWKurt_MRE, HPWKurt_MAE, HPWKurt_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HPWCV_MRE, HPWCV_MAE, HPWCV_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	EduNorm_MRE, EduNorm_MAE, EduNorm_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	EduSkew_MRE, EduSkew_MAE, EduSkew_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	EduKurt_MRE, EduKurt_MAE, EduKurt_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	EduCV_MRE, EduCV_MAE, EduCV_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	AgeHPW_CORR_MRE, AgeHPW_CORR_MAE, AgeHPW_CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	AGE_EDU_CORR_MRE, AGE_EDU_CORR_MAE, AGE_EDU_CORR_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	file, err := os.Create("output.txt")
	if err != nil {
		return
	}
	defer file.Close()

	for i := 0; i < int(EVAL_NUM); i++ {
		fmt.Println("[============ Age Test ============]")
		TIME := time.Now()
		zScoreNorm1, _ := engine.ZScoreNorm(age, 100.0)
		AGE_ZNORM_TIME := time.Since(TIME)
		zSNAge, _ := engine.Decrypt(zScoreNorm1)

		fmt.Println("Age ZNorm", zSNAge[0:1], utils.ZScoreNorm(ageSlice)[:1], AGE_ZNORM_TIME)

		TIME = time.Now()
		skew1, _ := engine.Skewness(age, B)
		AGE_SKEW_TIME := time.Since(TIME)
		skewAge, _ := engine.Decrypt(skew1)
		AgeSkew_MRE[i] = math.Abs(skewAge[0]-skew_age) / math.Abs(skew_age)
		AgeSkew_MAE[i] = math.Abs(skewAge[0] - skew_age)
		AgeSkew_TIME[i] = float64(AGE_SKEW_TIME.Seconds())
		fmt.Println("Age skewResult", skewAge[0], math.Abs(skewAge[0]-skew_age), math.Abs(skewAge[0]-skew_age)/math.Abs(skew_age), AGE_SKEW_TIME)

		TIME = time.Now()
		kurt1, _ := engine.Kurtosis(age, B)
		AGE_KURT_TIME := time.Since(TIME)
		kurtAge, _ := engine.Decrypt(kurt1)

		AgeKurt_MRE[i] = math.Abs(kurtAge[0]-kurt_age) / math.Abs(kurt_age)
		AgeKurt_MAE[i] = math.Abs(kurtAge[0] - kurt_age)
		AgeKurt_TIME[i] = float64(AGE_KURT_TIME.Seconds())
		fmt.Println("Age kurtResult", kurtAge[0], math.Abs(kurtAge[0]-kurt_age), math.Abs(kurtAge[0]-kurt_age)/math.Abs(kurt_age), AGE_KURT_TIME)

		TIME = time.Now()
		ceffvar1, _ := engine.CoeffVar(age, B)
		AGE_CV_TIME := time.Since(TIME)
		ceffvarAge, _ := engine.Decrypt(ceffvar1)
		AgeCV_MRE[i] = math.Abs(ceffvarAge[0]-cfvar_age) / math.Abs(cfvar_age)
		AgeCV_MAE[i] = math.Abs(ceffvarAge[0] - cfvar_age)
		AgeCV_TIME[i] = float64(AGE_CV_TIME.Seconds())
		fmt.Println("Age ceffvarResult", ceffvarAge[0], math.Abs(ceffvarAge[0]-cfvar_age), math.Abs(ceffvarAge[0]-cfvar_age)/math.Abs(cfvar_age), AGE_CV_TIME)

		fmt.Println("[============ HPW Test ============]")
		TIME = time.Now()
		zScoreNorm2, _ := engine.ZScoreNorm(hpw, 100.0)
		HPW_ZNORM_TIME := time.Since(TIME)
		zSNHpw, _ := engine.Decrypt(zScoreNorm2)

		fmt.Println("HPW ZNorm", zSNHpw[0:1], utils.ZScoreNorm(zSNHpw)[:1], HPW_ZNORM_TIME)

		TIME = time.Now()
		skew2, _ := engine.Skewness(hpw, B)
		HPW_SKEW_TIME := time.Since(TIME)
		skewHpw, _ := engine.Decrypt(skew2)

		HPWSkew_MRE[i] = math.Abs(skewHpw[0]-skew_hpw) / math.Abs(skew_hpw)
		HPWSkew_MAE[i] = math.Abs(skewHpw[0] - skew_hpw)
		HPWSkew_TIME[i] = float64(HPW_SKEW_TIME.Seconds())
		fmt.Println("HPW skewResult", skewHpw[0], math.Abs(skewHpw[0]-skew_hpw), math.Abs(skewHpw[0]-skew_hpw)/math.Abs(skew_hpw), HPW_SKEW_TIME)

		TIME = time.Now()
		kurt2, _ := engine.Kurtosis(hpw, B)
		HPW_KURT_TIME := time.Since(TIME)
		kurtHpw, _ := engine.Decrypt(kurt2)
		HPWKurt_MRE[i] = math.Abs(kurtHpw[0]-kurt_hpw) / math.Abs(kurt_hpw)
		HPWKurt_MAE[i] = math.Abs(kurtHpw[0] - kurt_hpw)
		HPWKurt_TIME[i] = float64(HPW_KURT_TIME.Seconds())
		fmt.Println("HPW kurtResult", kurtHpw[0], math.Abs(kurtHpw[0]-kurt_hpw), math.Abs(kurtHpw[0]-kurt_hpw)/math.Abs(kurt_hpw), HPW_KURT_TIME)

		TIME = time.Now()
		ceffvar2, _ := engine.CoeffVar(hpw, B)
		HPW_CV_TIME := time.Since(TIME)
		ceffvarAHpw, _ := engine.Decrypt(ceffvar2)

		HPWCV_MRE[i] = math.Abs(ceffvarAHpw[0]-cfvar_hpw) / math.Abs(cfvar_hpw)
		HPWCV_MAE[i] = math.Abs(ceffvarAHpw[0] - cfvar_hpw)
		HPWCV_TIME[i] = float64(HPW_CV_TIME.Seconds())
		fmt.Println("HPW ceffvarResult", ceffvarAHpw[0], math.Abs(ceffvarAHpw[0]-cfvar_hpw), math.Abs(ceffvarAHpw[0]-cfvar_hpw)/math.Abs(cfvar_hpw), HPW_CV_TIME)

		fmt.Println("[============ Edu Test ============]")
		TIME = time.Now()
		zScoreNorm3, _ := engine.ZScoreNorm(edu, 100.0)
		EDU_ZNORM_TIME := time.Since(TIME)
		zSNEdu, _ := engine.Decrypt(zScoreNorm3)

		fmt.Println("Edu ZNorm", zSNEdu[0:1], utils.ZScoreNorm(zSNEdu)[:1], EDU_ZNORM_TIME)

		TIME = time.Now()
		skew3, _ := engine.Skewness(edu, B)
		EDU_SKEW_TIME := time.Since(TIME)
		skewEdu, _ := engine.Decrypt(skew3)

		EduSkew_MRE[i] = math.Abs(skewEdu[0]-skew_edu) / math.Abs(skew_edu)
		EduSkew_MAE[i] = math.Abs(skewEdu[0] - skew_edu)
		EduSkew_TIME[i] = float64(EDU_SKEW_TIME.Seconds())
		fmt.Println("Edu skewResult", skewEdu[0], math.Abs(skewEdu[0]-skew_edu), math.Abs(skewEdu[0]-skew_edu)/math.Abs(skew_edu), EDU_SKEW_TIME)

		TIME = time.Now()
		kurt3, _ := engine.Kurtosis(edu, B)
		EDU_KURT_TIME := time.Since(TIME)
		kurtEdu, _ := engine.Decrypt(kurt3)

		EduKurt_MRE[i] = math.Abs(kurtEdu[0]-kurt_edu) / math.Abs(kurt_edu)
		EduKurt_MAE[i] = math.Abs(kurtEdu[0] - kurt_edu)
		EduKurt_TIME[i] = float64(EDU_KURT_TIME.Seconds())
		fmt.Println("Edu kurtResult", kurtEdu[0], math.Abs(kurtEdu[0]-kurt_edu), math.Abs(kurtEdu[0]-kurt_edu)/math.Abs(kurt_edu), EDU_KURT_TIME)

		TIME = time.Now()
		ceffvar3, _ := engine.CoeffVar(edu, B)
		EDU_CV_TIME := time.Since(TIME)
		ceffvarEdu, _ := engine.Decrypt(ceffvar3)

		EduCV_MRE[i] = math.Abs(ceffvarEdu[0]-cfvar_edu) / math.Abs(cfvar_edu)
		EduCV_MAE[i] = math.Abs(ceffvarEdu[0] - cfvar_edu)
		EduCV_TIME[i] = float64(EDU_CV_TIME.Seconds())
		fmt.Println("Edu ceffvarResult", ceffvarEdu[0], math.Abs(ceffvarEdu[0]-cfvar_edu), math.Abs(ceffvarEdu[0]-cfvar_edu)/math.Abs(cfvar_edu), EDU_CV_TIME)

		fmt.Println("[============ Corr Test ============]")
		_, corrr, _ := utils.Correlation(ageSlice, hpwSlice)
		TIME = time.Now()
		corr, _ := engine.PCorrCoeff(age, hpw, B)
		CORR_TIME := time.Since(TIME)
		corrtResult, _ := engine.Decrypt(corr)

		AgeHPW_CORR_MRE[i] = math.Abs(corrtResult[0]-corrr) / math.Abs(corrr)
		AgeHPW_CORR_MAE[i] = math.Abs(corrtResult[0] - corrr)
		AgeHPW_CORR_TIME[i] = float64(CORR_TIME.Seconds())
		fmt.Println("corrtResult(AGE vs HPW)", corrtResult[0], math.Abs(corrtResult[0]-corrr), math.Abs(corrtResult[0]-corrr)/math.Abs(corrr), CORR_TIME)

		_, corrr, _ = utils.Correlation(ageSlice, eduSlice)
		TIME = time.Now()
		corr, _ = engine.PCorrCoeff(age, edu, B)
		CORR_TIME = time.Since(TIME)
		corrtResult, _ = engine.Decrypt(corr)
		AGE_EDU_CORR_MRE[i] = math.Abs(corrtResult[0]-corrr) / math.Abs(corrr)
		AGE_EDU_CORR_MAE[i] = math.Abs(corrtResult[0] - corrr)
		AGE_EDU_CORR_TIME[i] = float64(CORR_TIME.Seconds())
		fmt.Println("corrtResult(AGE vs EDU)", corrtResult[0], math.Abs(corrtResult[0]-corrr), math.Abs(corrtResult[0]-corrr)/math.Abs(corrr), CORR_TIME)

		ageZNorm := utils.ZScoreNorm(ageSlice)
		hpwZNorm := utils.ZScoreNorm(hpwSlice)
		eduZNorm := utils.ZScoreNorm(eduSlice)
		_, zScoreMreAge := utils.CheckMRE(ageZNorm, zSNAge, zSNAge, len(ageSlice))
		_, zScoreMaeAge := utils.CheckMAE(ageZNorm, zSNAge, zSNAge, len(ageSlice))

		_, zScoreMreHpw := utils.CheckMRE(hpwZNorm, zSNHpw, zSNHpw, len(hpwSlice))
		_, zScoreMaeHpw := utils.CheckMAE(hpwZNorm, zSNHpw, zSNHpw, len(hpwSlice))

		_, zScoreMreEdu := utils.CheckMRE(eduZNorm, zSNEdu, zSNEdu, len(eduSlice))
		_, zScoreMaeEdu := utils.CheckMAE(eduZNorm, zSNEdu, zSNEdu, len(eduSlice))

		AgeNorm_MRE[i] = zScoreMreAge
		AgeNorm_MAE[i] = zScoreMaeAge
		AgeNorm_TIME[i] = float64(AGE_ZNORM_TIME.Seconds())
		HPWNorm_MRE[i] = zScoreMreHpw
		HPWNorm_MAE[i] = zScoreMaeHpw
		HPWNorm_TIME[i] = float64(HPW_ZNORM_TIME.Seconds())
		EduNorm_MRE[i] = zScoreMreEdu
		EduNorm_MAE[i] = zScoreMaeEdu
		EduNorm_TIME[i] = float64(EDU_ZNORM_TIME.Seconds())
		fmt.Println("ZNorm Age (MRE, MAE)", zScoreMreAge, zScoreMaeAge, AGE_ZNORM_TIME)
		fmt.Println("ZNorm Hpw (MRE, MAE)", zScoreMreHpw, zScoreMaeHpw, HPW_ZNORM_TIME)
		fmt.Println("ZNorm Edu (MRE, MAE)", zScoreMreEdu, zScoreMaeEdu, EDU_ZNORM_TIME)
	}

	fmt.Println("[============ END ============]")
	result := fmt.Sprintf("[AGE_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeNorm_MAE), utils.StdDev(AgeNorm_MAE), utils.Mean(AgeNorm_MRE), utils.StdDev(AgeNorm_MRE), utils.Mean(AgeNorm_TIME), utils.StdDev(AgeNorm_TIME))
	result += fmt.Sprintf("[AGE_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeSkew_MAE), utils.StdDev(AgeSkew_MAE), utils.Mean(AgeSkew_MRE), utils.StdDev(AgeSkew_MRE), utils.Mean(AgeSkew_TIME), utils.StdDev(AgeSkew_TIME))
	result += fmt.Sprintf("[AGE_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeKurt_MAE), utils.StdDev(AgeKurt_MAE), utils.Mean(AgeKurt_MRE), utils.StdDev(AgeKurt_MRE), utils.Mean(AgeKurt_TIME), utils.StdDev(AgeKurt_TIME))
	result += fmt.Sprintf("[AGE_CV] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeCV_MAE), utils.StdDev(AgeCV_MAE), utils.Mean(AgeCV_MRE), utils.StdDev(AgeCV_MRE), utils.Mean(AgeCV_TIME), utils.StdDev(AgeCV_TIME))
	result += fmt.Sprintf("[HPW_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWNorm_MAE), utils.StdDev(HPWNorm_MAE), utils.Mean(HPWNorm_MRE), utils.StdDev(HPWNorm_MRE), utils.Mean(HPWNorm_TIME), utils.StdDev(HPWNorm_TIME))
	result += fmt.Sprintf("[HPW_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWSkew_MAE), utils.StdDev(HPWSkew_MAE), utils.Mean(HPWSkew_MRE), utils.StdDev(HPWSkew_MRE), utils.Mean(HPWSkew_TIME), utils.StdDev(HPWSkew_TIME))
	result += fmt.Sprintf("[HPW_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWKurt_MAE), utils.StdDev(HPWKurt_MAE), utils.Mean(HPWKurt_MRE), utils.StdDev(HPWKurt_MRE), utils.Mean(HPWKurt_TIME), utils.StdDev(HPWKurt_TIME))
	result += fmt.Sprintf("[HPW_CV] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HPWCV_MAE), utils.StdDev(HPWCV_MAE), utils.Mean(HPWCV_MRE), utils.StdDev(HPWCV_MRE), utils.Mean(HPWCV_TIME), utils.StdDev(HPWCV_TIME))
	result += fmt.Sprintf("[EDU_ZSCORE] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduNorm_MAE), utils.StdDev(EduNorm_MAE), utils.Mean(EduNorm_MRE), utils.StdDev(EduNorm_MRE), utils.Mean(EduNorm_TIME), utils.StdDev(EduNorm_TIME))
	result += fmt.Sprintf("[EDU_SKEW] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduSkew_MAE), utils.StdDev(EduSkew_MAE), utils.Mean(EduSkew_MRE), utils.StdDev(EduSkew_MRE), utils.Mean(EduSkew_TIME), utils.StdDev(EduSkew_TIME))
	result += fmt.Sprintf("[EDU_KURT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduKurt_MAE), utils.StdDev(EduKurt_MAE), utils.Mean(EduKurt_MRE), utils.StdDev(EduKurt_MRE), utils.Mean(EduKurt_TIME), utils.StdDev(EduKurt_TIME))
	result += fmt.Sprintf("[EDU_CV] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(EduCV_MAE), utils.StdDev(EduCV_MAE), utils.Mean(EduCV_MRE), utils.StdDev(EduCV_MRE), utils.Mean(EduCV_TIME), utils.StdDev(EduCV_TIME))
	result += fmt.Sprintf("[AGEHPW_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AgeHPW_CORR_MAE), utils.StdDev(AgeHPW_CORR_MAE), utils.Mean(AgeHPW_CORR_MRE), utils.StdDev(AgeHPW_CORR_MRE), utils.Mean(AgeHPW_CORR_TIME), utils.StdDev(AgeHPW_CORR_TIME))
	result += fmt.Sprintf("[AGEEDU_CORR] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(AGE_EDU_CORR_MAE), utils.StdDev(AGE_EDU_CORR_MAE), utils.Mean(AGE_EDU_CORR_MRE), utils.StdDev(AGE_EDU_CORR_MRE), utils.Mean(AGE_EDU_CORR_TIME), utils.StdDev(AGE_EDU_CORR_TIME))
	io.WriteString(file, result)
}
