package main

import (
	"fmt"
	"math"
)

// binarySearch is an algorithm: the rules are written by a programmer.
func binarySearch(sorted []int, target int) int {
	left, right := 0, len(sorted)-1

	for left <= right {
		middle := left + (right-left)/2

		switch {
		case sorted[middle] == target:
			return middle
		case sorted[middle] < target:
			left = middle + 1
		default:
			right = middle - 1
		}
	}

	return -1
}

// LinearModel predicts y from x using y = weight*x + bias.
type LinearModel struct {
	Weight float64
	Bias   float64
}

func (m LinearModel) Predict(x float64) float64 {
	return m.Weight*x + m.Bias
}

// trainLinearModel learns the best straight line from labeled examples.
func trainLinearModel(features, labels []float64) LinearModel {
	if len(features) != len(labels) || len(features) == 0 {
		panic("features and labels must have the same non-zero length")
	}

	var meanX, meanY float64
	for i := range features {
		meanX += features[i]
		meanY += labels[i]
	}
	meanX /= float64(len(features))
	meanY /= float64(len(labels))

	var numerator, denominator float64
	for i := range features {
		deltaX := features[i] - meanX
		numerator += deltaX * (labels[i] - meanY)
		denominator += deltaX * deltaX
	}
	if denominator == 0 {
		panic("features must not all be the same")
	}

	weight := numerator / denominator
	return LinearModel{
		Weight: weight,
		Bias:   meanY - weight*meanX,
	}
}

func main() {
	fmt.Println("1. Algorithm: binary search")
	numbers := []int{3, 8, 12, 19, 24, 31, 42}
	target := 24
	index := binarySearch(numbers, target)
	fmt.Printf("Find %d in %v: index = %d\n\n", target, numbers, index)

	fmt.Println("2. Model: learn a line from data")
	studyHours := []float64{1, 2, 3, 4, 5}
	testScores := []float64{55, 62, 68, 76, 83}
	model := trainLinearModel(studyHours, testScores)

	futureHours := 6.0
	prediction := model.Predict(futureHours)
	fmt.Printf("Learned model: score = %.2f * hours + %.2f\n", model.Weight, model.Bias)
	fmt.Printf("Predicted score after %.0f hours: %.1f\n", futureHours, math.Round(prediction*10)/10)
}
