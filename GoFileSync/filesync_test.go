package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDirs(t *testing.T) (string, string, func()) {
	// 创建临时目录
	sourceDir, err := os.MkdirTemp("", "filesync_source_*")
	if err != nil {
		t.Fatal(err)
	}

	destDir, err := os.MkdirTemp("", "filesync_dest_*")
	if err != nil {
		t.Fatal(err)
	}

	// 清理函数
	cleanup := func() {
		os.RemoveAll(sourceDir)
		os.RemoveAll(destDir)
	}

	return sourceDir, destDir, cleanup
}

func TestFileSync(t *testing.T) {
	sourceDir, destDir, cleanup := setupTestDirs(t)
	defer cleanup()

	// 创建测试文件
	testFile1 := filepath.Join(sourceDir, "test1.txt")
	testFile2 := filepath.Join(sourceDir, "test2.txt")

	os.WriteFile(testFile1, []byte("content1"), 0644)
	os.WriteFile(testFile2, []byte("content2"), 0644)

	// 配置同步器
	config := &SyncConfig{
		SourceDir:     sourceDir,
		DestDir:       destDir,
		DeleteExtra:   true,
		IncludeHidden: false,
	}

	sync := NewFileSync(config)

	// 执行同步
	err := sync.Sync()
	if err != nil {
		t.Fatal("同步失败:", err)
	}

	// 检查文件是否被复制
	destFile1 := filepath.Join(destDir, "test1.txt")
	destFile2 := filepath.Join(destDir, "test2.txt")

	if _, err := os.Stat(destFile1); os.IsNotExist(err) {
		t.Error("文件1未被同步")
	}

	if _, err := os.Stat(destFile2); os.IsNotExist(err) {
		t.Error("文件2未被同步")
	}

	// 验证文件内容
	content1, _ := os.ReadFile(destFile1)
	if string(content1) != "content1" {
		t.Errorf("文件1内容不匹配，期望'content1'，实际'%s'", string(content1))
	}
}

func TestFileUpdate(t *testing.T) {
	sourceDir, destDir, cleanup := setupTestDirs(t)
	defer cleanup()

	testFile := filepath.Join(sourceDir, "test.txt")
	destFile := filepath.Join(destDir, "test.txt")

	// 创建初始文件
	os.WriteFile(testFile, []byte("original"), 0644)

	config := &SyncConfig{
		SourceDir:     sourceDir,
		DestDir:       destDir,
		DeleteExtra:   true,
		IncludeHidden: false,
	}

	sync := NewFileSync(config)

	// 首次同步
	sync.Sync()

	// 修改源文件
	time.Sleep(1 * time.Second) // 确保修改时间不同
	os.WriteFile(testFile, []byte("updated"), 0644)

	// 再次同步
	sync.Sync()

	// 检查目标文件是否被更新
	content, _ := os.ReadFile(destFile)
	if string(content) != "updated" {
		t.Errorf("文件未被更新，期望'updated'，实际'%s'", string(content))
	}
}

func TestDeleteExtraFiles(t *testing.T) {
	sourceDir, destDir, cleanup := setupTestDirs(t)
	defer cleanup()

	// 在目标目录创建多余文件
	extraFile := filepath.Join(destDir, "extra.txt")
	os.WriteFile(extraFile, []byte("extra"), 0644)

	// 源目录没有任何文件
	config := &SyncConfig{
		SourceDir:     sourceDir,
		DestDir:       destDir,
		DeleteExtra:   true,
		IncludeHidden: false,
	}

	sync := NewFileSync(config)

	// 执行同步
	sync.Sync()

	// 检查多余文件是否被删除
	if _, err := os.Stat(extraFile); !os.IsNotExist(err) {
		t.Error("多余文件未被删除")
	}
}

func TestHiddenFiles(t *testing.T) {
	sourceDir, destDir, cleanup := setupTestDirs(t)
	defer cleanup()

	// 创建普通文件和隐藏文件
	normalFile := filepath.Join(sourceDir, "normal.txt")
	hiddenFile := filepath.Join(sourceDir, ".hidden.txt")

	os.WriteFile(normalFile, []byte("normal"), 0644)
	os.WriteFile(hiddenFile, []byte("hidden"), 0644)

	// 配置为不包含隐藏文件
	config := &SyncConfig{
		SourceDir:     sourceDir,
		DestDir:       destDir,
		DeleteExtra:   false,
		IncludeHidden: false,
	}

	sync := NewFileSync(config)

	// 执行同步
	sync.Sync()

	// 检查普通文件被同步，隐藏文件未被同步
	destNormal := filepath.Join(destDir, "normal.txt")
	destHidden := filepath.Join(destDir, ".hidden.txt")

	if _, err := os.Stat(destNormal); os.IsNotExist(err) {
		t.Error("普通文件未被同步")
	}

	if _, err := os.Stat(destHidden); !os.IsNotExist(err) {
		t.Error("隐藏文件被同步了")
	}
}

func TestGetStats(t *testing.T) {
	sourceDir, destDir, cleanup := setupTestDirs(t)
	defer cleanup()

	// 创建测试文件
	os.WriteFile(filepath.Join(sourceDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(sourceDir, "file2.txt"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(destDir, "file3.txt"), []byte("content3"), 0644)

	config := &SyncConfig{
		SourceDir:     sourceDir,
		DestDir:       destDir,
		DeleteExtra:   false,
		IncludeHidden: false,
	}

	sync := NewFileSync(config)

	stats, err := sync.GetStats()
	if err != nil {
		t.Fatal("获取统计失败:", err)
	}

	if stats["source_files"] != 2 {
		t.Errorf("期望源目录2个文件，实际%d个", stats["source_files"])
	}

	if stats["dest_files"] != 1 {
		t.Errorf("期望目标目录1个文件，实际%d个", stats["dest_files"])
	}
}