package main

import (
	"math"
	"reflect"
	"testing"
)

func TestTrainLogisticModelSeparatesClasses(t *testing.T) {
	samples := []Sample{
		{Features: []float64{-3}, Label: 0},
		{Features: []float64{-2}, Label: 0},
		{Features: []float64{-1}, Label: 0},
		{Features: []float64{1}, Label: 1},
		{Features: []float64{2}, Label: 1},
		{Features: []float64{3}, Label: 1},
	}

	model, err := trainLogisticModel(samples, 0.1, 2000)
	if err != nil {
		t.Fatalf("trainLogisticModel returned an error: %v", err)
	}

	lowRiskProbability, err := model.PredictProbability([]float64{-2})
	if err != nil {
		t.Fatalf("PredictProbability returned an error: %v", err)
	}
	highRiskProbability, err := model.PredictProbability([]float64{2})
	if err != nil {
		t.Fatalf("PredictProbability returned an error: %v", err)
	}
	if lowRiskProbability >= highRiskProbability {
		t.Fatalf("low-risk probability %.6f is not lower than high-risk probability %.6f", lowRiskProbability, highRiskProbability)
	}
	if lowRiskProbability < 0 || lowRiskProbability > 1 || highRiskProbability < 0 || highRiskProbability > 1 {
		t.Fatal("probabilities must be in [0, 1]")
	}

	predictedClass, err := model.PredictClass([]float64{2}, 0.5)
	if err != nil {
		t.Fatalf("PredictClass returned an error: %v", err)
	}
	if predictedClass != 1 {
		t.Fatalf("predicted class = %d, want 1", predictedClass)
	}
}

func TestTrainLogisticModelUsesTrainingStatistics(t *testing.T) {
	samples := []Sample{
		{Features: []float64{10}, Label: 0},
		{Features: []float64{20}, Label: 0},
		{Features: []float64{30}, Label: 1},
		{Features: []float64{40}, Label: 1},
	}

	model, err := trainLogisticModel(samples, 0.1, 10)
	if err != nil {
		t.Fatalf("trainLogisticModel returned an error: %v", err)
	}

	assertClose(t, model.Means[0], 25)
	assertClose(t, model.Scales[0], math.Sqrt(125))
}

func TestTrainLogisticModelRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name         string
		samples      []Sample
		learningRate float64
		iterations   int
	}{
		{name: "empty", samples: nil, learningRate: 0.1, iterations: 1},
		{name: "no features", samples: []Sample{{Features: nil, Label: 0}}, learningRate: 0.1, iterations: 1},
		{name: "ragged samples", samples: []Sample{{Features: []float64{1}, Label: 0}, {Features: []float64{2, 3}, Label: 1}}, learningRate: 0.1, iterations: 1},
		{name: "invalid label", samples: []Sample{{Features: []float64{1}, Label: 2}}, learningRate: 0.1, iterations: 1},
		{name: "not finite feature", samples: []Sample{{Features: []float64{math.NaN()}, Label: 0}}, learningRate: 0.1, iterations: 1},
		{name: "invalid learning rate", samples: []Sample{{Features: []float64{1}, Label: 0}}, learningRate: 0, iterations: 1},
		{name: "invalid iterations", samples: []Sample{{Features: []float64{1}, Label: 0}}, learningRate: 0.1, iterations: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := trainLogisticModel(test.samples, test.learningRate, test.iterations); err == nil {
				t.Fatal("trainLogisticModel returned nil error")
			}
		})
	}
}

func TestSigmoidHandlesExtremeValues(t *testing.T) {
	if probability := sigmoid(1000); probability != 1 {
		t.Fatalf("sigmoid(1000) = %v, want 1", probability)
	}
	if probability := sigmoid(-1000); probability != 0 {
		t.Fatalf("sigmoid(-1000) = %v, want 0", probability)
	}
}

func TestPredictRejectsInvalidFeaturesAndThreshold(t *testing.T) {
	model := LogisticModel{
		Weights: []float64{1},
		Means:   []float64{0},
		Scales:  []float64{1},
	}

	if _, err := model.PredictProbability([]float64{1, 2}); err == nil {
		t.Fatal("PredictProbability accepted wrong feature count")
	}
	if _, err := model.PredictProbability([]float64{math.Inf(1)}); err == nil {
		t.Fatal("PredictProbability accepted infinite feature")
	}
	if _, err := model.PredictClass([]float64{1}, 0); err == nil {
		t.Fatal("PredictClass accepted invalid threshold")
	}
}

func TestSplitSamplesIsDeterministicAndDoesNotMutateInput(t *testing.T) {
	samples := []Sample{
		{Features: []float64{1}, Label: 0},
		{Features: []float64{2}, Label: 1},
		{Features: []float64{3}, Label: 0},
		{Features: []float64{4}, Label: 1},
		{Features: []float64{5}, Label: 0},
	}
	original := append([]Sample(nil), samples...)

	train, test, err := splitSamples(samples, 0.8, 42)
	if err != nil {
		t.Fatalf("splitSamples returned an error: %v", err)
	}
	otherTrain, otherTest, err := splitSamples(samples, 0.8, 42)
	if err != nil {
		t.Fatalf("splitSamples returned an error: %v", err)
	}

	if len(train) != 4 || len(test) != 1 {
		t.Fatalf("split sizes = %d and %d, want 4 and 1", len(train), len(test))
	}
	if !reflect.DeepEqual(train, otherTrain) || !reflect.DeepEqual(test, otherTest) {
		t.Fatal("splitSamples was not deterministic")
	}
	if !reflect.DeepEqual(samples, original) {
		t.Fatal("splitSamples mutated input")
	}

	seen := make(map[float64]bool)
	for _, sample := range train {
		seen[sample.Features[0]] = true
	}
	for _, sample := range test {
		if seen[sample.Features[0]] {
			t.Fatalf("sample %v appears in both splits", sample.Features)
		}
	}
}

func TestSplitSamplesRejectsInvalidInput(t *testing.T) {
	samples := []Sample{{Features: []float64{1}, Label: 0}, {Features: []float64{2}, Label: 1}}
	for _, ratio := range []float64{0, 1, -0.1, 1.1} {
		if _, _, err := splitSamples(samples, ratio, 42); err == nil {
			t.Fatalf("splitSamples(%v) returned nil error", ratio)
		}
	}
	if _, _, err := splitSamples(samples[:1], 0.8, 42); err == nil {
		t.Fatal("splitSamples accepted one sample")
	}
}

func TestEvaluateCalculatesClassificationMetrics(t *testing.T) {
	model := LogisticModel{
		Weights: []float64{1},
		Means:   []float64{0},
		Scales:  []float64{1},
	}
	samples := []Sample{
		{Features: []float64{2}, Label: 1},
		{Features: []float64{-2}, Label: 0},
		{Features: []float64{2}, Label: 0},
		{Features: []float64{-2}, Label: 1},
	}

	metrics, err := evaluate(model, samples, 0.5)
	if err != nil {
		t.Fatalf("evaluate returned an error: %v", err)
	}

	if metrics.TruePositive != 1 || metrics.TrueNegative != 1 || metrics.FalsePositive != 1 || metrics.FalseNegative != 1 {
		t.Fatalf("unexpected confusion matrix: %+v", metrics)
	}
	assertClose(t, metrics.Accuracy, 0.5)
	assertClose(t, metrics.Precision, 0.5)
	assertClose(t, metrics.Recall, 0.5)
}

func TestEvaluateRejectsInvalidInput(t *testing.T) {
	model := LogisticModel{Weights: []float64{1}, Means: []float64{0}, Scales: []float64{1}}
	if _, err := evaluate(model, nil, 0.5); err == nil {
		t.Fatal("evaluate accepted empty samples")
	}
	if _, err := evaluate(model, []Sample{{Features: []float64{1}, Label: 1}}, 1); err == nil {
		t.Fatal("evaluate accepted invalid threshold")
	}
}

func assertClose(t *testing.T, actual, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("got %.12f, want %.12f", actual, expected)
	}
}
