package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
)

// binarySearch 在升序数组中查找目标值，找到时返回下标，否则返回 -1。
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

// Sample 表示一条带标签的数据；Label 为 0 表示未违约，1 表示违约。
type Sample struct {
	Features []float64
	Label    float64
}

// LogisticModel 保存逻辑回归参数及训练集的特征标准化参数。
type LogisticModel struct {
	Weights []float64
	Bias    float64
	Means   []float64
	Scales  []float64
}

// PredictProbability 返回特征对应的违约概率，范围为 [0, 1]。
func (m LogisticModel) PredictProbability(features []float64) (float64, error) {
	if len(features) != len(m.Weights) {
		return 0, fmt.Errorf("expected %d features, got %d", len(m.Weights), len(features))
	}

	z := m.Bias
	for i, feature := range features {
		if !isFinite(feature) {
			return 0, fmt.Errorf("feature %d must be finite", i)
		}
		z += m.Weights[i] * standardize(feature, m.Means[i], m.Scales[i])
	}
	return sigmoid(z), nil
}

// PredictClass 依据阈值将违约概率转换为 0（未违约）或 1（违约）。
func (m LogisticModel) PredictClass(features []float64, threshold float64) (int, error) {
	if threshold <= 0 || threshold >= 1 {
		return 0, errors.New("threshold must be between 0 and 1")
	}

	probability, err := m.PredictProbability(features)
	if err != nil {
		return 0, err
	}
	if probability >= threshold {
		return 1, nil
	}
	return 0, nil
}

// trainLogisticModel 使用训练集标准化和全批量梯度下降训练逻辑回归模型。
func trainLogisticModel(samples []Sample, learningRate float64, iterations int) (LogisticModel, error) {
	if learningRate <= 0 || !isFinite(learningRate) {
		return LogisticModel{}, errors.New("learning rate must be a positive finite number")
	}
	if iterations <= 0 {
		return LogisticModel{}, errors.New("iterations must be positive")
	}

	featureCount, err := validateSamples(samples)
	if err != nil {
		return LogisticModel{}, err
	}

	means, scales := featureStatistics(samples, featureCount)
	weights := make([]float64, featureCount)
	bias := 0.0
	sampleCount := float64(len(samples))

	for iteration := 0; iteration < iterations; iteration++ {
		weightGradients := make([]float64, featureCount)
		biasGradient := 0.0

		for _, sample := range samples {
			z := bias
			for i, feature := range sample.Features {
				z += weights[i] * standardize(feature, means[i], scales[i])
			}

			prediction := sigmoid(z)
			errorValue := prediction - sample.Label
			biasGradient += errorValue
			for i, feature := range sample.Features {
				weightGradients[i] += errorValue * standardize(feature, means[i], scales[i])
			}
		}

		bias -= learningRate * biasGradient / sampleCount
		for i := range weights {
			weights[i] -= learningRate * weightGradients[i] / sampleCount
		}
	}

	return LogisticModel{
		Weights: weights,
		Bias:    bias,
		Means:   means,
		Scales:  scales,
	}, nil
}

func validateSamples(samples []Sample) (int, error) {
	if len(samples) == 0 {
		return 0, errors.New("at least one sample is required")
	}

	featureCount := len(samples[0].Features)
	if featureCount == 0 {
		return 0, errors.New("each sample must have at least one feature")
	}

	for sampleIndex, sample := range samples {
		if len(sample.Features) != featureCount {
			return 0, fmt.Errorf("sample %d has %d features, expected %d", sampleIndex, len(sample.Features), featureCount)
		}
		if sample.Label != 0 && sample.Label != 1 {
			return 0, fmt.Errorf("sample %d label must be 0 or 1", sampleIndex)
		}
		for featureIndex, feature := range sample.Features {
			if !isFinite(feature) {
				return 0, fmt.Errorf("sample %d feature %d must be finite", sampleIndex, featureIndex)
			}
		}
	}

	return featureCount, nil
}

// featureStatistics 只从训练集计算均值和标准差，避免测试集信息泄漏。
func featureStatistics(samples []Sample, featureCount int) ([]float64, []float64) {
	means := make([]float64, featureCount)
	for _, sample := range samples {
		for i, feature := range sample.Features {
			means[i] += feature
		}
	}
	for i := range means {
		means[i] /= float64(len(samples))
	}

	variances := make([]float64, featureCount)
	for _, sample := range samples {
		for i, feature := range sample.Features {
			delta := feature - means[i]
			variances[i] += delta * delta
		}
	}

	scales := make([]float64, featureCount)
	for i, variance := range variances {
		scales[i] = math.Sqrt(variance / float64(len(samples)))
	}
	return means, scales
}

func standardize(value, mean, scale float64) float64 {
	if scale == 0 {
		return 0
	}
	return (value - mean) / scale
}

// sigmoid 将任意实数映射到 [0, 1]，并避免指数计算溢出。
func sigmoid(value float64) float64 {
	if value >= 0 {
		return 1 / (1 + math.Exp(-value))
	}
	exp := math.Exp(value)
	return exp / (1 + exp)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

// splitSamples 以固定种子打乱样本后，按比例切分为训练集和测试集。
func splitSamples(samples []Sample, trainRatio float64, seed int64) ([]Sample, []Sample, error) {
	if trainRatio <= 0 || trainRatio >= 1 {
		return nil, nil, errors.New("train ratio must be between 0 and 1")
	}
	if len(samples) < 2 {
		return nil, nil, errors.New("at least two samples are required for a train/test split")
	}

	// 复制后再打乱，保证不改变调用方持有的原始样本顺序。
	shuffled := append([]Sample(nil), samples...)
	// 固定种子使每次切分结果可复现，便于学习和测试。
	random := rand.New(rand.NewSource(seed))
	random.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	trainSize := int(math.Floor(float64(len(shuffled)) * trainRatio))
	if trainSize == 0 || trainSize == len(shuffled) {
		return nil, nil, errors.New("train ratio must leave at least one sample in each split")
	}

	return shuffled[:trainSize], shuffled[trainSize:], nil
}

// ClassificationMetrics 保存二分类预测的混淆矩阵和常用评估指标。
type ClassificationMetrics struct {
	TruePositive  int
	TrueNegative  int
	FalsePositive int
	FalseNegative int
	Accuracy      float64
	Precision     float64
	Recall        float64
}

// evaluate 只使用测试样本计算混淆矩阵、准确率、精确率和召回率。
func evaluate(model LogisticModel, samples []Sample, threshold float64) (ClassificationMetrics, error) {
	if len(samples) == 0 {
		return ClassificationMetrics{}, errors.New("at least one test sample is required")
	}
	if threshold <= 0 || threshold >= 1 {
		return ClassificationMetrics{}, errors.New("threshold must be between 0 and 1")
	}

	metrics := ClassificationMetrics{}
	for _, sample := range samples {
		predictedClass, err := model.PredictClass(sample.Features, threshold)
		if err != nil {
			return ClassificationMetrics{}, err
		}

		switch {
		case predictedClass == 1 && sample.Label == 1:
			metrics.TruePositive++
		case predictedClass == 0 && sample.Label == 0:
			metrics.TrueNegative++
		case predictedClass == 1 && sample.Label == 0:
			metrics.FalsePositive++
		default:
			metrics.FalseNegative++
		}
	}

	total := float64(len(samples))
	metrics.Accuracy = float64(metrics.TruePositive+metrics.TrueNegative) / total
	positivePredictions := metrics.TruePositive + metrics.FalsePositive
	if positivePredictions > 0 {
		metrics.Precision = float64(metrics.TruePositive) / float64(positivePredictions)
	}
	actualPositives := metrics.TruePositive + metrics.FalseNegative
	if actualPositives > 0 {
		metrics.Recall = float64(metrics.TruePositive) / float64(actualPositives)
	}
	return metrics, nil
}

func className(class int) string {
	if class == 1 {
		return "default"
	}
	return "non-default"
}

func main() {
	fmt.Println("1. Algorithm: binary search")
	numbers := []int{3, 8, 12, 19, 24, 31, 42}
	target := 24
	index := binarySearch(numbers, target)
	fmt.Printf("Find %d in %v: index = %d\n\n", target, numbers, index)

	fmt.Println("2. Model: logistic regression for default probability")
	fmt.Println("Educational synthetic example only; not for credit decisions.")
	featureNames := []string{"debt-to-income ratio", "prior late payments", "on-time repayment ratio"}
	samples := []Sample{
		{Features: []float64{0.12, 0, 0.99}, Label: 0},
		{Features: []float64{0.18, 0, 0.96}, Label: 0},
		{Features: []float64{0.22, 1, 0.93}, Label: 0},
		{Features: []float64{0.28, 0, 0.95}, Label: 0},
		{Features: []float64{0.32, 1, 0.88}, Label: 0},
		{Features: []float64{0.38, 1, 0.83}, Label: 0},
		{Features: []float64{0.44, 2, 0.78}, Label: 1},
		{Features: []float64{0.51, 2, 0.72}, Label: 1},
		{Features: []float64{0.58, 3, 0.66}, Label: 1},
		{Features: []float64{0.63, 3, 0.59}, Label: 1},
		{Features: []float64{0.71, 4, 0.54}, Label: 1},
		{Features: []float64{0.78, 5, 0.48}, Label: 1},
		{Features: []float64{0.83, 5, 0.42}, Label: 1},
		{Features: []float64{0.89, 6, 0.36}, Label: 1},
		{Features: []float64{0.35, 0, 0.97}, Label: 0},
	}

	trainSamples, testSamples, err := splitSamples(samples, 0.8, 42)
	if err != nil {
		panic(err)
	}
	model, err := trainLogisticModel(trainSamples, 0.1, 2000)
	if err != nil {
		panic(err)
	}

	const threshold = 0.5
	metrics, err := evaluate(model, testSamples, threshold)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Features: %v\n", featureNames)
	fmt.Printf("Train samples: %d, test samples: %d\n", len(trainSamples), len(testSamples))
	fmt.Printf("Model: PD = sigmoid(%.2f", model.Bias)
	for i, weight := range model.Weights {
		fmt.Printf(" + %.2f * standardized(%s)", weight, featureNames[i])
	}
	fmt.Println(")")

	fmt.Printf("Test predictions (threshold %.1f):\n", threshold)
	for _, sample := range testSamples {
		probability, _ := model.PredictProbability(sample.Features)
		predictedClass, _ := model.PredictClass(sample.Features, threshold)
		fmt.Printf("  features=%v, PD=%.3f, actual=%s, predicted=%s\n", sample.Features, probability, className(int(sample.Label)), className(predictedClass))
	}
	fmt.Printf("Confusion matrix: TP=%d, TN=%d, FP=%d, FN=%d\n", metrics.TruePositive, metrics.TrueNegative, metrics.FalsePositive, metrics.FalseNegative)
	fmt.Printf("Test accuracy: %.2f, precision: %.2f, recall: %.2f\n", metrics.Accuracy, metrics.Precision, metrics.Recall)
}
