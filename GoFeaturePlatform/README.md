# 统一特征计算平台

这是一个统一的特征计算平台，支持特征的创建、转换、存储和处理，为机器学习应用提供特征工程能力。

## 架构设计

### 核心组件

1. **Feature** - 特征接口
   - Name(): 特征名称
   - Value(): 特征值
   - Type(): 特征类型

2. **FeatureSet** - 特征集合
   - 管理用户的多个特征
   - 支持特征的增删改查

3. **FeatureTransformer** - 特征转换器接口
   - Transform(): 转换特征

4. **FeatureStore** - 特征存储
   - 存储特征集合
   - 支持TTL过期清理

5. **FeatureEngine** - 特征计算引擎
   - 管理转换器流水线
   - 处理特征集合

6. **FeaturePipeline** - 特征处理管道
   - 端到端特征处理流程

## 特征类型详解

### NumericFeature 数值特征

```go
type NumericFeature struct {
    name  string
    value float64
}
```

适用于连续数值，如年龄、收入等。

### CategoricalFeature 类别特征

```go
type CategoricalFeature struct {
    name  string
    value string
}
```

适用于离散类别，如城市、性别等。

### VectorFeature 向量特征

```go
type VectorFeature struct {
    name  string
    value []float64
}
```

适用于嵌入向量、独热编码等向量形式特征。

## 特征转换器详解

### StandardScaler 标准化转换器

```go
type StandardScaler struct {
    mean float64  // 均值
    std  float64  // 标准差
}
```

**Fit 方法**: 计算训练数据的均值和标准差
```go
func (ss *StandardScaler) Fit(features []*NumericFeature) {
    // 计算均值
    sum := 0.0
    for _, f := range features {
        sum += f.value
    }
    ss.mean = sum / float64(len(features))

    // 计算标准差
    sumSq := 0.0
    for _, f := range features {
        diff := f.value - ss.mean
        sumSq += diff * diff
    }
    ss.std = math.Sqrt(sumSq / float64(len(features)))
}
```

**Transform 方法**: 标准化转换
```go
func (ss *StandardScaler) Transform(feature Feature) Feature {
    if numFeat, ok := feature.(*NumericFeature); ok {
        normalized := (numFeat.value - ss.mean) / ss.std
        return NewNumericFeature(feature.Name(), normalized)
    }
    return feature
}
```

### OneHotEncoder 独热编码器

```go
type OneHotEncoder struct {
    categories map[string][]string  // 每个特征的类别列表
}
```

**Fit 方法**: 收集所有可能的类别值
```go
func (ohe *OneHotEncoder) Fit(features []*CategoricalFeature) {
    categoryMap := make(map[string]map[string]bool)

    for _, f := range features {
        if categoryMap[f.name] == nil {
            categoryMap[f.name] = make(map[string]bool)
        }
        categoryMap[f.name][f.value] = true
    }

    // 排序确保一致性
    for name, values := range categoryMap {
        var sortedValues []string
        for value := range values {
            sortedValues = append(sortedValues, value)
        }
        sort.Strings(sortedValues)
        ohe.categories[name] = sortedValues
    }
}
```

**Transform 方法**: 转换为独热向量
```go
func (ohe *OneHotEncoder) Transform(feature Feature) Feature {
    if catFeat, ok := feature.(*CategoricalFeature); ok {
        categories := ohe.categories[feature.Name()]
        vector := make([]float64, len(categories))

        for i, cat := range categories {
            if cat == catFeat.value {
                vector[i] = 1.0
                break
            }
        }

        return NewVectorFeature(feature.Name()+"_onehot", vector)
    }
    return feature
}
```

## 存储和缓存详解

### FeatureStore 特征存储

```go
type FeatureStore struct {
    data   map[string]*FeatureSet  // 用户ID -> 特征集合
    mutex  sync.RWMutex           // 读写锁
    ttl    time.Duration          // 过期时间
}
```

**Store 方法**: 存储特征集合
```go
func (fs *FeatureStore) Store(featureSet *FeatureSet) {
    fs.mutex.Lock()
    defer fs.mutex.Unlock()
    fs.data[featureSet.userID] = featureSet
}
```

**cleanup 方法**: 自动清理过期数据
```go
func (fs *FeatureStore) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)

    for range ticker.C {
        fs.mutex.Lock()
        for userID, featureSet := range fs.data {
            if time.Since(featureSet.timestamp) > fs.ttl {
                delete(fs.data, userID)
            }
        }
        fs.mutex.Unlock()
    }
}
```

## 特征处理管道详解

### FeatureEngine 特征计算引擎

```go
type FeatureEngine struct {
    transformers []FeatureTransformer  // 转换器列表
    store        *FeatureStore         // 特征存储
}
```

**ProcessFeatureSet 方法**: 处理特征集合
```go
func (fe *FeatureEngine) ProcessFeatureSet(featureSet *FeatureSet) *FeatureSet {
    processed := NewFeatureSet(featureSet.userID)

    // 复制原始特征
    for name, feature := range featureSet.features {
        processed.features[name] = feature
    }

    // 应用所有转换器
    for _, transformer := range fe.transformers {
        for name, feature := range processed.features {
            transformed := transformer.Transform(feature)
            if transformed.Name() != name {
                // 新特征
                processed.features[transformed.Name()] = transformed
            } else {
                // 原地转换
                processed.features[name] = transformed
            }
        }
    }

    return processed
}
```

### FeaturePipeline 特征处理管道

```go
type FeaturePipeline struct {
    engine *FeatureEngine
    store  *FeatureStore
}
```

**ProcessAndStore 方法**: 端到端处理流程
```go
func (fp *FeaturePipeline) ProcessAndStore(featureSet *FeatureSet) {
    processed := fp.engine.ProcessFeatureSet(featureSet)  // 处理
    fp.store.Store(processed)                             // 存储
}
```

## 高级特征处理详解

### FeatureHasher 特征哈希器

```go
type FeatureHasher struct {
    numFeatures int  // 哈希空间大小
}
```

**Hash 方法**: 将特征名称哈希为索引
```go
func (fh *FeatureHasher) Hash(featureName string) int {
    h := fnv.New32a()
    h.Write([]byte(featureName))
    return int(h.Sum32()) % fh.numFeatures
}
```

适用于大规模特征的降维处理。

### FeatureCombiner 特征组合器

```go
func (fc *FeatureCombiner) CombineFeatures(features []Feature) *VectorFeature {
    vector := make([]float64, 0)

    for _, feature := range features {
        switch f := feature.(type) {
        case *NumericFeature:
            vector = append(vector, f.value)
        case *VectorFeature:
            vector = append(vector, f.value...)
        case *CategoricalFeature:
            // 字符串哈希为数值
            hash := fnv.New32a()
            hash.Write([]byte(f.value))
            vector = append(vector, float64(hash.Sum32()))
        }
    }

    return NewVectorFeature("combined_features", vector)
}
```

将多个特征组合成一个向量特征。

### FeatureSelector 特征选择器

```go
type FeatureSelector struct {
    selectedFeatures map[string]bool  // 选中的特征
}
```

**Select 方法**: 选择指定特征
```go
func (fs *FeatureSelector) Select(featureSet *FeatureSet) *FeatureSet {
    selected := NewFeatureSet(featureSet.userID)

    for name, feature := range featureSet.features {
        if fs.selectedFeatures[name] {
            selected.AddFeature(feature)
        }
    }

    return selected
}
```

## 使用方法

### 1. 编译运行
```bash
go run main.go
```

程序会演示：
- 创建各种类型的特征
- 配置和使用特征转换器
- 处理和存储特征
- 特征选择、组合、哈希等高级功能

### 2. 运行测试
```bash
go test -v
```

## 测试覆盖

- `TestNumericFeature`: 数值特征测试
- `TestCategoricalFeature`: 类别特征测试
- `TestFeatureSet`: 特征集合测试
- `TestStandardScaler`: 标准化转换器测试
- `TestOneHotEncoder`: 独热编码器测试
- `TestFeatureStore`: 特征存储测试
- `TestFeatureEngine`: 特征引擎测试
- `TestFeaturePipeline`: 特征管道测试
- `TestFeatureHasher`: 特征哈希器测试
- `TestFeatureCombiner`: 特征组合器测试
- `TestFeatureSelector`: 特征选择器测试

## 扩展思路

1. **特征重要性**: 计算特征重要性评分
2. **特征交叉**: 自动生成特征交叉组合
3. **特征监控**: 特征质量和分布监控
4. **在线学习**: 支持在线特征更新
5. **特征版本管理**: 特征模式的版本控制
6. **分布式处理**: 支持大规模特征处理
7. **特征缓存**: 多级缓存优化性能