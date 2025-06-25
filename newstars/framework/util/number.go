package util

import (
	"math"
	"math/rand"
)

func Precision(f float64, prec int, round bool) float64 {
	pow10N := math.Pow10(prec)
	if round {
		return (math.Trunc(f+0.5/pow10N) * pow10N) / pow10N
	}
	return math.Trunc((f)*pow10N) / pow10N
}

func HandleFloat(wealth float64) float64 {
	fwealth := rand.Float64()
	newvalue := Precision(float64(wealth)+fwealth, 3, false)
	return newvalue
}
