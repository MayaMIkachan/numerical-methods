package integrate

import (
	"errors"
	"math"
	"math/rand"
)

var (
	// ErrWrongDimensions is ...
	ErrWrongDimensions = errors.New("wrong dimensions")
)

// MonteCarlo returns multidim integral
func MonteCarlo(
	f func([]float64) float64,
	left, right []float64,
	points int,
) (float64, error) {
	n := len(left)
	if n != len(right) {
		return 0, ErrWrongDimensions
	}

	var (
		x      = make([]float64, n)
		sum    = 0.0
		volume = 1.0
	)

	for i := 0; i < n; i++ {
		volume *= right[i] - left[i]
	}

	for i := 0; i < points; i++ {
		randomVector(left, right, x)
		sum += f(x)
	}

	return sum * volume / float64(points), nil
}

func randomVector(a, b, v []float64) {
	n := len(a)

	for i := 0; i < n; i++ {
		v[i] = a[i] + rand.Float64()*(b[i]-a[i])
	}
}

// Quadrature returns multidim integral
func Quadrature(
	f func([]float64) float64,
	left, right []float64,
	h []float64,
) (float64, error) {
	n := len(left)
	if n != len(right) || n != len(h) {
		return 0, ErrWrongDimensions
	}

	var (
		calculate func() float64
		k         int
		x         = make([]float64, n)
	)
	copy(x, left)

	calculate = func() float64 {
		var (
			sum   = 0.0
			w     = right[k] - left[k]
			count = int(math.Round(w / h[k]))
		)
		if k < n-1 {
			k++
			sum += 0.5 * calculate()

			for i := 1; i < count; i++ {
				x[k] = left[k] + float64(i)*h[k]
				k++
				sum += calculate()
			}

			x[k] = right[k]
			k++
			sum += 0.5 * calculate()
		} else {
			sum += 0.5 * f(x)
			for i := 1; i < count; i++ {
				x[k] = left[k] + float64(i)*h[k]
				sum += f(x)
			}
			x[k] = right[k]
			sum += 0.5 * f(x)
		}
		sum *= h[k]

		x[k] = left[k]
		k--

		return sum
	}

	return calculate(), nil
}
