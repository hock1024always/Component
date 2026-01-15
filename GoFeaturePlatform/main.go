package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Feature 特征接口
type Feature interface {
	Name() string
	Value() interface{}
	Type() string
}

// NumericFeature 数值特征
type NumericFeature struct {
	name  string
	value float64
}

// NewNumericFeature 创建数值特征
func NewNumericFeature(name string, value float64) *NumericFeature {
	return &NumericFeature{name: name, value: value}
}

func (f *NumericFeature) Name() string       { return f.name }
func (f *NumericFeature) Value() interface{} { return f.value }
func (f *NumericFeature) Type() string       { return "numeric" }

// CategoricalFeature 类别特征
type CategoricalFeature struct {
	name  string
	value string
}

// NewCategoricalFeature 创建类别特征
func NewCategoricalFeature(name string, value string) *CategoricalFeature {
	return &CategoricalFeature{name: name, value: value}
}

func (f *CategoricalFeature) Name() string       { return f.name }
func (f *CategoricalFeature) Value() interface{} { return f.value }
func (f *CategoricalFeature) Type() string       { return "categorical" }

// VectorFeature 向量特征
type VectorFeature struct {
	name  string
	value []float64
}

// NewVectorFeature 创建向量特征
func NewVectorFeature(name string, value []float64) *VectorFeature {
	return &VectorFeature{name: name, value: value}
}

func (f *VectorFeature) Name() string       { return f.name }
func (f *VectorFeature) Value() interface{} { return f.value }
func (f *VectorFeature) Type() string       { return "vector" }

// FeatureSet 特征集合
type FeatureSet struct {
	features map[string]Feature
	userID   string
	timestamp time.Time
}

// NewFeatureSet 创建特征集合
func NewFeatureSet(userID string) *FeatureSet {
	return &FeatureSet{
		features:  make(map[string]Feature),
		userID:    userID,
		timestamp: time.Now(),
	}
}

// AddFeature 添加特征
func (fs *FeatureSet) AddFeature(feature Feature) {
	fs.features[feature.Name()] = feature
}

// GetFeature 获取特征
func (fs *FeatureSet) GetFeature(name string) (Feature, bool) {
	feature, exists := fs.features[name]
	return feature, exists
}

// GetAllFeatures 获取所有特征
func (fs *FeatureSet) GetAllFeatures() map[string]Feature {
	return fs.features
}

// FeatureTransformer 特征转换器接口
type FeatureTransformer interface {
	Transform(feature Feature) Feature
}

// StandardScaler 标准化转换器
type StandardScaler struct {
	mean float64
	std  float64
}

// NewStandardScaler 创建标准化转换器
func NewStandardScaler() *StandardScaler {
	return &StandardScaler{}
}

// Fit 拟合数据计算均值和标准差
func (ss *StandardScaler) Fit(features []*NumericFeature) {
	if len(features) == 0 {
		return
	}

	sum := 0.0
	for _, f := range features {
		sum += f.value
	}
	ss.mean = sum / float64(len(features))

	sumSq := 0.0
	for _, f := range features {
		diff := f.value - ss.mean
		sumSq += diff * diff
	}
	ss.std = math.Sqrt(sumSq / float64(len(features)))
}

// Transform 标准化转换
func (ss *StandardScaler) Transform(feature Feature) Feature {
	if numFeat, ok := feature.(*NumericFeature); ok {
		if ss.std == 0 {
			return NewNumericFeature(feature.Name(), 0)
		}
		normalized := (numFeat.value - ss.mean) / ss.std
		return NewNumericFeature(feature.Name(), normalized)
	}
	return feature
}

// OneHotEncoder 独热编码器
type OneHotEncoder struct {
	categories map[string][]string
}

// NewOneHotEncoder 创建独热编码器
func NewOneHotEncoder() *OneHotEncoder {
	return &OneHotEncoder{
		categories: make(map[string][]string),
	}
}

// Fit 拟合数据收集类别
func (ohe *OneHotEncoder) Fit(features []*CategoricalFeature) {
	categoryMap := make(map[string]map[string]bool)

	for _, f := range features {
		if categoryMap[f.name] == nil {
			categoryMap[f.name] = make(map[string]bool)
		}
		categoryMap[f.name][f.value] = true
	}

	for name, values := range categoryMap {
		var sortedValues []string
		for value := range values {
			sortedValues = append(sortedValues, value)
		}
		sort.Strings(sortedValues)
		ohe.categories[name] = sortedValues
	}
}

// Transform 独热编码转换
func (ohe *OneHotEncoder) Transform(feature Feature) Feature {
	if catFeat, ok := feature.(*CategoricalFeature); ok {
		categories, exists := ohe.categories[feature.Name()]
		if !exists {
			return feature
		}

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

// FeatureStore 特征存储
type FeatureStore struct {
	data   map[string]*FeatureSet
	mutex  sync.RWMutex
	ttl    time.Duration
}

// NewFeatureStore 创建特征存储
func NewFeatureStore(ttl time.Duration) *FeatureStore {
	store := &FeatureStore{
		data:  make(map[string]*FeatureSet),
		ttl:   ttl,
	}

	// 启动清理协程
	go store.cleanup()

	return store
}

// Store 存储特征集合
func (fs *FeatureStore) Store(featureSet *FeatureSet) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.data[featureSet.userID] = featureSet
}

// Get 获取特征集合
func (fs *FeatureStore) Get(userID string) (*FeatureSet, bool) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	featureSet, exists := fs.data[userID]
	return featureSet, exists
}

// Delete 删除特征集合
func (fs *FeatureStore) Delete(userID string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	delete(fs.data, userID)
}

// cleanup 清理过期数据
func (fs *FeatureStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

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

// FeatureEngine 特征计算引擎
type FeatureEngine struct {
	transformers []FeatureTransformer
	store        *FeatureStore
}

// NewFeatureEngine 创建特征计算引擎
func NewFeatureEngine(store *FeatureStore) *FeatureEngine {
	return &FeatureEngine{
		transformers: make([]FeatureTransformer, 0),
		store:        store,
	}
}

// AddTransformer 添加转换器
func (fe *FeatureEngine) AddTransformer(transformer FeatureTransformer) {
	fe.transformers = append(fe.transformers, transformer)
}

// ProcessFeatureSet 处理特征集合
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
				// 如果是新特征，添加它
				processed.features[transformed.Name()] = transformed
			} else {
				// 如果是原地转换，更新特征
				processed.features[name] = transformed
			}
		}
	}

	return processed
}

// FeaturePipeline 特征处理管道
type FeaturePipeline struct {
	engine *FeatureEngine
	store  *FeatureStore
}

// NewFeaturePipeline 创建特征处理管道
func NewFeaturePipeline() *FeaturePipeline {
	store := NewFeatureStore(1 * time.Hour)
	engine := NewFeatureEngine(store)

	return &FeaturePipeline{
		engine: engine,
		store:  store,
	}
}

// ProcessAndStore 处理并存储特征
func (fp *FeaturePipeline) ProcessAndStore(featureSet *FeatureSet) {
	// 处理特征
	processed := fp.engine.ProcessFeatureSet(featureSet)

	// 存储结果
	fp.store.Store(processed)

	fmt.Printf("处理并存储用户 %s 的特征，特征数量: %d\n",
		featureSet.userID, len(processed.features))
}

// GetProcessedFeatures 获取处理后的特征
func (fp *FeaturePipeline) GetProcessedFeatures(userID string) (*FeatureSet, bool) {
	return fp.store.Get(userID)
}

// BatchProcess 批量处理特征
func (fp *FeaturePipeline) BatchProcess(featureSets []*FeatureSet) {
	for _, featureSet := range featureSets {
		fp.ProcessAndStore(featureSet)
	}
}

// FeatureHasher 特征哈希器
type FeatureHasher struct {
	numFeatures int
}

// NewFeatureHasher 创建特征哈希器
func NewFeatureHasher(numFeatures int) *FeatureHasher {
	return &FeatureHasher{numFeatures: numFeatures}
}

// Hash 哈希特征为索引
func (fh *FeatureHasher) Hash(featureName string) int {
	h := fnv.New32a()
	h.Write([]byte(featureName))
	return int(h.Sum32()) % fh.numFeatures
}

// FeatureCombiner 特征组合器
type FeatureCombiner struct{}

// NewFeatureCombiner 创建特征组合器
func NewFeatureCombiner() *FeatureCombiner {
	return &FeatureCombiner{}
}

// CombineFeatures 组合特征
func (fc *FeatureCombiner) CombineFeatures(features []Feature) *VectorFeature {
	vector := make([]float64, 0)

	for _, feature := range features {
		switch f := feature.(type) {
		case *NumericFeature:
			vector = append(vector, f.value)
		case *VectorFeature:
			vector = append(vector, f.value...)
		case *CategoricalFeature:
			// 简单的字符串哈希转换为数值
			hash := fnv.New32a()
			hash.Write([]byte(f.value))
			vector = append(vector, float64(hash.Sum32()))
		}
	}

	return NewVectorFeature("combined_features", vector)
}

// FeatureSelector 特征选择器
type FeatureSelector struct {
	selectedFeatures map[string]bool
}

// NewFeatureSelector 创建特征选择器
func NewFeatureSelector(selectedFeatures []string) *FeatureSelector {
	selector := &FeatureSelector{
		selectedFeatures: make(map[string]bool),
	}

	for _, feature := range selectedFeatures {
		selector.selectedFeatures[feature] = true
	}

	return selector
}

// Select 选择特征
func (fs *FeatureSelector) Select(featureSet *FeatureSet) *FeatureSet {
	selected := NewFeatureSet(featureSet.userID)

	for name, feature := range featureSet.features {
		if fs.selectedFeatures[name] {
			selected.AddFeature(feature)
		}
	}

	return selected
}

func main() {
	// 创建特征处理管道
	pipeline := NewFeaturePipeline()

	// 添加特征转换器
	scaler := NewStandardScaler()
	encoder := NewOneHotEncoder()
	pipeline.engine.AddTransformer(scaler)
	pipeline.engine.AddTransformer(encoder)

	// 准备训练数据用于拟合转换器
	trainFeatures := []*NumericFeature{
		NewNumericFeature("age", 25),
		NewNumericFeature("age", 30),
		NewNumericFeature("age", 35),
		NewNumericFeature("income", 50000),
		NewNumericFeature("income", 60000),
		NewNumericFeature("income", 70000),
	}

	scaler.Fit(trainFeatures)

	trainCatFeatures := []*CategoricalFeature{
		NewCategoricalFeature("city", "北京"),
		NewCategoricalFeature("city", "上海"),
		NewCategoricalFeature("city", "深圳"),
		NewCategoricalFeature("gender", "男"),
		NewCategoricalFeature("gender", "女"),
	}

	encoder.Fit(trainCatFeatures)

	// 创建用户特征集合
	userFeatures := NewFeatureSet("user123")

	// 添加各种类型的特征
	userFeatures.AddFeature(NewNumericFeature("age", 28))
	userFeatures.AddFeature(NewNumericFeature("income", 55000))
	userFeatures.AddFeature(NewCategoricalFeature("city", "北京"))
	userFeatures.AddFeature(NewCategoricalFeature("gender", "男"))
	userFeatures.AddFeature(NewVectorFeature("interests", []float64{0.8, 0.6, 0.3, 0.9}))

	// 处理并存储特征
	pipeline.ProcessAndStore(userFeatures)

	// 获取处理后的特征
	processed, exists := pipeline.GetProcessedFeatures("user123")
	if exists {
		fmt.Println("\n=== 处理后的特征 ===")
		for name, feature := range processed.GetAllFeatures() {
			fmt.Printf("%s (%s): %v\n", name, feature.Type(), feature.Value())
		}
	}

	// 演示特征选择器
	selector := NewFeatureSelector([]string{"age", "city_onehot"})
	selected := selector.Select(processed)
	fmt.Println("\n=== 选择的特征 ===")
	for name, feature := range selected.GetAllFeatures() {
		fmt.Printf("%s (%s): %v\n", name, feature.Type(), feature.Value())
	}

	// 演示特征组合器
	combiner := NewFeatureCombiner()
	featuresToCombine := []Feature{
		NewNumericFeature("score", 0.85),
		NewVectorFeature("embedding", []float64{0.1, 0.2, 0.3}),
	}
	combined := combiner.CombineFeatures(featuresToCombine)
	fmt.Printf("\n=== 组合特征 ===\n")
	fmt.Printf("%s: %v\n", combined.Name(), combined.Value())

	// 演示特征哈希器
	hasher := NewFeatureHasher(100)
	hashValue := hasher.Hash("user_age")
	fmt.Printf("\n=== 特征哈希 ===\n")
	fmt.Printf("user_age 哈希值: %d\n", hashValue)

	// 批量处理演示
	fmt.Println("\n=== 批量处理演示 ===")
	batchFeatures := []*FeatureSet{
		NewFeatureSet("user456"),
		NewFeatureSet("user789"),
	}

	// 为批量用户添加特征
	for i, fs := range batchFeatures {
		fs.AddFeature(NewNumericFeature("age", float64(20+i*5)))
		fs.AddFeature(NewCategoricalFeature("city", []string{"北京", "上海", "深圳"}[i%3]))
	}

	pipeline.BatchProcess(batchFeatures)

	// 显示统计信息
	stats := make(map[string]int)
	for _, fs := range batchFeatures {
		if processed, exists := pipeline.GetProcessedFeatures(fs.userID); exists {
			stats["processed_users"]++
			stats["total_features"] += len(processed.GetAllFeatures())
		}
	}

	fmt.Printf("\n=== 统计信息 ===\n")
	fmt.Printf("处理用户数: %d\n", stats["processed_users"])
	fmt.Printf("总特征数: %d\n", stats["total_features"])
}