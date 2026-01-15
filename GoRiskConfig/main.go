package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// ConfigItem 配置项
type ConfigItem struct {
	Key         string      `json:"key"`
	Value       interface{} `json:"value"`
	Description string      `json:"description"`
	Version     int         `json:"version"`
	UpdatedAt   time.Time   `json:"updated_at"`
	UpdatedBy   string      `json:"updated_by"`
}

// ConfigGroup 配置组
type ConfigGroup struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Items       map[string]*ConfigItem `json:"items"`
	Version     int                    `json:"version"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RiskConfig 风控配置中心
type RiskConfig struct {
	groups     map[string]*ConfigGroup
	listeners  []ConfigListener
	mutex      sync.RWMutex
	version    int
	history    []*ConfigChange
	maxHistory int
}

// ConfigListener 配置监听器
type ConfigListener interface {
	OnConfigChange(groupName, key string, oldValue, newValue interface{})
}

// ConfigChange 配置变更记录
type ConfigChange struct {
	GroupName string
	Key       string
	OldValue  interface{}
	NewValue  interface{}
	UpdatedBy string
	Timestamp time.Time
	Version   int
}

// NewRiskConfig 创建风控配置中心
func NewRiskConfig() *RiskConfig {
	return &RiskConfig{
		groups:     make(map[string]*ConfigGroup),
		listeners:  make([]ConfigListener, 0),
		history:    make([]*ConfigChange, 0),
		maxHistory: 1000,
	}
}

// CreateGroup 创建配置组
func (rc *RiskConfig) CreateGroup(name, description string) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	if _, exists := rc.groups[name]; exists {
		return fmt.Errorf("配置组 %s 已存在", name)
	}

	rc.groups[name] = &ConfigGroup{
		Name:        name,
		Description: description,
		Items:       make(map[string]*ConfigItem),
		Version:     1,
		UpdatedAt:   time.Now(),
	}

	fmt.Printf("创建配置组: %s\n", name)
	return nil
}

// SetConfig 设置配置项
func (rc *RiskConfig) SetConfig(groupName, key string, value interface{}, description, updatedBy string) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	group, exists := rc.groups[groupName]
	if !exists {
		return fmt.Errorf("配置组 %s 不存在", groupName)
	}

	oldValue := interface{}(nil)
	var oldItem *ConfigItem

	if item, exists := group.Items[key]; exists {
		oldValue = item.Value
		oldItem = item
	}

	// 创建或更新配置项
	newItem := &ConfigItem{
		Key:         key,
		Value:       value,
		Description: description,
		Version:     1,
		UpdatedAt:   time.Now(),
		UpdatedBy:   updatedBy,
	}

	if oldItem != nil {
		newItem.Version = oldItem.Version + 1
	}

	group.Items[key] = newItem
	group.Version++
	group.UpdatedAt = time.Now()
	rc.version++

	// 记录变更历史
	change := &ConfigChange{
		GroupName: groupName,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  value,
		UpdatedBy: updatedBy,
		Timestamp: time.Now(),
		Version:   rc.version,
	}

	rc.history = append(rc.history, change)
	if len(rc.history) > rc.maxHistory {
		rc.history = rc.history[1:] // 移除最旧的记录
	}

	// 通知监听器
	go rc.notifyListeners(groupName, key, oldValue, value)

	fmt.Printf("设置配置: %s.%s = %v (by %s)\n", groupName, key, value, updatedBy)
	return nil
}

// GetConfig 获取配置项
func (rc *RiskConfig) GetConfig(groupName, key string) (interface{}, error) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	group, exists := rc.groups[groupName]
	if !exists {
		return nil, fmt.Errorf("配置组 %s 不存在", groupName)
	}

	item, exists := group.Items[key]
	if !exists {
		return nil, fmt.Errorf("配置项 %s.%s 不存在", groupName, key)
	}

	return item.Value, nil
}

// GetGroup 获取配置组
func (rc *RiskConfig) GetGroup(groupName string) (*ConfigGroup, error) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	group, exists := rc.groups[groupName]
	if !exists {
		return nil, fmt.Errorf("配置组 %s 不存在", groupName)
	}

	return group, nil
}

// DeleteConfig 删除配置项
func (rc *RiskConfig) DeleteConfig(groupName, key, deletedBy string) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	group, exists := rc.groups[groupName]
	if !exists {
		return fmt.Errorf("配置组 %s 不存在", groupName)
	}

	item, exists := group.Items[key]
	if !exists {
		return fmt.Errorf("配置项 %s.%s 不存在", groupName, key)
	}

	oldValue := item.Value
	delete(group.Items, key)
	group.Version++
	group.UpdatedAt = time.Now()
	rc.version++

	// 记录变更历史
	change := &ConfigChange{
		GroupName: groupName,
		Key:       key,
		OldValue:  oldValue,
		NewValue:  nil,
		UpdatedBy: deletedBy,
		Timestamp: time.Now(),
		Version:   rc.version,
	}

	rc.history = append(rc.history, change)
	if len(rc.history) > rc.maxHistory {
		rc.history = rc.history[1:]
	}

	// 通知监听器
	go rc.notifyListeners(groupName, key, oldValue, nil)

	fmt.Printf("删除配置: %s.%s (by %s)\n", groupName, key, deletedBy)
	return nil
}

// AddListener 添加配置监听器
func (rc *RiskConfig) AddListener(listener ConfigListener) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.listeners = append(rc.listeners, listener)
}

// notifyListeners 通知所有监听器
func (rc *RiskConfig) notifyListeners(groupName, key string, oldValue, newValue interface{}) {
	for _, listener := range rc.listeners {
		listener.OnConfigChange(groupName, key, oldValue, newValue)
	}
}

// GetHistory 获取变更历史
func (rc *RiskConfig) GetHistory(limit int) []*ConfigChange {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	if limit <= 0 || limit > len(rc.history) {
		limit = len(rc.history)
	}

	result := make([]*ConfigChange, limit)
	copy(result, rc.history[len(rc.history)-limit:])
	return result
}

// ExportConfig 导出配置
func (rc *RiskConfig) ExportConfig() ([]byte, error) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	return json.MarshalIndent(rc.groups, "", "  ")
}

// ImportConfig 导入配置
func (rc *RiskConfig) ImportConfig(data []byte, importedBy string) error {
	var groups map[string]*ConfigGroup
	if err := json.Unmarshal(data, &groups); err != nil {
		return err
	}

	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	for name, group := range groups {
		rc.groups[name] = group
		fmt.Printf("导入配置组: %s (by %s)\n", name, importedBy)
	}

	return nil
}

// GetStats 获取统计信息
func (rc *RiskConfig) GetStats() map[string]int {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()

	stats := map[string]int{
		"groups":      len(rc.groups),
		"total_items": 0,
		"history":     len(rc.history),
		"listeners":   len(rc.listeners),
		"version":     rc.version,
	}

	for _, group := range rc.groups {
		stats["total_items"] += len(group.Items)
	}

	return stats
}

// DemoListener 演示配置监听器
type DemoListener struct{}

func (dl *DemoListener) OnConfigChange(groupName, key string, oldValue, newValue interface{}) {
	fmt.Printf("配置变更通知: %s.%s 从 %v 变为 %v\n", groupName, key, oldValue, newValue)
}

func main() {
	// 创建配置中心
	config := NewRiskConfig()

	// 添加演示监听器
	config.AddListener(&DemoListener{})

	// 创建配置组
	config.CreateGroup("risk_limits", "风控限额配置")
	config.CreateGroup("blacklist", "黑名单配置")

	// 设置风控配置
	config.SetConfig("risk_limits", "max_daily_amount", 10000.0, "每日最大交易金额", "admin")
	config.SetConfig("risk_limits", "max_single_amount", 5000.0, "单笔最大交易金额", "admin")
	config.SetConfig("risk_limits", "daily_transaction_count", 50, "每日最大交易次数", "admin")

	// 设置黑名单配置
	config.SetConfig("blacklist", "enabled", true, "启用黑名单检查", "admin")
	config.SetConfig("blacklist", "check_ip", true, "检查IP黑名单", "admin")
	config.SetConfig("blacklist", "check_device", true, "检查设备黑名单", "admin")

	// 获取配置
	maxAmount, _ := config.GetConfig("risk_limits", "max_daily_amount")
	fmt.Printf("每日最大金额: %v\n", maxAmount)

	// 更新配置
	config.SetConfig("risk_limits", "max_daily_amount", 15000.0, "每日最大交易金额(已更新)", "operator")

	// 删除配置
	config.DeleteConfig("risk_limits", "daily_transaction_count", "admin")

	// 显示历史记录
	fmt.Println("\n=== 变更历史 ===")
	history := config.GetHistory(5)
	for _, change := range history {
		action := "更新"
		if change.NewValue == nil {
			action = "删除"
		} else if change.OldValue == nil {
			action = "创建"
		}
		fmt.Printf("[%s] %s %s.%s: %v -> %v (by %s)\n",
			change.Timestamp.Format("15:04:05"), action,
			change.GroupName, change.Key, change.OldValue, change.NewValue, change.UpdatedBy)
	}

	// 显示统计信息
	stats := config.GetStats()
	fmt.Printf("\n=== 统计信息 ===\n")
	fmt.Printf("配置组数量: %d\n", stats["groups"])
	fmt.Printf("配置项总数: %d\n", stats["total_items"])
	fmt.Printf("历史记录数: %d\n", stats["history"])
	fmt.Printf("监听器数量: %d\n", stats["listeners"])
	fmt.Printf("当前版本号: %d\n", stats["version"])
}