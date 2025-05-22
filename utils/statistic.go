package utils

import (
	"fmt"
	"math"
)

func Inverse(data []float64) []float64 {
	inv := make([]float64, len(data))
	for i, v := range data {
		if v != 0 {
			inv[i] = 1 / v
		} else {
			inv[i] = 0.0
		}
	}
	return inv
}

// Mean returns the mean of the input slice.
func Mean(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// Variance returns the variance of the input slice.
func Variance(data []float64) float64 {
	m := Mean(data)
	sum := 0.0
	for _, v := range data {
		sum += (v - m) * (v - m)
	}
	return sum / float64(len(data))
}

// StdDev returns the standard deviation of the input slice.
func StdDev(data []float64) float64 {
	return math.Sqrt(Variance(data))
}

func ZScoreNorm(data []float64) []float64 {
	mean := Mean(data)
	invSigma := 1 / StdDev(data)
	result := make([]float64, len(data))
	for i := range len(data) {
		result[i] = (data[i] - mean) * invSigma
	}
	return result
}

// Covariance returns the covariance between two input slices x and y.
func Covariance(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("두 슬라이스의 길이가 동일해야 합니다")
	}
	n := len(x)
	if n < 2 {
		return 0, fmt.Errorf("데이터 개수가 2개 이상이어야 공분산을 계산할 수 있습니다")
	}

	meanX := Mean(x)
	meanY := Mean(y)
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += (x[i] - meanX) * (y[i] - meanY)
	}
	return sum / float64(n), nil
}

// Correlation returns the Pearson correlation coefficient between two slices x and y.
func Correlation(x, y []float64) (float64, float64, error) {
	if len(x) != len(y) {
		return 0, 0, fmt.Errorf("Length of two slinces must be same.")
	}
	if len(x) < 2 {
		return 0, 0, fmt.Errorf("The correlation coefficient can be calculated only when the number of data is at least two")
	}

	cov, err := Covariance(x, y)
	if err != nil {
		return 0, 0, err
	}

	stdX := StdDev(x)
	stdY := StdDev(y)

	if stdX == 0 || stdY == 0 {
		return 0, 0, fmt.Errorf("The correlation coefficient cannot be calculated because the standard deviation is zero.")
	}

	return cov, cov / (stdX * stdY), nil
}

// Kurtosis returns the kurtosis of input slice x.
func Kurtosis(data []float64) (float64, float64, float64) {
	n := float64(len(data))
	if n < 4 {
		panic("hi") // 첨도 계산을 위해 최소 4개 이상의 데이터 필요
	}

	mean := Mean(data)
	stdDev := StdDev(data)

	sum := 0.0
	for _, v := range data {
		sum += math.Pow((v-mean)/stdDev, 4)
	}

	kurtosis := (sum / n) - 3 // Excess Kurtosis (정규분포의 첨도 3을 기준으로)
	return mean, stdDev, kurtosis
}

// Kurtosis returns the kurtosis of input slice x.
func Skewness(data []float64) (float64, float64, float64) {
	n := float64(len(data))
	if n < 4 {
		panic("hi") // 첨도 계산을 위해 최소 4개 이상의 데이터 필요
	}

	mean := Mean(data)
	stdDev := StdDev(data)

	sum := 0.0
	for _, v := range data {
		sum += math.Pow((v-mean)/stdDev, 3)
	}

	kurtosis := (sum / n) //- 3 // Excess Kurtosis (정규분포의 첨도 3을 기준으로)
	return mean, stdDev, kurtosis
}

// CoeffVar calculates the coefficient of variation of a slice of float64 numbers.
// Returns an error if the slice is empty or the mean is zero.
func CoeffVar(data []float64) (mean float64, stdDev float64, coeffVar float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, value := range data {
		sum += value
	}
	mean = sum / float64(len(data))

	if mean == 0 {
		return 0, 0, 0
	}

	// Calculate standard deviation
	varianceSum := 0.0
	for _, value := range data {
		diff := value - mean
		varianceSum += diff * diff
	}
	stdDev = math.Sqrt(varianceSum / float64(len(data)))

	// Calculate coefficient of variation
	coeffVar = stdDev / mean
	return mean, stdDev, coeffVar
}
