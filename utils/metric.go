package utils

import "math/rand/v2"

// Linspace generates a slice of n evenly spaced values between start and stop
func Linspace(start, stop float64, n int) []float64 {
	if n < 2 {
		panic("n must be at least 2")
	}
	step := (stop - start) / float64(n-1)
	linspace := make([]float64, n)

	for i := 0; i < n; i++ {
		linspace[i] = start + step*float64(i)
	}

	return linspace
}

// DataGenerator generates a 2D slice of random float64 values
func DataGenerator(num int, len int, ran float64) [][]float64 {
	values := make([][]float64, num)
	for i := range num {
		values[i] = make([]float64, len)
		for j := range len {
			values[i][j] = (rand.Float64() - 0.5) * ran
		}
	}
	return values
}
