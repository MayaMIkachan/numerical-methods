package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/MayaMIkachan/numerical-methods/internal/integrate"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func f(args []float64) float64 {
	var sum float64
	for _, value := range args {
		sum += value * value
	}
	return sum
}

func main() {
	outputPath := flag.String("output", "report.json", "output path")
	flag.Parse()

	log, err := newLogger()
	if err != nil {
		errorf("%v\n", err)
	}

	var (
		points = []int{
			7500,
			195000,
			802500,
			2070000,
			4237501,
			7545001,
			12232501,
			18540001,
			26707500,
			36975001,
		}
		h = []float64{
			0.2,
			0.166667,
			0.125,
			0.111111,
			0.1,
			0.1,
			0.0909091,
			0.0833333,
			0.0769231,
			0.0769231,
		}
		refValues = []float64{1.0 / 3.0, 20.0 / 3.0, 27.0, 208.0 / 3.0, 425.0 / 3.0, 252.0, 1225.0 / 3.0, 1856.0 / 3, 891.0, 3700.0 / 3.0}
	)

	stats := []report{}
	for i := 1; i <= 9; i++ {
		var (
			a    = make([]float64, i)
			b    = make([]float64, i)
			grid = make([]float64, i)
		)

		for j := 0; j < i; j++ {
			a[j] = float64(2 * j)
			b[j] = float64(2*j + 1)
			grid[j] = h[i-1]
		}
		stat, err := calculate(
			log,
			i,
			a,
			b,
			points[i-1],
			grid,
			refValues[i-1],
		)
		if err != nil {
			errorf("%v\n", err)
		}
		stats = append(stats, stat)
	}

	marshaled, err := json.Marshal(stats)
	if err != nil {
		errorf("%v\n", err)
	}
	file, err := os.Create(*outputPath)
	if err != nil {
		errorf("%v\n", err)
	}
	if _, err := file.Write(marshaled); err != nil {
		errorf("%v\n", err)
	}
}

func newLogger() (*zap.Logger, error) {
	conf := zap.NewDevelopmentConfig()
	conf.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return conf.Build()
}

type stat struct {
	InvokeCount   int64
	ExecutionTime time.Duration
}

type report struct {
	MonteCarlo stat
	Quadrature stat
}

func calculate(
	log *zap.Logger,
	n int,
	a []float64,
	b []float64,
	points int,
	h []float64,
	value float64,
) (report, error) {
	r := report{}

	{
		log := log.Named("MonteCarlo")
		f, counter := withInvokeCounter(f)
		start := time.Now()
		sum, err := integrate.MonteCarlo(
			f,
			a,
			b,
			points,
		)
		if err != nil {
			log.Error("calculate", zap.Error(err))
			return report{}, err
		}
		r.MonteCarlo = stat{
			InvokeCount:   counter.Load(),
			ExecutionTime: time.Since(start),
		}
		log.Info(
			"calculate",
			zap.Int("n", n),
			zap.Float64("sum", sum),
			zap.Int64("invokeCount", r.MonteCarlo.InvokeCount),
			zap.Duration("executionTime", r.MonteCarlo.ExecutionTime),
			zap.Float64("error", math.Abs(value-sum)),
		)
	}

	{
		log := log.Named("Quadrature")
		f, counter := withInvokeCounter(f)
		start := time.Now()
		sum, err := integrate.Quadrature(
			f,
			a,
			b,
			h,
		)
		if err != nil {
			log.Error("calculate", zap.Error(err))
			return report{}, err
		}
		r.Quadrature = stat{
			InvokeCount:   counter.Load(),
			ExecutionTime: time.Since(start),
		}
		log.Info(
			"calculate",
			zap.Int("n", n),
			zap.Float64("sum", sum),
			zap.Int64("invokeCount", r.Quadrature.InvokeCount),
			zap.Duration("executionTime", r.Quadrature.ExecutionTime),
			zap.Float64("error", math.Abs(value-sum)),
		)
	}

	return r, nil
}

func withInvokeCounter(f func([]float64) float64) (func([]float64) float64, *atomic.Int64) {
	counter := atomic.NewInt64(0)

	return func(args []float64) float64 {
		counter.Add(1)
		return f(args)
	}, counter
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
