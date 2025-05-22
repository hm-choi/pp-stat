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
	A := 32 * 32 * 32
	B := 100.0
	START, MIDDLE, STOP := 0.001, 1.0, B
	test1 := utils.Linspace(START, MIDDLE, A/2)
	test2 := utils.Linspace(MIDDLE, STOP, A/2)
	test := make([]float64, A)
	invS := make([]float64, A)
	for i := 0; i < A; i++ {
		if i < A/2 {
			test[i] = test1[i]
		} else {
			test[i] = test2[i-A/2]
		}
		invS[i] = 1.0 / math.Sqrt(test[i])
	}

	EVAL_NUM := 10

	file, err := os.Create("output.txt")
	if err != nil {
		return
	}
	defer file.Close()

	PIVOT_MRE, PIVOT_MAE, PIVOT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	HSTAT_MRE, HSTAT_MAE, HSTAT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)
	INVSQRT_MRE, INVSQRT_MAE, INVSQRT_TIME := make([]float64, EVAL_NUM), make([]float64, EVAL_NUM), make([]float64, EVAL_NUM)

	for i := 0; i < EVAL_NUM; i++ {
		ct, _ := engine.Encrypt(test)

		TIME := time.Now()
		pivotTangent, _ := invsqrt.PivotTangent(engine, ct, 12, 7)
		PivotTangentTime := time.Since(TIME)
		ptResult, _ := engine.Decrypt(pivotTangent)
		_, ptMRE := utils.CheckMRE(invS, invS, ptResult, ct.Size())
		_, ptMAE := utils.CheckMAE(invS, invS, ptResult, ct.Size())
		PIVOT_MRE[i], PIVOT_MAE[i], PIVOT_TIME[i] = ptMRE, ptMAE, PivotTangentTime.Seconds()
		fmt.Println("PT", ptResult[:3], ptMRE, ptMAE, PivotTangentTime)

		TIME = time.Now()
		heStat, _ := invsqrt.HEStat(engine, ct, 21, B)
		HEStatTime := time.Since(TIME)
		hsResult, _ := engine.Decrypt(heStat)
		_, hsMRE := utils.CheckMRE(invS, invS, hsResult, ct.Size())
		_, hsMAE := utils.CheckMAE(invS, invS, hsResult, ct.Size())
		HSTAT_MRE[i], HSTAT_MAE[i], HSTAT_TIME[i] = hsMRE, hsMAE, HEStatTime.Seconds()
		fmt.Println("HS", hsResult[:3], hsMRE, hsMAE, HEStatTime)

		TIME = time.Now()
		cryptoInvSqrt, _ := engine.CryptoInvSqrt(ct)
		cryptoInvSqrtTime := time.Since(TIME)
		cisResult, _ := engine.Decrypt(cryptoInvSqrt)
		_, cisMRE := utils.CheckMRE(invS, invS, cisResult, ct.Size())
		_, cisMAE := utils.CheckMAE(invS, invS, cisResult, ct.Size())
		INVSQRT_MRE[i], INVSQRT_MAE[i], INVSQRT_TIME[i] = cisMRE, cisMAE, cryptoInvSqrtTime.Seconds()
		fmt.Println("CIS", cisResult[:3], cisMRE, cisMAE, cryptoInvSqrtTime)
	}
	result := fmt.Sprintf("[PIVOT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(PIVOT_MAE), utils.StdDev(PIVOT_MAE), utils.Mean(PIVOT_MRE), utils.StdDev(PIVOT_MRE), utils.Mean(PIVOT_TIME), utils.StdDev(PIVOT_TIME))
	result += fmt.Sprintf("[HSTAT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(HSTAT_MAE), utils.StdDev(HSTAT_MAE), utils.Mean(HSTAT_MRE), utils.StdDev(HSTAT_MRE), utils.Mean(HSTAT_TIME), utils.StdDev(HSTAT_TIME))
	result += fmt.Sprintf("[PSTAT] MAE %e (%e), MRE %e (%e), TIME %f (%f)\n", utils.Mean(INVSQRT_MAE), utils.StdDev(INVSQRT_MAE), utils.Mean(INVSQRT_MRE), utils.StdDev(INVSQRT_MRE), utils.Mean(INVSQRT_TIME), utils.StdDev(INVSQRT_TIME))
	io.WriteString(file, result)
}
