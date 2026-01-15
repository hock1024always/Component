package main

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Tensor 张量结构
type Tensor struct {
	Data   []float64
	Shape  []int
	Grad   []float64
	RequiresGrad bool
}

// NewTensor 创建新张量
func NewTensor(data []float64, shape []int) *Tensor {
	return &Tensor{
		Data:   data,
		Shape:  shape,
		Grad:   make([]float64, len(data)),
		RequiresGrad: false,
	}
}

// Add 张量加法
func (t *Tensor) Add(other *Tensor) *Tensor {
	if len(t.Data) != len(other.Data) {
		panic("张量维度不匹配")
	}

	result := make([]float64, len(t.Data))
	for i := range t.Data {
		result[i] = t.Data[i] + other.Data[i]
	}

	return NewTensor(result, t.Shape)
}

// Mul 张量乘法
func (t *Tensor) Mul(other *Tensor) *Tensor {
	if len(t.Data) != len(other.Data) {
		panic("张量维度不匹配")
	}

	result := make([]float64, len(t.Data))
	for i := range t.Data {
		result[i] = t.Data[i] * other.Data[i]
	}

	return NewTensor(result, t.Shape)
}

// MatMul 矩阵乘法
func (t *Tensor) MatMul(other *Tensor) *Tensor {
	if len(t.Shape) != 2 || len(other.Shape) != 2 {
		panic("矩阵乘法需要二维张量")
	}
	if t.Shape[1] != other.Shape[0] {
		panic("矩阵维度不匹配")
	}

	rows := t.Shape[0]
	cols := other.Shape[1]
	inner := t.Shape[1]

	result := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			sum := 0.0
			for k := 0; k < inner; k++ {
				sum += t.Data[i*inner+k] * other.Data[k*cols+j]
			}
			result[i*cols+j] = sum
		}
	}

	return NewTensor(result, []int{rows, cols})
}

// Sum 求和
func (t *Tensor) Sum() float64 {
	sum := 0.0
	for _, v := range t.Data {
		sum += v
	}
	return sum
}

// Mean 求平均值
func (t *Tensor) Mean() float64 {
	return t.Sum() / float64(len(t.Data))
}

// Layer 神经网络层接口
type Layer interface {
	Forward(input *Tensor) *Tensor
	Backward(grad *Tensor) *Tensor
	GetParameters() []*Tensor
}

// Linear 全连接层
type Linear struct {
	Weight *Tensor
	Bias   *Tensor
	Input  *Tensor
}

// NewLinear 创建全连接层
func NewLinear(inFeatures, outFeatures int) *Linear {
	// Xavier初始化
	scale := math.Sqrt(2.0 / float64(inFeatures))

	weightData := make([]float64, inFeatures*outFeatures)
	for i := range weightData {
		weightData[i] = rand.NormFloat64() * scale
	}

	biasData := make([]float64, outFeatures)
	for i := range biasData {
		biasData[i] = 0.0
	}

	return &Linear{
		Weight: NewTensor(weightData, []int{inFeatures, outFeatures}),
		Bias:   NewTensor(biasData, []int{outFeatures}),
	}
}

// Forward 前向传播
func (l *Linear) Forward(input *Tensor) *Tensor {
	l.Input = input
	// y = x * W + b
	return input.MatMul(l.Weight).Add(l.Bias)
}

// Backward 反向传播
func (l *Linear) Backward(grad *Tensor) *Tensor {
	// dL/dx = dL/dy * W^T
	weightT := transpose(l.Weight)
	dx := grad.MatMul(weightT)

	// dL/dW = x^T * dL/dy
	inputT := transpose(l.Input)
	dW := inputT.MatMul(grad)

	// dL/db = sum(dL/dy, axis=0)
	db := make([]float64, len(l.Bias.Data))
	for i := 0; i < grad.Shape[0]; i++ {
		for j := 0; j < grad.Shape[1]; j++ {
			db[j] += grad.Data[i*grad.Shape[1]+j]
		}
	}

	// 更新梯度
	copy(l.Weight.Grad, dW.Data)
	copy(l.Bias.Grad, db)

	return dx
}

// GetParameters 获取参数
func (l *Linear) GetParameters() []*Tensor {
	return []*Tensor{l.Weight, l.Bias}
}

// ReLU 激活函数层
type ReLU struct {
	Input *Tensor
}

// NewReLU 创建ReLU层
func NewReLU() *ReLU {
	return &ReLU{}
}

// Forward 前向传播
func (r *ReLU) Forward(input *Tensor) *Tensor {
	r.Input = input
	result := make([]float64, len(input.Data))
	for i, v := range input.Data {
		if v > 0 {
			result[i] = v
		} else {
			result[i] = 0
		}
	}
	return NewTensor(result, input.Shape)
}

// Backward 反向传播
func (r *ReLU) Backward(grad *Tensor) *Tensor {
	result := make([]float64, len(grad.Data))
	for i, v := range r.Input.Data {
		if v > 0 {
			result[i] = grad.Data[i]
		} else {
			result[i] = 0
		}
	}
	return NewTensor(result, grad.Shape)
}

// GetParameters 获取参数
func (r *ReLU) GetParameters() []*Tensor {
	return []*Tensor{}
}

// MSELoss 均方误差损失函数
type MSELoss struct{}

// NewMSELoss 创建MSE损失函数
func NewMSELoss() *MSELoss {
	return &MSELoss{}
}

// Forward 前向传播
func (m *MSELoss) Forward(pred, target *Tensor) *Tensor {
	if len(pred.Data) != len(target.Data) {
		panic("预测值和目标值维度不匹配")
	}

	result := make([]float64, len(pred.Data))
	for i := range pred.Data {
		diff := pred.Data[i] - target.Data[i]
		result[i] = diff * diff
	}

	return NewTensor(result, pred.Shape)
}

// Backward 反向传播
func (m *MSELoss) Backward(pred, target *Tensor) *Tensor {
	result := make([]float64, len(pred.Data))
	for i := range pred.Data {
		result[i] = 2 * (pred.Data[i] - target.Data[i])
	}
	return NewTensor(result, pred.Shape)
}

// NeuralNetwork 神经网络
type NeuralNetwork struct {
	Layers []Layer
	Loss   *MSELoss
}

// NewNeuralNetwork 创建神经网络
func NewNeuralNetwork() *NeuralNetwork {
	return &NeuralNetwork{
		Layers: make([]Layer, 0),
		Loss:   NewMSELoss(),
	}
}

// AddLayer 添加层
func (nn *NeuralNetwork) AddLayer(layer Layer) {
	nn.Layers = append(nn.Layers, layer)
}

// Forward 前向传播
func (nn *NeuralNetwork) Forward(input *Tensor) *Tensor {
	output := input
	for _, layer := range nn.Layers {
		output = layer.Forward(output)
	}
	return output
}

// Backward 反向传播
func (nn *NeuralNetwork) Backward(pred, target *Tensor) {
	// 计算损失梯度
	lossGrad := nn.Loss.Backward(pred, target)

	// 反向传播
	grad := lossGrad
	for i := len(nn.Layers) - 1; i >= 0; i-- {
		grad = nn.Layers[i].Backward(grad)
	}
}

// GetParameters 获取所有参数
func (nn *NeuralNetwork) GetParameters() []*Tensor {
	var params []*Tensor
	for _, layer := range nn.Layers {
		params = append(params, layer.GetParameters()...)
	}
	return params
}

// Optimizer 优化器接口
type Optimizer interface {
	Step(params []*Tensor)
}

// SGD 随机梯度下降优化器
type SGD struct {
	LearningRate float64
}

// NewSGD 创建SGD优化器
func NewSGD(lr float64) *SGD {
	return &SGD{LearningRate: lr}
}

// Step 执行优化步骤
func (s *SGD) Step(params []*Tensor) {
	for _, param := range params {
		for i := range param.Data {
			param.Data[i] -= s.LearningRate * param.Grad[i]
			param.Grad[i] = 0 // 清空梯度
		}
	}
}

// Trainer 训练器
type Trainer struct {
	Network  *NeuralNetwork
	Optimizer Optimizer
	Epochs   int
}

// NewTrainer 创建训练器
func NewTrainer(network *NeuralNetwork, optimizer Optimizer, epochs int) *Trainer {
	return &Trainer{
		Network:  network,
		Optimizer: optimizer,
		Epochs:   epochs,
	}
}

// Train 训练网络
func (t *Trainer) Train(inputs, targets []*Tensor) {
	fmt.Printf("开始训练 %d 个epoch\n", t.Epochs)

	for epoch := 0; epoch < t.Epochs; epoch++ {
		totalLoss := 0.0

		for i, input := range inputs {
			// 前向传播
			pred := t.Network.Forward(input)

			// 计算损失
			loss := t.Network.Loss.Forward(pred, targets[i])
			totalLoss += loss.Sum()

			// 反向传播
			t.Network.Backward(pred, targets[i])
		}

		// 优化步骤
		t.Optimizer.Step(t.Network.GetParameters())

		if (epoch+1)%10 == 0 {
			fmt.Printf("Epoch %d, Loss: %.6f\n", epoch+1, totalLoss/float64(len(inputs)))
		}
	}

	fmt.Println("训练完成")
}

// Predict 预测
func (t *Trainer) Predict(input *Tensor) *Tensor {
	return t.Network.Forward(input)
}

// 辅助函数
func transpose(tensor *Tensor) *Tensor {
	if len(tensor.Shape) != 2 {
		panic("转置需要二维张量")
	}

	rows, cols := tensor.Shape[0], tensor.Shape[1]
	result := make([]float64, len(tensor.Data))

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			result[j*rows+i] = tensor.Data[i*cols+j]
		}
	}

	return NewTensor(result, []int{cols, rows})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// 创建神经网络
	network := NewNeuralNetwork()
	network.AddLayer(NewLinear(2, 4))  // 输入2维，隐藏层4维
	network.AddLayer(NewReLU())        // ReLU激活函数
	network.AddLayer(NewLinear(4, 1))  // 输出1维

	// 创建优化器
	optimizer := NewSGD(0.01)

	// 创建训练器
	trainer := NewTrainer(network, optimizer, 100)

	// 生成训练数据 (XOR问题)
	inputs := []*Tensor{
		NewTensor([]float64{0, 0}, []int{1, 2}),
		NewTensor([]float64{0, 1}, []int{1, 2}),
		NewTensor([]float64{1, 0}, []int{1, 2}),
		NewTensor([]float64{1, 1}, []int{1, 2}),
	}

	targets := []*Tensor{
		NewTensor([]float64{0}, []int{1, 1}),
		NewTensor([]float64{1}, []int{1, 1}),
		NewTensor([]float64{1}, []int{1, 1}),
		NewTensor([]float64{0}, []int{1, 1}),
	}

	// 训练网络
	trainer.Train(inputs, targets)

	// 测试预测
	fmt.Println("\n=== 预测结果 ===")
	for i, input := range inputs {
		pred := trainer.Predict(input)
		expected := targets[i].Data[0]
		actual := pred.Data[0]
		fmt.Printf("输入: %.0f,%.0f -> 期望: %.0f, 预测: %.4f\n",
			input.Data[0], input.Data[1], expected, actual)
	}

	// 显示网络参数
	fmt.Printf("\n=== 网络参数 ===\n")
	params := network.GetParameters()
	for i, param := range params {
		fmt.Printf("参数%d 形状: %v, 均值: %.4f\n", i, param.Shape, param.Mean())
	}
}