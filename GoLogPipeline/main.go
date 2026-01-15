package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// LogEntry 日志条目结构
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
}

// LogProcessor 日志处理器
type LogProcessor struct {
	logChan chan string
	entries []LogEntry
}

// NewLogProcessor 创建日志处理器
func NewLogProcessor() *LogProcessor {
	return &LogProcessor{
		logChan: make(chan string, 100),
		entries: make([]LogEntry, 0),
	}
}

// ProcessLog 处理单条日志
func (lp *LogProcessor) ProcessLog(line string) {
	// 简单的日志解析：格式 "[时间] [级别] 消息"
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", parts[0])
	if err != nil {
		return
	}

	entry := LogEntry{
		Timestamp: timestamp,
		Level:     strings.Trim(parts[1], "[]"),
		Message:   parts[2],
	}

	lp.entries = append(lp.entries, entry)
	fmt.Printf("处理日志: [%s] %s\n", entry.Level, entry.Message)
}

// FilterLogs 按级别过滤日志
func (lp *LogProcessor) FilterLogs(level string) []LogEntry {
	var filtered []LogEntry
	for _, entry := range lp.entries {
		if entry.Level == level {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

// GenerateReport 生成日志报告
func (lp *LogProcessor) GenerateReport() map[string]int {
	report := make(map[string]int)
	for _, entry := range lp.entries {
		report[entry.Level]++
	}
	return report
}

// FileReader 文件读取器
type FileReader struct {
	processor *LogProcessor
}

// NewFileReader 创建文件读取器
func NewFileReader(processor *LogProcessor) *FileReader {
	return &FileReader{processor: processor}
}

// ReadFromFile 从文件读取日志
func (fr *FileReader) ReadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fr.processor.ProcessLog(line)
	}

	return scanner.Err()
}

// StreamReader 流式读取器
type StreamReader struct {
	processor *LogProcessor
}

// NewStreamReader 创建流式读取器
func NewStreamReader(processor *LogProcessor) *StreamReader {
	return &StreamReader{processor: processor}
}

// ReadFromStdin 从标准输入读取日志
func (sr *StreamReader) ReadFromStdin() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("请输入日志行 (按Ctrl+C退出):")
	for scanner.Scan() {
		line := scanner.Text()
		sr.processor.ProcessLog(line)
	}
}

func main() {
	// 创建日志处理器
	processor := NewLogProcessor()

	// 创建文件读取器和流式读取器
	fileReader := NewFileReader(processor)
	streamReader := NewStreamReader(processor)

	// 演示文件读取
	fmt.Println("=== 文件读取演示 ===")
	if err := fileReader.ReadFromFile("sample_logs.txt"); err != nil {
		log.Printf("读取文件失败: %v", err)
	}

	// 演示流式读取（这里只是演示，实际使用时可以替换为实时流）
	fmt.Println("\n=== 流式读取演示 ===")
	streamReader.ReadFromStdin()

	// 生成报告
	fmt.Println("\n=== 日志报告 ===")
	report := processor.GenerateReport()
	for level, count := range report {
		fmt.Printf("%s: %d 条\n", level, count)
	}

	// 过滤ERROR日志
	fmt.Println("\n=== ERROR级别日志 ===")
	errorLogs := processor.FilterLogs("ERROR")
	for _, entry := range errorLogs {
		fmt.Printf("[%s] %s\n", entry.Timestamp.Format("15:04:05"), entry.Message)
	}
}
