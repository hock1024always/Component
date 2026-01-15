package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FileInfo 文件信息结构体
type FileInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	Hash    string
}

// SyncConfig 同步配置
type SyncConfig struct {
	SourceDir      string
	DestDir        string
	SyncInterval   time.Duration
	DeleteExtra    bool
	IncludeHidden  bool
}

// FileSync 文件同步器
type FileSync struct {
	config   *SyncConfig
	stopChan chan bool
}

// NewFileSync 创建文件同步器
func NewFileSync(config *SyncConfig) *FileSync {
	return &FileSync{
		config:   config,
		stopChan: make(chan bool),
	}
}

// calculateHash 计算文件MD5哈希
func (fs *FileSync) calculateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// scanDirectory 扫描目录获取文件信息
func (fs *FileSync) scanDirectory(dir string) (map[string]*FileInfo, error) {
	files := make(map[string]*FileInfo)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 跳过隐藏文件（如果配置了）
		if !fs.config.IncludeHidden && filepath.Base(path)[0] == '.' {
			return nil
		}

		// 获取相对路径
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// 计算文件哈希
		hash, err := fs.calculateHash(path)
		if err != nil {
			log.Printf("计算文件哈希失败 %s: %v", path, err)
			return nil
		}

		files[relPath] = &FileInfo{
			Path:    relPath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Hash:    hash,
		}

		return nil
	})

	return files, err
}

// syncFile 同步单个文件
func (fs *FileSync) syncFile(srcPath, destPath string, fileInfo *FileInfo) error {
	// 确保目标目录存在
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败 %s: %v", destDir, err)
	}

	// 复制文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("打开源文件失败 %s: %v", srcPath, err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败 %s: %v", destPath, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("复制文件失败 %s -> %s: %v", srcPath, destPath, err)
	}

	// 设置修改时间
	if err := os.Chtimes(destPath, time.Now(), fileInfo.ModTime); err != nil {
		log.Printf("设置文件时间失败 %s: %v", destPath, err)
	}

	fmt.Printf("已同步: %s\n", fileInfo.Path)
	return nil
}

// deleteFile 删除文件
func (fs *FileSync) deleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败 %s: %v", filePath, err)
	}
	fmt.Printf("已删除: %s\n", filepath.Base(filePath))
	return nil
}

// Sync 执行一次同步
func (fs *FileSync) Sync() error {
	fmt.Println("开始同步...")

	// 扫描源目录
	srcFiles, err := fs.scanDirectory(fs.config.SourceDir)
	if err != nil {
		return fmt.Errorf("扫描源目录失败: %v", err)
	}

	// 扫描目标目录
	destFiles, err := fs.scanDirectory(fs.config.DestDir)
	if err != nil {
		return fmt.Errorf("扫描目标目录失败: %v", err)
	}

	// 同步文件
	for relPath, srcInfo := range srcFiles {
		destPath := filepath.Join(fs.config.DestDir, relPath)
		srcPath := filepath.Join(fs.config.SourceDir, relPath)

		// 检查目标文件是否存在或需要更新
		destInfo, exists := destFiles[relPath]
		if !exists || srcInfo.Hash != destInfo.Hash {
			if err := fs.syncFile(srcPath, destPath, srcInfo); err != nil {
				log.Printf("同步文件失败 %s: %v", relPath, err)
			}
		}
	}

	// 删除多余的文件
	if fs.config.DeleteExtra {
		for relPath := range destFiles {
			if _, exists := srcFiles[relPath]; !exists {
				destPath := filepath.Join(fs.config.DestDir, relPath)
				if err := fs.deleteFile(destPath); err != nil {
					log.Printf("删除文件失败 %s: %v", relPath, err)
				}
			}
		}
	}

	fmt.Printf("同步完成，源目录%d个文件，目标目录%d个文件\n", len(srcFiles), len(destFiles))
	return nil
}

// Start 开始定期同步
func (fs *FileSync) Start() {
	fmt.Printf("文件同步器已启动，间隔: %v\n", fs.config.SyncInterval)

	ticker := time.NewTicker(fs.config.SyncInterval)
	defer ticker.Stop()

	// 立即执行一次同步
	if err := fs.Sync(); err != nil {
		log.Printf("初始同步失败: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := fs.Sync(); err != nil {
				log.Printf("定期同步失败: %v", err)
			}
		case <-fs.stopChan:
			fmt.Println("文件同步器已停止")
			return
		}
	}
}

// Stop 停止同步
func (fs *FileSync) Stop() {
	close(fs.stopChan)
}

// GetStats 获取同步统计信息
func (fs *FileSync) GetStats() (map[string]int, error) {
	srcFiles, err := fs.scanDirectory(fs.config.SourceDir)
	if err != nil {
		return nil, err
	}

	destFiles, err := fs.scanDirectory(fs.config.DestDir)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"source_files": len(srcFiles),
		"dest_files":   len(destFiles),
	}

	return stats, nil
}

func main() {
	// 创建测试目录
	os.MkdirAll("source", 0755)
	os.MkdirAll("dest", 0755)

	// 创建测试文件
	testFile1 := "source/test1.txt"
	testFile2 := "source/test2.txt"

	os.WriteFile(testFile1, []byte("这是测试文件1"), 0644)
	os.WriteFile(testFile2, []byte("这是测试文件2"), 0644)

	// 配置同步器
	config := &SyncConfig{
		SourceDir:     "source",
		DestDir:       "dest",
		SyncInterval:  5 * time.Second,
		DeleteExtra:   true,
		IncludeHidden: false,
	}

	sync := NewFileSync(config)

	// 执行一次性同步
	fmt.Println("=== 执行一次性同步 ===")
	if err := sync.Sync(); err != nil {
		log.Fatal("同步失败:", err)
	}

	// 修改源文件
	time.Sleep(1 * time.Second)
	os.WriteFile(testFile1, []byte("这是修改后的测试文件1"), 0644)

	// 再次同步
	fmt.Println("\n=== 再次同步 ===")
	if err := sync.Sync(); err != nil {
		log.Fatal("同步失败:", err)
	}

	// 显示统计信息
	stats, err := sync.GetStats()
	if err != nil {
		log.Fatal("获取统计失败:", err)
	}

	fmt.Printf("\n=== 统计信息 ===\n")
	fmt.Printf("源目录文件数: %d\n", stats["source_files"])
	fmt.Printf("目标目录文件数: %d\n", stats["dest_files"])
}