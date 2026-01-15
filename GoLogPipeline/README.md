# 日志数据处理管道

这是一个简单的日志数据处理管道系统，支持从文件或流式输入读取日志，进行解析、过滤和统计分析。

## 架构设计

### 核心组件

1. **LogEntry** - 日志条目结构体
   - Timestamp: 时间戳
   - Level: 日志级别 (INFO, ERROR, WARN, DEBUG)
   - Message: 日志消息

2. **LogProcessor** - 日志处理器
   - ProcessLog(): 处理单条日志
   - FilterLogs(): 按级别过滤日志
   - GenerateReport(): 生成统计报告

3. **FileReader** - 文件读取器
   - ReadFromFile(): 从文件读取日志

4. **StreamReader** - 流式读取器
   - ReadFromStdin(): 从标准输入流式读取

## 使用方法

### 1. 编译运行
```bash
go run main.go
```

### 2. 文件读取
程序会自动读取 `sample_logs.txt` 文件中的日志。

### 3. 流式读取
程序会等待从标准输入读取日志行。

### 4. 运行测试
```bash
go test -v
```

## 代码结构解析

### main.go 详解

```go
// LogEntry 日志条目结构
type LogEntry struct {
    Timestamp time.Time  // 时间戳
    Level     string     // 日志级别
    Message   string     // 消息内容
}

// LogProcessor 日志处理器核心
type LogProcessor struct {
    logChan chan string     // 日志通道 (预留给并发处理)
    entries []LogEntry      // 存储所有日志条目
}
```

**ProcessLog 方法**: 解析单行日志
```go
func (lp *LogProcessor) ProcessLog(line string) {
    // 1. 按空格分割: "[时间] [级别] 消息"
    parts := strings.SplitN(line, " ", 3)
    if len(parts) < 3 {
        return // 格式不正确跳过
    }

    // 2. 解析时间戳
    timestamp, err := time.Parse("2006-01-02 15:04:05", parts[0])
    if err != nil {
        return // 时间格式错误跳过
    }

    // 3. 创建日志条目并存储
    entry := LogEntry{
        Timestamp: timestamp,
        Level:     strings.Trim(parts[1], "[]"), // 移除方括号
        Message:   parts[2],
    }
    lp.entries = append(lp.entries, entry)
}
```

**FilterLogs 方法**: 按级别过滤
```go
func (lp *LogProcessor) FilterLogs(level string) []LogEntry {
    var filtered []LogEntry
    for _, entry := range lp.entries {
        if entry.Level == level { // 匹配指定级别
            filtered = append(filtered, entry)
        }
    }
    return filtered
}
```

**GenerateReport 方法**: 统计各级别日志数量
```go
func (lp *LogProcessor) GenerateReport() map[string]int {
    report := make(map[string]int)
    for _, entry := range lp.entries {
        report[entry.Level]++ // 计数器递增
    }
    return report
}
```

### FileReader 文件读取器

```go
type FileReader struct {
    processor *LogProcessor // 持有处理器引用
}

func (fr *FileReader) ReadFromFile(filename string) error {
    file, err := os.Open(filename)     // 打开文件
    if err != nil {
        return err
    }
    defer file.Close()                 // 确保文件关闭

    scanner := bufio.NewScanner(file)  // 创建扫描器
    for scanner.Scan() {              // 逐行扫描
        line := scanner.Text()
        fr.processor.ProcessLog(line) // 处理每一行
    }

    return scanner.Err()              // 返回扫描错误
}
```

### StreamReader 流式读取器

```go
func (sr *StreamReader) ReadFromStdin() {
    scanner := bufio.NewScanner(os.Stdin) // 从标准输入读取
    fmt.Println("请输入日志行 (按Ctrl+C退出):")
    for scanner.Scan() {
        line := scanner.Text()
        sr.processor.ProcessLog(line) // 实时处理
    }
}
```

## 测试覆盖

- `TestLogProcessor`: 测试日志解析功能
- `TestFilterLogs`: 测试日志过滤功能
- `TestGenerateReport`: 测试统计报告功能
- `TestFileReader`: 测试文件读取功能

## 扩展思路

1. **并发处理**: 使用goroutine并发处理多文件
2. **配置化**: 支持配置日志格式和过滤规则
3. **持久化**: 将处理结果保存到数据库
4. **监控**: 添加性能监控和健康检查