package main

import (
	"testing"
	"time"
)

func TestCreateGroup(t *testing.T) {
	config := NewRiskConfig()

	// 测试创建配置组
	err := config.CreateGroup("test_group", "测试配置组")
	if err != nil {
		t.Errorf("创建配置组失败: %v", err)
	}

	// 测试重复创建
	err = config.CreateGroup("test_group", "重复的配置组")
	if err == nil {
		t.Error("期望重复创建失败，但成功了")
	}
}

func TestSetAndGetConfig(t *testing.T) {
	config := NewRiskConfig()
	config.CreateGroup("test_group", "测试配置组")

	// 测试设置配置
	err := config.SetConfig("test_group", "test_key", "test_value", "测试配置", "admin")
	if err != nil {
		t.Errorf("设置配置失败: %v", err)
	}

	// 测试获取配置
	value, err := config.GetConfig("test_group", "test_key")
	if err != nil {
		t.Errorf("获取配置失败: %v", err)
	}

	if value != "test_value" {
		t.Errorf("期望值'test_value'，实际值'%v'", value)
	}

	// 测试获取不存在的配置
	_, err = config.GetConfig("test_group", "nonexistent")
	if err == nil {
		t.Error("期望获取不存在配置失败，但成功了")
	}

	// 测试获取不存在的配置组
	_, err = config.GetConfig("nonexistent", "test_key")
	if err == nil {
		t.Error("期望获取不存在配置组失败，但成功了")
	}
}

func TestConfigUpdate(t *testing.T) {
	config := NewRiskConfig()
	config.CreateGroup("test_group", "测试配置组")

	// 设置初始值
	config.SetConfig("test_group", "test_key", "old_value", "测试配置", "admin")

	// 更新配置
	config.SetConfig("test_group", "test_key", "new_value", "更新后的配置", "operator")

	// 验证新值
	value, _ := config.GetConfig("test_group", "test_key")
	if value != "new_value" {
		t.Errorf("期望值'new_value'，实际值'%v'", value)
	}
}

func TestDeleteConfig(t *testing.T) {
	config := NewRiskConfig()
	config.CreateGroup("test_group", "测试配置组")

	// 设置配置
	config.SetConfig("test_group", "test_key", "test_value", "测试配置", "admin")

	// 删除配置
	err := config.DeleteConfig("test_group", "test_key", "admin")
	if err != nil {
		t.Errorf("删除配置失败: %v", err)
	}

	// 验证配置已被删除
	_, err = config.GetConfig("test_group", "test_key")
	if err == nil {
		t.Error("期望配置已被删除，但仍能获取到")
	}
}

func TestConfigHistory(t *testing.T) {
	config := NewRiskConfig()
	config.CreateGroup("test_group", "测试配置组")

	// 执行一系列操作
	config.SetConfig("test_group", "key1", "value1", "配置1", "admin")
	config.SetConfig("test_group", "key2", "value2", "配置2", "admin")
	config.SetConfig("test_group", "key1", "new_value1", "更新配置1", "operator")
	config.DeleteConfig("test_group", "key2", "admin")

	// 获取历史记录
	history := config.GetHistory(10)

	if len(history) != 4 {
		t.Errorf("期望4条历史记录，实际%d条", len(history))
	}

	// 验证最后一条记录是删除操作
	lastChange := history[len(history)-1]
	if lastChange.NewValue != nil {
		t.Error("期望最后一条记录是删除操作")
	}
}

func TestConfigListener(t *testing.T) {
	config := NewRiskConfig()
	config.CreateGroup("test_group", "测试配置组")

	// 创建监听器
	changeCount := 0
	var lastChange struct {
		group, key string
		oldVal, newVal interface{}
	}

	listener := &testListener{
		onChange: func(groupName, key string, oldValue, newValue interface{}) {
			changeCount++
			lastChange.group = groupName
			lastChange.key = key
			lastChange.oldVal = oldValue
			lastChange.newVal = newValue
		},
	}

	config.AddListener(listener)

	// 设置配置
	config.SetConfig("test_group", "test_key", "test_value", "测试", "admin")

	// 等待异步通知完成
	time.Sleep(100 * time.Millisecond)

	if changeCount != 1 {
		t.Errorf("期望1次变更通知，实际%d次", changeCount)
	}

	if lastChange.group != "test_group" || lastChange.key != "test_key" {
		t.Errorf("变更通知信息不正确")
	}
}

// testListener 测试用监听器
type testListener struct {
	onChange func(string, string, interface{}, interface{})
}

func (tl *testListener) OnConfigChange(groupName, key string, oldValue, newValue interface{}) {
	if tl.onChange != nil {
		tl.onChange(groupName, key, oldValue, newValue)
	}
}

func TestGetStats(t *testing.T) {
	config := NewRiskConfig()

	// 创建配置组和配置项
	config.CreateGroup("group1", "组1")
	config.CreateGroup("group2", "组2")
	config.SetConfig("group1", "key1", "value1", "配置1", "admin")
	config.SetConfig("group1", "key2", "value2", "配置2", "admin")
	config.SetConfig("group2", "key3", "value3", "配置3", "admin")

	stats := config.GetStats()

	if stats["groups"] != 2 {
		t.Errorf("期望2个配置组，实际%d个", stats["groups"])
	}

	if stats["total_items"] != 3 {
		t.Errorf("期望3个配置项，实际%d个", stats["total_items"])
	}
}

func TestExportImportConfig(t *testing.T) {
	config1 := NewRiskConfig()
	config1.CreateGroup("test_group", "测试组")
	config1.SetConfig("test_group", "key1", "value1", "配置1", "admin")

	// 导出配置
	data, err := config1.ExportConfig()
	if err != nil {
		t.Fatalf("导出配置失败: %v", err)
	}

	// 导入到另一个配置中心
	config2 := NewRiskConfig()
	err = config2.ImportConfig(data, "importer")
	if err != nil {
		t.Fatalf("导入配置失败: %v", err)
	}

	// 验证导入的配置
	value, err := config2.GetConfig("test_group", "key1")
	if err != nil {
		t.Errorf("获取导入的配置失败: %v", err)
	}

	if value != "value1" {
		t.Errorf("期望值'value1'，实际值'%v'", value)
	}
}