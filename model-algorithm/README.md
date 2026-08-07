# 算法与模型：Golang 入门示例

本项目通过一个可直接运行的 Go 程序，同时展示两个基础概念：

1. **算法（Algorithm）**：二分查找。
2. **模型（Model）**：一元线性回归。

代码位于 [main.go](main.go)。它们看起来都在“根据输入得到输出”，但规则的来源完全不同：算法的规则由程序员写死；模型的参数从样本数据中计算出来。

## 运行程序

需要本机安装 Go。进入项目根目录后执行：

```bash
go run main.go
```

预期输出：

```text
1. Algorithm: binary search
Find 24 in [3 8 12 19 24 31 42]: index = 4

2. Model: learn a line from data
Learned model: score = 7.00 * hours + 47.80
Predicted score after 6 hours: 89.8
```

## 程序结构

```text
main.go
├── binarySearch       二分查找算法
├── LinearModel        线性模型的数据结构
├── LinearModel.Predict 使用模型预测
├── trainLinearModel   根据样本训练模型
└── main               准备数据、调用并打印结果
```

---

## 第一部分：算法——二分查找

### 什么是算法

算法是**为解决问题预先定义的一系列明确步骤**。同样的输入、同样的步骤，会得到同样的输出。

程序中的 `binarySearch` 函数用于在一个**已排序**的整数数组中查找目标值：

```go
func binarySearch(sorted []int, target int) int
```

- `sorted`：必须是从小到大排列的数组。
- `target`：要查找的数字。
- 返回值：若找到，返回它的下标；若未找到，返回 `-1`。

Go 的数组下标从 `0` 开始。例如：

```text
下标:  0  1  2  3  4  5  6
数据:  3  8 12 19 24 31 42
```

所以 `24` 位于下标 `4`。

### 代码逐段讲解

```go
left, right := 0, len(sorted)-1
```

`left` 和 `right` 表示当前需要搜索的闭区间边界。

- `left = 0`：从第一个元素开始。
- `right = len(sorted)-1`：最后一个有效下标。
- 对 `{3, 8, 12, 19, 24, 31, 42}` 而言，初始区间是 `[0, 6]`。

```go
for left <= right {
```

只要区间仍然存在元素，就继续查找。当 `left > right` 时，说明所有可能位置都排除了，目标不存在。

```go
middle := left + (right-left)/2
```

计算中间位置。这里没有直接写 `(left + right) / 2`，是为了避免在非常大的整数范围内 `left + right` 溢出；这是常见的稳妥写法。

```go
switch {
case sorted[middle] == target:
    return middle
case sorted[middle] < target:
    left = middle + 1
default:
    right = middle - 1
}
```

每一轮只需要比较中间值：

1. **相等**：已找到，立即返回下标。
2. **中间值小于目标**：目标只能位于右半部分，因此把 `left` 移到 `middle + 1`。
3. **中间值大于目标**：目标只能位于左半部分，因此把 `right` 移到 `middle - 1`。

### 查找 24 的执行过程

输入数组为 `[3, 8, 12, 19, 24, 31, 42]`，目标是 `24`。

| 轮次 | 搜索区间 | `middle` | 中间值 | 判断 | 下一步 |
| --- | --- | ---: | ---: | --- | --- |
| 1 | `[0, 6]` | 3 | 19 | `19 < 24` | 搜索右半部分 `[4, 6]` |
| 2 | `[4, 6]` | 5 | 31 | `31 > 24` | 搜索左半部分 `[4, 4]` |
| 3 | `[4, 4]` | 4 | 24 | `24 == 24` | 返回下标 `4` |

二分查找每次排除约一半元素，因此时间复杂度为 **O(log n)**。例如，一百万个有序元素通常只需约 20 次比较即可定位。

### 二分查找的重要前提

数组必须有序。若数组无序，例如 `[19, 3, 24, 8]`，比较中间值后就无法可靠地判断目标位于左边还是右边，二分查找会失效。

这体现了算法的特点：它有明确规则，也有明确的使用前提。

---

## 第二部分：模型——一元线性回归

### 什么是模型

模型表示一种从数据中得到的规律。程序并没有手动规定“每多学习一小时加几分”，而是给出已有样本，让训练过程计算出这样的关系。

示例使用的样本是：

| 学习时长（小时） | 考试分数 |
| ---: | ---: |
| 1 | 55 |
| 2 | 62 |
| 3 | 68 |
| 4 | 76 |
| 5 | 83 |

其中：

- 学习时长是**特征**（feature），用 `x` 表示。
- 考试分数是**标签**或目标值（label/target），用 `y` 表示。
- 这些一一对应的样本称为**有标签数据**。

一元线性回归假设两者的关系能近似用一条直线表达：

```text
y = w × x + b
```

- `x`：输入特征，例如学习时长。
- `y`：预测目标，例如考试分数。
- `w`：权重（weight），也可理解为斜率；表示 `x` 每增加 1 单位时，`y` 的平均变化量。
- `b`：偏置（bias），也可理解为截距；直线与 `y` 轴的交点。

在本程序训练后的结果中：

```text
score = 7.00 × hours + 47.80
```

它的含义是：在这些样本所呈现的整体趋势中，学习时长每增加 1 小时，预测分数平均提高约 7 分。

### `LinearModel`：保存训练得到的参数

```go
type LinearModel struct {
    Weight float64
    Bias   float64
}
```

`LinearModel` 是 Go 的结构体。它保存一条直线的两个参数：

- `Weight` 保存 `w`。
- `Bias` 保存 `b`。
- 使用 `float64`，因为训练和预测结果通常包含小数。

训练完成后，模型不需要再保存所有原始样本也能预测；只要保留 `w` 与 `b` 即可。这是一个很小的模型实例。

### `Predict`：使用模型做预测

```go
func (m LinearModel) Predict(x float64) float64 {
    return m.Weight*x + m.Bias
}
```

这是定义在 `LinearModel` 上的方法。

- `m` 是当前模型。
- 接收一个新的特征值 `x`。
- 按照模型公式 `w × x + b` 得到预测值。

训练得到 `w = 7.00`、`b = 47.80` 后，预测学习 6 小时的分数：

```text
y = 7.00 × 6 + 47.80
  = 89.80
```

因此程序打印 `89.8`。

### `trainLinearModel`：从数据训练模型

```go
func trainLinearModel(features, labels []float64) LinearModel
```

这个函数的职责是根据样本计算出最合适的 `Weight` 和 `Bias`。该过程称为**训练**或**拟合**。

#### 1. 校验输入

```go
if len(features) != len(labels) || len(features) == 0 {
    panic("features and labels must have the same non-zero length")
}
```

每个特征必须对应一个标签：

```text
features: [1, 2, 3]
labels:   [55, 62, 68]
```

如果两者长度不同，就无法知道每个学习时长对应哪个分数；空样本也无法计算规律。因此在不满足条件时，函数通过 `panic` 终止程序并提示错误。

#### 2. 计算平均值

```go
var meanX, meanY float64
for i := range features {
    meanX += features[i]
    meanY += labels[i]
}
meanX /= float64(len(features))
meanY /= float64(len(labels))
```

这段代码先累加全部 `x` 和 `y`，再除以样本数，得到：

```text
meanX = (1 + 2 + 3 + 4 + 5) / 5 = 3
meanY = (55 + 62 + 68 + 76 + 83) / 5 = 68.8
```

`float64(len(features))` 将整数样本数转换为浮点数，确保这里执行的是小数除法。

#### 3. 计算权重（斜率）

```go
var numerator, denominator float64
for i := range features {
    deltaX := features[i] - meanX
    numerator += deltaX * (labels[i] - meanY)
    denominator += deltaX * deltaX
}
weight := numerator / denominator
```

这段代码实现了最小二乘法的一元线性回归公式：

```text
w = Σ((xi - x̄) × (yi - ȳ)) / Σ((xi - x̄)²)
```

其中 `x̄`、`ȳ` 分别是 `x` 与 `y` 的平均值。

- `numerator` 衡量学习时长和分数是否一起增减。
- `denominator` 衡量学习时长本身的变化程度。
- 两者相除得到最适合样本整体趋势的斜率。

对本例数据计算后，`w = 7.0`。

```go
if denominator == 0 {
    panic("features must not all be the same")
}
```

如果所有学习时长完全相同，例如 `[2, 2, 2]`，则无法通过“学习时长的变化”推导它对分数的影响，此时分母为零，不能除法。

#### 4. 计算偏置（截距）

```go
Bias: meanY - weight*meanX,
```

偏置的公式为：

```text
b = ȳ - w × x̄
```

这是因为拟合的最佳直线会经过样本的平均点 `(x̄, ȳ)`。

对于本例：

```text
b = 68.8 - 7.0 × 3
  = 47.8
```

最后函数返回：

```go
LinearModel{
    Weight: 7.0,
    Bias:   47.8,
}
```

---

## `main`：将算法和模型串起来

`main` 是 Go 程序的入口，按顺序执行两个示例。

### 调用算法

```go
numbers := []int{3, 8, 12, 19, 24, 31, 42}
target := 24
index := binarySearch(numbers, target)
```

这里直接将固定规则 `binarySearch` 应用于输入数据。没有训练步骤，也没有待学习的参数。

### 训练并使用模型

```go
studyHours := []float64{1, 2, 3, 4, 5}
testScores := []float64{55, 62, 68, 76, 83}
model := trainLinearModel(studyHours, testScores)
```

先提供训练样本，再调用 `trainLinearModel` 得到模型参数。

```go
futureHours := 6.0
prediction := model.Predict(futureHours)
```

然后输入一个训练集里没有出现过的新值 `6.0`，调用 `Predict` 进行预测。这一步通常称为**推理**（inference）或**预测**（prediction）。

```go
math.Round(prediction*10) / 10
```

浮点数计算可能显示出较多小数位。此表达式通过“先乘以 10、四舍五入、再除以 10”的方式，让结果保留一位小数；它只改变展示精度，不改变模型本身的参数。

---

## 算法与模型的对比

| 维度 | 二分查找算法 | 一元线性回归模型 |
| --- | --- | --- |
| 规则来源 | 程序员明确编写 | 由训练数据计算参数 |
| 训练步骤 | 不需要 | 需要 `trainLinearModel` |
| 使用时的输入 | 有序数组和目标值 | 新特征值，例如学习时长 |
| 输出 | 确定的下标或 `-1` | 连续数值的预测结果 |
| 结果是否受数据质量影响 | 主要取决于输入是否有序 | 强烈依赖训练样本的数量、代表性和误差 |
| 本例函数 | `binarySearch` | `trainLinearModel` 和 `Predict` |

可以将两者理解为：

- **算法**：人先写好“怎么做”。
- **模型**：人先规定模型形式，例如直线；具体的参数由数据学习得到。

实际机器学习系统同样离不开算法：训练线性回归时使用的最小二乘法本身就是一个算法。区别在于，算法的输出之一是模型参数，而这些参数再用于处理新的输入。

## 这个示例的边界

本示例用于理解基本概念，不能直接用于真实的考试成绩预测：

- 样本只有 5 条，太少，无法代表真实人群。
- 分数受基础水平、题目难度、睡眠、课程质量等多种因素影响，不能只由学习时长决定。
- 线性关系只是一个假设；真实关系可能不是直线。
- 当前代码用全部数据训练后立刻预测，没有使用训练集和测试集评估模型误差。

后续可以尝试：增加更多特征、划分训练集与测试集、计算预测误差，或使用梯度下降理解更通用的训练算法。
