package utils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
)

func CheckMAE(x []float64, input []float64, targets []float64, len int) (float64, float64) {
	sum := 0.0
	max := 0.0
	elm := 0.0
	for i := range len {
		diff := math.Abs(x[i] - targets[i])
		sum += diff
		if diff > max {
			elm = input[i]
		}
	}
	avg := sum / float64(len)
	return elm, avg
}

func CheckMRE(x []float64, input []float64, targets []float64, len int) (float64, float64) {
	sum := 0.0
	max := 0.0
	elm := 0.0
	for i := range len {
		diff := math.Abs(1 - x[i]/targets[i])
		sum += diff
		if diff > max {
			max = diff
			elm = input[i]
		}
	}
	avg := sum / float64(len)
	return elm, avg
}

/*
- num:
- len:
- ran:
*/
func DataPosGenerator(num int, len int, ran float64) [][]float64 {
	values := make([][]float64, num)
	for i := range num {
		values[i] = make([]float64, len)
		for j := range len {
			values[i][j] = (rand.Float64()*ran + 0.0)
			// values[i][j] = (rand.Float64()*0.001 + 0.0001) * 1
		}
	}
	return values
}

// IsPowerOfTwo returns true if x is a power of two (≥ 1).
func IsPowerOfTwo(x float64) bool {
	if x < 1 {
		return false
	}
	_, frac := math.Modf(math.Log2(x))
	return frac == 0
}

func ReadCSV(fileName string, index int) ([]float64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("file open failed: %w", err)
	}
	rdr := csv.NewReader(bufio.NewReader(file))

	rows, err := rdr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("file read failed: %w", err)
	}

	data := []float64{}
	for i, v := range rows {
		if i != 0 {
			f, err := strconv.ParseFloat(v[index], 64)
			if err == nil {
				data = append(data, f)
			} else {
				if v[index] == "yes" {
					data = append(data, 1.0)
				} else if v[index] == "no" {
					data = append(data, 0.0)
				}
			}
		}
	}
	return data, nil
}
