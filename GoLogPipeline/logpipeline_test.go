package main

import (
	"os"
	"testing"
)

func TestLogProcessor(t *testing.T) {
	processor := NewLogProcessor()

	// 测试日志处理
	testLog := "2024-01-15 10:30:15 [INFO] 测试消息"
	processor.ProcessLog(testLog)

	if len(processor.entries) != 1 {
		t.Errorf("期望1条日志，实际%d条", len(processor.entries))
	}

	entry := processor.entries[0]
	if entry.Level != "INFO" {
		t.Errorf("期望级别INFO，实际%s", entry.Level)
	}
	if entry.Message != "测试消息" {
		t.Errorf("期望消息'测试消息'，实际'%s'", entry.Message)
	}
}

func TestFilterLogs(t *testing.T) {
	processor := NewLogProcessor()

	// 添加测试日志
	processor.ProcessLog("2024-01-15 10:30:15 [INFO] 信息日志")
	processor.ProcessLog("2024-01-15 10:30:16 [ERROR] 错误日志")
	processor.ProcessLog("2024-01-15 10:30:17 [WARN] 警告日志")
	processor.ProcessLog("2024-01-15 10:30:18 [ERROR] 另一个错误")

	// 过滤ERROR日志
	errorLogs := processor.FilterLogs("ERROR")
	if len(errorLogs) != 2 {
		t.Errorf("期望2条ERROR日志，实际%d条", len(errorLogs))
	}

	// 过滤INFO日志
	infoLogs := processor.FilterLogs("INFO")
	if len(infoLogs) != 1 {
		t.Errorf("期望1条INFO日志，实际%d条", len(infoLogs))
	}
}

func TestGenerateReport(t *testing.T) {
	processor := NewLogProcessor()

	// 添加测试日志
	processor.ProcessLog("2024-01-15 10:30:15 [INFO] 日志1")
	processor.ProcessLog("2024-01-15 10:30:16 [ERROR] 日志2")
	processor.ProcessLog("2024-01-15 10:30:17 [INFO] 日志3")
	processor.ProcessLog("2024-01-15 10:30:18 [ERROR] 日志4")

	report := processor.GenerateReport()

	if report["INFO"] != 2 {
		t.Errorf("期望INFO级别2条，实际%d条", report["INFO"])
	}
	if report["ERROR"] != 2 {
		t.Errorf("期望ERROR级别2条，实际%d条", report["ERROR"])
	}
}

func TestFileReader(t *testing.T) {
	// 创建临时测试文件
	testContent := "2024-01-15 10:30:15 [INFO] 测试文件日志\n"
	testContent += "2024-01-15 10:30:16 [ERROR] 测试错误日志\n"

	tmpFile, err := os.CreateTemp("", "test_log_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// 测试文件读取
	processor := NewLogProcessor()
	fileReader := NewFileReader(processor)

	err = fileReader.ReadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(processor.entries) != 2 {
		t.Errorf("期望读取2条日志，实际%d条", len(processor.entries))
	}
}
