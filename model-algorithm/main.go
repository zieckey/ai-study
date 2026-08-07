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

// Sample 表示一条带标签的数据：特征用于预测，标签是真实结果。
type Sample struct {
	Features []float64
	Label    float64
}

// LinearModel 保存多元线性回归训练得到的权重和偏置。
type LinearModel struct {
	Weights []float64
	Bias    float64
}

// Predict 根据 y = w₁x₁ + ... + wₙxₙ + b 计算预测值。
func (m LinearModel) Predict(features []float64) (float64, error) {
	if len(features) != len(m.Weights) {
		return 0, fmt.Errorf("expected %d features, got %d", len(m.Weights), len(features))
	}

	prediction := m.Bias
	for i, feature := range features {
		prediction += m.Weights[i] * feature
	}
	return prediction, nil
}

// trainLinearModel 使用普通最小二乘法训练多元线性回归模型。
func trainLinearModel(samples []Sample) (LinearModel, error) {
	if len(samples) == 0 {
		return LinearModel{}, errors.New("at least one sample is required")
	}

	featureCount := len(samples[0].Features)
	if featureCount == 0 {
		return LinearModel{}, errors.New("each sample must have at least one feature")
	}
	if len(samples) < featureCount+1 {
		return LinearModel{}, fmt.Errorf("need at least %d samples for %d features", featureCount+1, featureCount)
	}

	// 增加截距列后，参数数量等于特征数量加上偏置 b。
	parameterCount := featureCount + 1
	matrix := make([][]float64, parameterCount)
	vector := make([]float64, parameterCount)
	for i := range matrix {
		matrix[i] = make([]float64, parameterCount)
	}

	for sampleIndex, sample := range samples {
		if len(sample.Features) != featureCount {
			return LinearModel{}, fmt.Errorf("sample %d has %d features, expected %d", sampleIndex, len(sample.Features), featureCount)
		}

		// 首列固定为 1，使偏置也能作为参数向量的一部分参与计算。
		row := make([]float64, parameterCount)
		row[0] = 1
		copy(row[1:], sample.Features)
		// 累积正规方程 (XᵀX)θ = Xᵀy 的左右两侧。
		for i := range row {
			vector[i] += row[i] * sample.Label
			for j := range row {
				matrix[i][j] += row[i] * row[j]
			}
		}
	}

	parameters, err := solveLinearSystem(matrix, vector)
	if err != nil {
		return LinearModel{}, err
	}

	return LinearModel{
		Bias:    parameters[0],
		Weights: parameters[1:],
	}, nil
}

// solveLinearSystem 用带部分主元选择的高斯消元法求解 A×x=b。
func solveLinearSystem(matrix [][]float64, vector []float64) ([]float64, error) {
	n := len(matrix)
	if n == 0 || len(vector) != n {
		return nil, errors.New("matrix and vector dimensions do not match")
	}

	augmented := make([][]float64, n)
	for i := range matrix {
		if len(matrix[i]) != n {
			return nil, errors.New("matrix must be square")
		}
		augmented[i] = make([]float64, n+1)
		copy(augmented[i], matrix[i])
		augmented[i][n] = vector[i]
	}

	for column := 0; column < n; column++ {
		// 选择当前列绝对值最大的主元，降低浮点计算误差。
		pivot := column
		for row := column + 1; row < n; row++ {
			if math.Abs(augmented[row][column]) > math.Abs(augmented[pivot][column]) {
				pivot = row
			}
		}
		if math.Abs(augmented[pivot][column]) < 1e-12 {
			return nil, errors.New("features produce a singular matrix")
		}

		augmented[column], augmented[pivot] = augmented[pivot], augmented[column]
		// 将主元下方的元素消为 0，得到上三角矩阵。
		for row := column + 1; row < n; row++ {
			factor := augmented[row][column] / augmented[column][column]
			for currentColumn := column; currentColumn <= n; currentColumn++ {
				augmented[row][currentColumn] -= factor * augmented[column][currentColumn]
			}
		}
	}

	// 从最后一行向上回代，依次得到每个未知参数。
	solution := make([]float64, n)
	for row := n - 1; row >= 0; row-- {
		sum := augmented[row][n]
		for column := row + 1; column < n; column++ {
			sum -= augmented[row][column] * solution[column]
		}
		solution[row] = sum / augmented[row][row]
	}

	return solution, nil
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

// Metrics 保存测试集上的两种预测误差指标。
type Metrics struct {
	MAE  float64
	RMSE float64
}

// evaluate 只使用测试样本计算平均绝对误差（MAE）和均方根误差（RMSE）。
func evaluate(model LinearModel, samples []Sample) (Metrics, error) {
	if len(samples) == 0 {
		return Metrics{}, errors.New("at least one test sample is required")
	}

	var absoluteError, squaredError float64
	for _, sample := range samples {
		prediction, err := model.Predict(sample.Features)
		if err != nil {
			return Metrics{}, err
		}
		// 残差为预测值减真实值；正值代表预测偏高。
		errorValue := prediction - sample.Label
		absoluteError += math.Abs(errorValue)
		squaredError += errorValue * errorValue
	}

	sampleCount := float64(len(samples))
	return Metrics{
		MAE:  absoluteError / sampleCount,
		RMSE: math.Sqrt(squaredError / sampleCount),
	}, nil
}

func main() {
	// 对照示例：算法根据预先写好的规则执行。
	fmt.Println("1. Algorithm: binary search")
	numbers := []int{3, 8, 12, 19, 24, 31, 42}
	target := 24
	index := binarySearch(numbers, target)
	fmt.Printf("Find %d in %v: index = %d\n\n", target, numbers, index)

	fmt.Println("2. Model: multifeature linear regression")
	featureNames := []string{"study hours", "attendance rate", "previous score"}
	// 每条样本包含学习时长、出勤率、之前成绩，以及对应的真实考试分数。
	samples := []Sample{
		{Features: []float64{1, 0.62, 52}, Label: 57},
		{Features: []float64{2, 0.68, 57}, Label: 63},
		{Features: []float64{2.5, 0.74, 61}, Label: 70},
		{Features: []float64{3, 0.71, 65}, Label: 74},
		{Features: []float64{3.5, 0.79, 68}, Label: 80},
		{Features: []float64{4, 0.83, 72}, Label: 85},
		{Features: []float64{4.5, 0.88, 76}, Label: 91},
		{Features: []float64{5, 0.91, 80}, Label: 96},
		{Features: []float64{5.5, 0.95, 83}, Label: 100},
		{Features: []float64{6, 0.97, 87}, Label: 105},
		{Features: []float64{6.5, 0.92, 90}, Label: 107},
		{Features: []float64{7, 0.98, 93}, Label: 114},
	}

	// 80% 用于训练，剩余 20% 留作从未参与训练的测试数据。
	trainSamples, testSamples, err := splitSamples(samples, 0.8, 42)
	if err != nil {
		panic(err)
	}
	// 仅用训练集学习参数，避免测试数据泄漏到模型中。
	model, err := trainLinearModel(trainSamples)
	if err != nil {
		panic(err)
	}
	// 只在测试集上计算误差，用于观察模型面对未知样本时的表现。
	metrics, err := evaluate(model, testSamples)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Features: %v\n", featureNames)
	fmt.Printf("Train samples: %d, test samples: %d\n", len(trainSamples), len(testSamples))
	fmt.Printf("Learned model: score = %.2f", model.Bias)
	for i, weight := range model.Weights {
		fmt.Printf(" + %.2f * %s", weight, featureNames[i])
	}
	fmt.Println()

	fmt.Println("Test predictions:")
	for _, sample := range testSamples {
		prediction, _ := model.Predict(sample.Features)
		fmt.Printf("  features=%v, predicted=%.1f, actual=%.1f, residual=%.1f\n", sample.Features, prediction, sample.Label, prediction-sample.Label)
	}
	fmt.Printf("Test MAE: %.2f\n", metrics.MAE)
	fmt.Printf("Test RMSE: %.2f\n", metrics.RMSE)
}
