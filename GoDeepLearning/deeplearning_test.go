package main

import (
	"testing"
)

func TestTensorAdd(t *testing.T) {
	t1 := NewTensor([]float64{1, 2, 3}, []int{3})
	t2 := NewTensor([]float64{4, 5, 6}, []int{3})

	result := t1.Add(t2)

	expected := []float64{5, 7, 9}
	for i, v := range result.Data {
		if v != expected[i] {
			t.Errorf("期望%v，实际%v", expected, result.Data)
			break
		}
	}
}

func TestTensorMul(t *testing.T) {
	t1 := NewTensor([]float64{1, 2, 3}, []int{3})
	t2 := NewTensor([]float64{2, 3, 4}, []int{3})

	result := t1.Mul(t2)

	expected := []float64{2, 6, 12}
	for i, v := range result.Data {
		if v != expected[i] {
			t.Errorf("期望%v，实际%v", expected, result.Data)
			break
		}
	}
}

func TestTensorMatMul(t *testing.T) {
	t1 := NewTensor([]float64{1, 2, 3, 4}, []int{2, 2}) // [[1,2],[3,4]]
	t2 := NewTensor([]float64{5, 6, 7, 8}, []int{2, 2}) // [[5,6],[7,8]]

	result := t1.MatMul(t2)

	// 期望结果: [[1*5+2*7, 1*6+2*8], [3*5+4*7, 3*6+4*8]] = [[19, 22], [43, 50]]
	expected := []float64{19, 22, 43, 50}
	for i, v := range result.Data {
		if v != expected[i] {
			t.Errorf("期望%v，实际%v", expected, result.Data)
			break
		}
	}
}

func TestLinearLayer(t *testing.T) {
	layer := NewLinear(2, 3)
	input := NewTensor([]float64{1, 2}, []int{1, 2})

	output := layer.Forward(input)

	// 检查输出形状
	if len(output.Data) != 3 {
		t.Errorf("期望输出长度3，实际%d", len(output.Data))
	}

	// 检查输出形状
	expectedShape := []int{1, 3}
	for i, v := range output.Shape {
		if v != expectedShape[i] {
			t.Errorf("期望形状%v，实际%v", expectedShape, output.Shape)
			break
		}
	}
}

func TestReLU(t *testing.T) {
	relu := NewReLU()
	input := NewTensor([]float64{-1, 0, 1, 2}, []int{4})

	output := relu.Forward(input)

	expected := []float64{0, 0, 1, 2}
	for i, v := range output.Data {
		if v != expected[i] {
			t.Errorf("期望%v，实际%v", expected, output.Data)
			break
		}
	}
}

func TestMSELoss(t *testing.T) {
	loss := NewMSELoss()
	pred := NewTensor([]float64{1, 2, 3}, []int{3})
	target := NewTensor([]float64{1, 2, 4}, []int{3})

	result := loss.Forward(pred, target)

	// MSE: ((1-1)^2 + (2-2)^2 + (3-4)^2) / 3 = (0 + 0 + 1) = 1
	expected := []float64{0, 0, 1}
	for i, v := range result.Data {
		if v != expected[i] {
			t.Errorf("期望%v，实际%v", expected, result.Data)
			break
		}
	}
}

func TestNeuralNetwork(t *testing.T) {
	network := NewNeuralNetwork()
	network.AddLayer(NewLinear(2, 2))
	network.AddLayer(NewReLU())

	input := NewTensor([]float64{1, 2}, []int{1, 2})
	output := network.Forward(input)

	// 检查输出不为空
	if len(output.Data) == 0 {
		t.Error("网络输出为空")
	}

	// 检查网络有参数
	params := network.GetParameters()
	if len(params) == 0 {
		t.Error("网络没有参数")
	}
}

func TestSGDOptimizer(t *testing.T) {
	// 创建一个简单的参数张量
	param := NewTensor([]float64{1.0, 2.0}, []int{2})
	param.Grad = []float64{0.1, 0.2} // 设置梯度
	param.RequiresGrad = true

	optimizer := NewSGD(0.1)
	optimizer.Step([]*Tensor{param})

	// 检查参数是否更新: 1.0 - 0.1*0.1 = 0.99, 2.0 - 0.1*0.2 = 1.98
	expected := []float64{0.99, 1.98}
	for i, v := range param.Data {
		if v != expected[i] {
			t.Errorf("期望%v，实际%v", expected, param.Data)
			break
		}
	}

	// 检查梯度是否清空
	for _, g := range param.Grad {
		if g != 0 {
			t.Error("梯度应该被清空")
			break
		}
	}
}

func TestTrainer(t *testing.T) {
	// 创建简单的网络
	network := NewNeuralNetwork()
	network.AddLayer(NewLinear(1, 1))

	optimizer := NewSGD(0.01)
	trainer := NewTrainer(network, optimizer, 1)

	// 简单的训练数据: y = 2x
	inputs := []*Tensor{
		NewTensor([]float64{1}, []int{1, 1}),
		NewTensor([]float64{2}, []int{1, 1}),
	}
	targets := []*Tensor{
		NewTensor([]float64{2}, []int{1, 1}),
		NewTensor([]float64{4}, []int{1, 1}),
	}

	// 训练
	trainer.Train(inputs, targets)

	// 检查训练后可以预测
	pred := trainer.Predict(inputs[0])
	if len(pred.Data) != 1 {
		t.Error("预测结果维度错误")
	}
}