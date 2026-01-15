package main

import (
	"testing"
	"time"
)

func TestNumericFeature(t *testing.T) {
	feature := NewNumericFeature("age", 25.5)

	if feature.Name() != "age" {
		t.Errorf("期望名称'age'，实际'%s'", feature.Name())
	}

	if feature.Value() != 25.5 {
		t.Errorf("期望值25.5，实际%v", feature.Value())
	}

	if feature.Type() != "numeric" {
		t.Errorf("期望类型'numeric'，实际'%s'", feature.Type())
	}
}

func TestCategoricalFeature(t *testing.T) {
	feature := NewCategoricalFeature("city", "北京")

	if feature.Name() != "city" {
		t.Errorf("期望名称'city'，实际'%s'", feature.Name())
	}

	if feature.Value() != "北京" {
		t.Errorf("期望值'北京'，实际'%v'", feature.Value())
	}

	if feature.Type() != "categorical" {
		t.Errorf("期望类型'categorical'，实际'%s'", feature.Type())
	}
}

func TestFeatureSet(t *testing.T) {
	fs := NewFeatureSet("user123")

	// 添加特征
	numFeat := NewNumericFeature("age", 30)
	catFeat := NewCategoricalFeature("city", "上海")

	fs.AddFeature(numFeat)
	fs.AddFeature(catFeat)

	// 获取特征
	retrieved, exists := fs.GetFeature("age")
	if !exists {
		t.Error("期望找到特征'age'")
	}

	if retrieved.Value() != 30.0 {
		t.Errorf("期望值30，实际%v", retrieved.Value())
	}

	// 获取所有特征
	all := fs.GetAllFeatures()
	if len(all) != 2 {
		t.Errorf("期望2个特征，实际%d个", len(all))
	}
}

func TestStandardScaler(t *testing.T) {
	scaler := NewStandardScaler()

	// 拟合数据
	features := []*NumericFeature{
		NewNumericFeature("test", 1.0),
		NewNumericFeature("test", 2.0),
		NewNumericFeature("test", 3.0),
	}
	scaler.Fit(features)

	// 转换特征
	input := NewNumericFeature("test", 2.0)
	output := scaler.Transform(input)

	if output.Type() != "numeric" {
		t.Errorf("期望输出类型numeric，实际%s", output.Type())
	}

	// 对于均值2，标准差1的数据，2.0应该标准化为0
	normalized := output.Value().(float64)
	if normalized < -0.1 || normalized > 0.1 {
		t.Errorf("期望标准化值为0附近，实际%v", normalized)
	}
}

func TestOneHotEncoder(t *testing.T) {
	encoder := NewOneHotEncoder()

	// 拟合数据
	features := []*CategoricalFeature{
		NewCategoricalFeature("color", "red"),
		NewCategoricalFeature("color", "blue"),
		NewCategoricalFeature("color", "red"),
	}
	encoder.Fit(features)

	// 转换特征
	input := NewCategoricalFeature("color", "red")
	output := encoder.Transform(input)

	if output.Type() != "vector" {
		t.Errorf("期望输出类型vector，实际%s", output.Type())
	}

	vector := output.Value().([]float64)
	// 应该有2个类别 (blue, red)，red是第1个(索引1)
	if len(vector) != 2 {
		t.Errorf("期望向量长度2，实际%d", len(vector))
	}

	if vector[1] != 1.0 {
		t.Errorf("期望red编码为[?, 1.0]，实际%v", vector)
	}
}

func TestFeatureStore(t *testing.T) {
	store := NewFeatureStore(1 * time.Hour)

	// 创建特征集合
	fs := NewFeatureSet("user123")
	fs.AddFeature(NewNumericFeature("age", 25))

	// 存储
	store.Store(fs)

	// 获取
	retrieved, exists := store.Get("user123")
	if !exists {
		t.Error("期望找到存储的特征集合")
	}

	if retrieved.userID != "user123" {
		t.Errorf("期望用户ID'user123'，实际'%s'", retrieved.userID)
	}

	// 删除
	store.Delete("user123")
	_, exists = store.Get("user123")
	if exists {
		t.Error("期望删除后找不到特征集合")
	}
}

func TestFeatureEngine(t *testing.T) {
	store := NewFeatureStore(1 * time.Hour)
	engine := NewFeatureEngine(store)

	// 添加转换器
	scaler := NewStandardScaler()
	features := []*NumericFeature{NewNumericFeature("test", 1.0)}
	scaler.Fit(features)
	engine.AddTransformer(scaler)

	// 处理特征集合
	input := NewFeatureSet("user123")
	input.AddFeature(NewNumericFeature("test", 1.0))

	output := engine.ProcessFeatureSet(input)

	if len(output.features) == 0 {
		t.Error("期望输出包含特征")
	}
}

func TestFeaturePipeline(t *testing.T) {
	pipeline := NewFeaturePipeline()

	// 创建特征集合
	fs := NewFeatureSet("user123")
	fs.AddFeature(NewNumericFeature("age", 30))

	// 处理并存储
	pipeline.ProcessAndStore(fs)

	// 获取处理结果
	result, exists := pipeline.GetProcessedFeatures("user123")
	if !exists {
		t.Error("期望找到处理后的特征")
	}

	if result.userID != "user123" {
		t.Errorf("期望用户ID'user123'，实际'%s'", result.userID)
	}
}

func TestFeatureHasher(t *testing.T) {
	hasher := NewFeatureHasher(100)

	hash1 := hasher.Hash("feature1")
	hash2 := hasher.Hash("feature2")

	if hash1 < 0 || hash1 >= 100 {
		t.Errorf("哈希值应该在[0,100)范围内，实际%d", hash1)
	}

	if hash2 < 0 || hash2 >= 100 {
		t.Errorf("哈希值应该在[0,100)范围内，实际%d", hash2)
	}

	// 相同输入应该产生相同哈希
	hash1Again := hasher.Hash("feature1")
	if hash1 != hash1Again {
		t.Error("相同输入应该产生相同哈希")
	}
}

func TestFeatureCombiner(t *testing.T) {
	combiner := NewFeatureCombiner()

	features := []Feature{
		NewNumericFeature("score", 0.85),
		NewVectorFeature("embedding", []float64{0.1, 0.2}),
	}

	combined := combiner.CombineFeatures(features)

	if combined.Type() != "vector" {
		t.Errorf("期望组合结果类型vector，实际%s", combined.Type())
	}

	vector := combined.Value().([]float64)
	if len(vector) != 3 { // 1 + 2 = 3
		t.Errorf("期望向量长度3，实际%d", len(vector))
	}

	if vector[0] != 0.85 {
		t.Errorf("期望第一个元素为0.85，实际%v", vector[0])
	}
}

func TestFeatureSelector(t *testing.T) {
	selector := NewFeatureSelector([]string{"age", "city"})

	// 创建特征集合
	fs := NewFeatureSet("user123")
	fs.AddFeature(NewNumericFeature("age", 30))
	fs.AddFeature(NewNumericFeature("income", 50000))
	fs.AddFeature(NewCategoricalFeature("city", "北京"))

	// 选择特征
	selected := selector.Select(fs)

	features := selected.GetAllFeatures()
	if len(features) != 2 {
		t.Errorf("期望选择2个特征，实际%d个", len(features))
	}

	if _, exists := features["age"]; !exists {
		t.Error("期望包含'age'特征")
	}

	if _, exists := features["city"]; !exists {
		t.Error("期望包含'city'特征")
	}

	if _, exists := features["income"]; exists {
		t.Error("不期望包含'income'特征")
	}
}