package main

import (
	"math"
	"reflect"
	"testing"
)

func TestTrainLinearModelRecoversKnownParameters(t *testing.T) {
	samples := []Sample{
		{Features: []float64{0, 0}, Label: 5},
		{Features: []float64{1, 0}, Label: 7},
		{Features: []float64{0, 1}, Label: 2},
		{Features: []float64{2, 1}, Label: 6},
		{Features: []float64{1, 3}, Label: -2},
	}

	model, err := trainLinearModel(samples)
	if err != nil {
		t.Fatalf("trainLinearModel returned an error: %v", err)
	}

	assertClose(t, model.Bias, 5)
	assertClose(t, model.Weights[0], 2)
	assertClose(t, model.Weights[1], -3)

	prediction, err := model.Predict([]float64{4, 2})
	if err != nil {
		t.Fatalf("Predict returned an error: %v", err)
	}
	assertClose(t, prediction, 7)
}

func TestTrainLinearModelRejectsInvalidSamples(t *testing.T) {
	tests := []struct {
		name    string
		samples []Sample
	}{
		{name: "empty", samples: nil},
		{name: "no features", samples: []Sample{{Features: nil, Label: 1}}},
		{name: "insufficient samples", samples: []Sample{{Features: []float64{1, 2}, Label: 3}, {Features: []float64{2, 3}, Label: 4}}},
		{name: "ragged samples", samples: []Sample{{Features: []float64{1}, Label: 2}, {Features: []float64{2, 3}, Label: 4}}},
		{name: "singular features", samples: []Sample{{Features: []float64{1}, Label: 2}, {Features: []float64{1}, Label: 3}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := trainLinearModel(test.samples); err == nil {
				t.Fatal("trainLinearModel returned nil error")
			}
		})
	}
}

func TestPredictRejectsWrongFeatureCount(t *testing.T) {
	model := LinearModel{Weights: []float64{1, 2}, Bias: 3}
	if _, err := model.Predict([]float64{1}); err == nil {
		t.Fatal("Predict returned nil error")
	}
}

func TestSplitSamplesIsDeterministicAndDoesNotMutateInput(t *testing.T) {
	samples := []Sample{
		{Features: []float64{1}, Label: 1},
		{Features: []float64{2}, Label: 2},
		{Features: []float64{3}, Label: 3},
		{Features: []float64{4}, Label: 4},
		{Features: []float64{5}, Label: 5},
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
		seen[sample.Label] = true
	}
	for _, sample := range test {
		if seen[sample.Label] {
			t.Fatalf("label %v appears in both splits", sample.Label)
		}
	}
}

func TestSplitSamplesRejectsInvalidInput(t *testing.T) {
	samples := []Sample{{Features: []float64{1}, Label: 1}, {Features: []float64{2}, Label: 2}}
	for _, ratio := range []float64{0, 1, -0.1, 1.1} {
		if _, _, err := splitSamples(samples, ratio, 42); err == nil {
			t.Fatalf("splitSamples(%v) returned nil error", ratio)
		}
	}
	if _, _, err := splitSamples(samples[:1], 0.8, 42); err == nil {
		t.Fatal("splitSamples accepted one sample")
	}
}

func TestEvaluateCalculatesMAEAndRMSE(t *testing.T) {
	model := LinearModel{Weights: []float64{1}}
	samples := []Sample{
		{Features: []float64{1}, Label: 0},
		{Features: []float64{2}, Label: 4},
	}

	metrics, err := evaluate(model, samples)
	if err != nil {
		t.Fatalf("evaluate returned an error: %v", err)
	}

	assertClose(t, metrics.MAE, 1.5)
	assertClose(t, metrics.RMSE, math.Sqrt(2.5))
}

func TestEvaluateRejectsEmptySamples(t *testing.T) {
	if _, err := evaluate(LinearModel{}, nil); err == nil {
		t.Fatal("evaluate returned nil error")
	}
}

func assertClose(t *testing.T, actual, expected float64) {
	t.Helper()
	if math.Abs(actual-expected) > 1e-9 {
		t.Fatalf("got %.12f, want %.12f", actual, expected)
	}
}
