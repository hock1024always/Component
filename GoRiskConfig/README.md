# 风控系统配置中心

这是一个轻量级的配置管理中心，专为风控系统设计，支持配置分组、版本控制、变更监听和历史记录等功能。

## 架构设计

### 核心组件

1. **ConfigItem** - 配置项
   - Key: 配置键
   - Value: 配置值
   - Description: 描述
   - Version: 版本号
   - UpdatedAt: 更新时间
   - UpdatedBy: 更新者

2. **ConfigGroup** - 配置组
   - Name: 组名
   - Description: 描述
   - Items: 配置项集合
   - Version: 组版本号

3. **RiskConfig** - 配置中心核心
   - 管理所有配置组
   - 支持配置监听
   - 维护变更历史

4. **ConfigListener** - 配置监听器接口
   - OnConfigChange(): 配置变更回调

5. **ConfigChange** - 配置变更记录
   - 记录每次配置变更的详细信息

## 核心特性

1. **配置分组**: 支持按业务模块分组管理配置
2. **版本控制**: 每个配置项都有版本号，配置组也有整体版本
3. **变更监听**: 支持配置变更的实时监听
4. **历史记录**: 记录所有配置变更历史
5. **并发安全**: 基于读写锁保证并发访问安全
6. **导入导出**: 支持配置的JSON格式导入导出

## 代码结构解析

### ConfigItem 结构体详解

```go
type ConfigItem struct {
    Key         string      // 配置键名
    Value       interface{} // 配置值，支持任意类型
    Description string      // 配置描述
    Version     int         // 配置版本号
    UpdatedAt   time.Time   // 最后更新时间
    UpdatedBy   string      // 更新者
}
```

### ConfigGroup 结构体详解

```go
type ConfigGroup struct {
    Name        string                 // 组名
    Description string                 // 组描述
    Items       map[string]*ConfigItem // 配置项映射
    Version     int                    // 组版本号
    UpdatedAt   time.Time              // 组最后更新时间
}
```

### RiskConfig 核心类详解

```go
type RiskConfig struct {
    groups     map[string]*ConfigGroup // 配置组存储
    listeners  []ConfigListener        // 配置监听器列表
    mutex      sync.RWMutex            // 读写锁保证并发安全
    version    int                     // 全局版本号
    history    []*ConfigChange         // 变更历史记录
    maxHistory int                     // 最大历史记录数
}
```

### SetConfig 方法：设置配置

```go
func (rc *RiskConfig) SetConfig(groupName, key string, value interface{}, description, updatedBy string) error {
    rc.mutex.Lock()  // 加写锁保证原子性
    defer rc.mutex.Unlock()

    // 1. 查找配置组
    group, exists := rc.groups[groupName]
    if !exists {
        return fmt.Errorf("配置组 %s 不存在", groupName)
    }

    // 2. 保存旧值用于历史记录
    oldValue := interface{}(nil)
    if item, exists := group.Items[key]; exists {
        oldValue = item.Value
    }

    // 3. 创建或更新配置项
    newItem := &ConfigItem{
        Key:         key,
        Value:       value,
        Description: description,
        Version:     1,  // 新配置从版本1开始
        UpdatedAt:   time.Now(),
        UpdatedBy:   updatedBy,
    }

    if oldItem != nil {
        newItem.Version = oldItem.Version + 1  // 版本递增
    }

    // 4. 更新配置
    group.Items[key] = newItem
    group.Version++      // 组版本递增
    group.UpdatedAt = time.Now()
    rc.version++         // 全局版本递增

    // 5. 记录变更历史
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

    // 6. 异步通知监听器
    go rc.notifyListeners(groupName, key, oldValue, value)

    return nil
}
```

### AddListener 方法：添加监听器

```go
func (rc *RiskConfig) AddListener(listener ConfigListener) {
    rc.mutex.Lock()
    defer rc.mutex.Unlock()
    rc.listeners = append(rc.listeners, listener)
}
```

### notifyListeners 方法：通知监听器

```go
func (rc *RiskConfig) notifyListeners(groupName, key string, oldValue, newValue interface{}) {
    // 异步通知所有监听器，避免阻塞主线程
    for _, listener := range rc.listeners {
        listener.OnConfigChange(groupName, key, oldValue, newValue)
    }
}
```

### GetHistory 方法：获取变更历史

```go
func (rc *RiskConfig) GetHistory(limit int) []*ConfigChange {
    rc.mutex.RLock()
    defer rc.mutex.RUnlock()

    if limit <= 0 || limit > len(rc.history) {
        limit = len(rc.history)
    }

    // 返回最新的limit条记录
    result := make([]*ConfigChange, limit)
    copy(result, rc.history[len(rc.history)-limit:])
    return result
}
```

### ExportConfig 方法：导出配置

```go
func (rc *RiskConfig) ExportConfig() ([]byte, error) {
    rc.mutex.RLock()
    defer rc.mutex.RUnlock()

    // 导出为格式化的JSON
    return json.MarshalIndent(rc.groups, "", "  ")
}
```

## 使用方法

### 1. 编译运行
```bash
go run main.go
```

程序会演示：
- 创建配置组
- 设置风控配置项
- 更新和删除配置
- 显示变更历史
- 显示统计信息

### 2. 运行测试
```bash
go test -v
```

## 测试覆盖

- `TestCreateGroup`: 测试配置组创建
- `TestSetAndGetConfig`: 测试配置设置和获取
- `TestConfigUpdate`: 测试配置更新
- `TestDeleteConfig`: 测试配置删除
- `TestConfigHistory`: 测试变更历史记录
- `TestConfigListener`: 测试配置监听器
- `TestGetStats`: 测试统计信息
- `TestExportImportConfig`: 测试配置导入导出

## 扩展思路

1. **持久化存储**: 支持数据库或文件系统持久化
2. **配置验证**: 添加配置值类型和范围验证
3. **权限控制**: 基于角色的配置管理权限
4. **配置模板**: 支持配置模板和继承
5. **分布式同步**: 支持多实例间的配置同步
6. **配置回滚**: 支持配置变更的回滚操作
7. **监控告警**: 配置变更的监控和告警机制