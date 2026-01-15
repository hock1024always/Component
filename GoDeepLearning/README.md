# 深度学习训练框架

这是一个用Go语言实现的简易深度学习训练框架，支持自动微分、神经网络构建和训练。

## 架构设计

### 核心组件

1. **Tensor** - 张量结构
   - Data: 存储张量数据
   - Shape: 张量形状
   - Grad: 梯度数据
   - RequiresGrad: 是否需要梯度

2. **Layer** - 神经网络层接口
   - Forward(): 前向传播
   - Backward(): 反向传播
   - GetParameters(): 获取参数

3. **NeuralNetwork** - 神经网络
   - Layers: 网络层列表
   - Loss: 损失函数

4. **Optimizer** - 优化器接口
   - Step(): 执行优化步骤

5. **Trainer** - 训练器
   - 管理训练过程

## 核心特性

1. **自动微分**: 支持反向传播自动计算梯度
2. **模块化设计**: 层、损失函数、优化器分离
3. **张量运算**: 基础的张量加法、乘法、矩阵乘法
4. **多种层类型**: 全连接层、ReLU激活函数
5. **训练框架**: 完整的训练循环和预测功能

## 张量操作详解

### Tensor 结构体详解

```go
type Tensor struct {
    Data   []float64  // 存储实际数据
    Shape  []int      // 张量形状，如 [2,3] 表示2x3矩阵
    Grad   []float64  // 梯度数据，与Data一一对应
    RequiresGrad bool // 是否需要计算梯度
}
```

### 张量运算

**Add 方法**: 张量加法
```go
func (t *Tensor) Add(other *Tensor) *Tensor {
    result := make([]float64, len(t.Data))
    for i := range t.Data {
        result[i] = t.Data[i] + other.Data[i]
    }
    return NewTensor(result, t.Shape)
}
```

**MatMul 方法**: 矩阵乘法
```go
func (t *Tensor) MatMul(other *Tensor) *Tensor {
    // 实现标准的矩阵乘法算法
    rows, cols := t.Shape[0], other.Shape[1]
    inner := t.Shape[1]

    for i := 0; i < rows; i++ {
        for j := 0; j < cols; j++ {
            sum := 0.0
            for k := 0; k < inner; k++ {
                sum += t.Data[i*inner+k] * other.Data[k*cols+j]
            }
            result[i*cols+j] = sum
        }
    }
}
```

## 神经网络层详解

### Linear 全连接层

```go
type Linear struct {
    Weight *Tensor  // 权重矩阵 (in_features, out_features)
    Bias   *Tensor  // 偏置向量 (out_features,)
    Input  *Tensor  // 保存输入用于反向传播
}
```

**Forward 方法**: 前向传播
```go
func (l *Linear) Forward(input *Tensor) *Tensor {
    l.Input = input                              // 保存输入
    return input.MatMul(l.Weight).Add(l.Bias)    // y = xW + b
}
```

**Backward 方法**: 反向传播
```go
func (l *Linear) Backward(grad *Tensor) *Tensor {
    // dL/dx = dL/dy * W^T
    weightT := transpose(l.Weight)
    dx := grad.MatMul(weightT)

    // dL/dW = x^T * dL/dy
    inputT := transpose(l.Input)
    dW := inputT.MatMul(grad)

    // dL/db = sum(dL/dy, axis=0)
    db := sumOverAxis(grad, 0)

    // 保存梯度
    copy(l.Weight.Grad, dW.Data)
    copy(l.Bias.Grad, db.Data)

    return dx
}
```

### ReLU 激活函数层

```go
type ReLU struct {
    Input *Tensor  // 保存输入用于反向传播
}
```

**Forward 方法**:
```go
func (r *ReLU) Forward(input *Tensor) *Tensor {
    r.Input = input
    result := make([]float64, len(input.Data))
    for i, v := range input.Data {
        result[i] = math.Max(0, v)  // max(0, x)
    }
    return NewTensor(result, input.Shape)
}
```

**Backward 方法**:
```go
func (r *ReLU) Backward(grad *Tensor) *Tensor {
    result := make([]float64, len(grad.Data))
    for i, v := range r.Input.Data {
        if v > 0 {
            result[i] = grad.Data[i]  // 导数为1
        } else {
            result[i] = 0            // 导数为0
        }
    }
    return NewTensor(result, grad.Shape)
}
```

## 训练框架详解

### NeuralNetwork 神经网络

```go
type NeuralNetwork struct {
    Layers []Layer   // 网络层列表
    Loss   *MSELoss  // 损失函数
}
```

**Forward 方法**: 网络前向传播
```go
func (nn *NeuralNetwork) Forward(input *Tensor) *Tensor {
    output := input
    for _, layer := range nn.Layers {
        output = layer.Forward(output)  // 逐层前向传播
    }
    return output
}
```

**Backward 方法**: 网络反向传播
```go
func (nn *NeuralNetwork) Backward(pred, target *Tensor) {
    lossGrad := nn.Loss.Backward(pred, target)  // 计算损失梯度

    grad := lossGrad
    for i := len(nn.Layers) - 1; i >= 0; i-- {  // 从后往前反向传播
        grad = nn.Layers[i].Backward(grad)
    }
}
```

### SGD 优化器

```go
type SGD struct {
    LearningRate float64  // 学习率
}
```

**Step 方法**: 参数更新
```go
func (s *SGD) Step(params []*Tensor) {
    for _, param := range params {
        for i := range param.Data {
            param.Data[i] -= s.LearningRate * param.Grad[i]  // θ = θ - lr * ∇θ
            param.Grad[i] = 0  // 清空梯度
        }
    }
}
```

### Trainer 训练器

```go
type Trainer struct {
    Network  *NeuralNetwork
    Optimizer Optimizer
    Epochs   int
}
```

**Train 方法**: 训练循环
```go
func (t *Trainer) Train(inputs, targets []*Tensor) {
    for epoch := 0; epoch < t.Epochs; epoch++ {
        totalLoss := 0.0

        for i, input := range inputs {
            pred := t.Network.Forward(input)      // 前向传播
            loss := t.Network.Loss.Forward(pred, targets[i])  // 计算损失
            totalLoss += loss.Sum()

            t.Network.Backward(pred, targets[i])  // 反向传播
        }

        t.Optimizer.Step(t.Network.GetParameters())  // 优化步骤

        // 每10个epoch打印一次损失
        if (epoch+1)%10 == 0 {
            avgLoss := totalLoss / float64(len(inputs))
            fmt.Printf("Epoch %d, Loss: %.6f\n", epoch+1, avgLoss)
        }
    }
}
```

## 使用方法

### 1. 编译运行
```bash
go run main.go
```

程序会：
- 创建一个简单的神经网络 (2-4-1)
- 训练XOR问题
- 显示训练过程和最终预测结果
- 输出网络参数统计

### 2. 运行测试
```bash
go test -v
```

## 测试覆盖

- `TestTensorAdd`: 张量加法测试
- `TestTensorMul`: 张量乘法测试
- `TestTensorMatMul`: 矩阵乘法测试
- `TestLinearLayer`: 全连接层测试
- `TestReLU`: ReLU激活函数测试
- `TestMSELoss`: 均方误差损失测试
- `TestNeuralNetwork`: 神经网络测试
- `TestSGDOptimizer`: SGD优化器测试
- `TestTrainer`: 训练器测试

## 扩展思路

1. **更多层类型**: 卷积层、池化层、Dropout、BatchNorm
2. **更多优化器**: Adam、RMSprop、学习率调度
3. **更多损失函数**: 交叉熵、L1损失
4. **数据加载器**: 支持批量数据加载和预处理
5. **GPU支持**: 使用CUDA加速计算
6. **模型保存加载**: 支持模型序列化和反序列化
7. **分布式训练**: 支持多机多卡训练