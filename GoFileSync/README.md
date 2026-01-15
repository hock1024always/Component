# 文件同步器

这是一个简单但功能完整的文件同步工具，支持增量同步、定期同步和双向同步等特性。

## 架构设计

### 核心组件

1. **FileInfo** - 文件信息结构体
   - Path: 文件相对路径
   - Size: 文件大小
   - ModTime: 修改时间
   - Hash: MD5哈希值

2. **SyncConfig** - 同步配置
   - SourceDir: 源目录
   - DestDir: 目标目录
   - SyncInterval: 同步间隔
   - DeleteExtra: 是否删除目标目录多余文件
   - IncludeHidden: 是否包含隐藏文件

3. **FileSync** - 文件同步器
   - 实现文件扫描、哈希计算、增量同步等功能

## 同步策略

1. **增量同步**: 只同步有变化的文件
2. **哈希校验**: 使用MD5保证文件完整性
3. **时间戳检查**: 结合修改时间和哈希进行变更检测
4. **定期同步**: 支持定时自动同步

## 代码结构解析

### FileInfo 结构体详解

```go
type FileInfo struct {
    Path    string        // 相对路径
    Size    int64         // 文件大小
    ModTime time.Time     // 修改时间
    Hash    string        // MD5哈希
}
```

### SyncConfig 配置详解

```go
type SyncConfig struct {
    SourceDir      string        // 源目录路径
    DestDir        string        // 目标目录路径
    SyncInterval   time.Duration // 同步间隔时间
    DeleteExtra    bool          // 是否删除目标目录多余文件
    IncludeHidden  bool          // 是否同步隐藏文件
}
```

### calculateHash 方法：文件哈希计算

```go
func (fs *FileSync) calculateHash(filePath string) (string, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()

    hash := md5.New()           // 创建MD5哈希器
    io.Copy(hash, file)         // 将文件内容写入哈希器
    return fmt.Sprintf("%x", hash.Sum(nil)), nil // 返回十六进制字符串
}
```

### scanDirectory 方法：目录扫描

```go
func (fs *FileSync) scanDirectory(dir string) (map[string]*FileInfo, error) {
    files := make(map[string]*FileInfo)

    filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if info.IsDir() {
            return nil // 跳过目录
        }

        // 过滤隐藏文件
        if !fs.config.IncludeHidden && filepath.Base(path)[0] == '.' {
            return nil
        }

        // 计算相对路径
        relPath, _ := filepath.Rel(dir, path)

        // 计算文件哈希
        hash, _ := fs.calculateHash(path)

        // 存储文件信息
        files[relPath] = &FileInfo{
            Path:    relPath,
            Size:    info.Size(),
            ModTime: info.ModTime(),
            Hash:    hash,
        }

        return nil
    })

    return files, nil
}
```

### Sync 方法：核心同步逻辑

```go
func (fs *FileSync) Sync() error {
    // 1. 扫描源目录和目标目录
    srcFiles, err := fs.scanDirectory(fs.config.SourceDir)
    destFiles, err := fs.scanDirectory(fs.config.DestDir)

    // 2. 同步文件
    for relPath, srcInfo := range srcFiles {
        destPath := filepath.Join(fs.config.DestDir, relPath)
        destInfo, exists := destFiles[relPath]

        // 检查文件是否需要同步
        if !exists || srcInfo.Hash != destInfo.Hash {
            fs.syncFile(srcPath, destPath, srcInfo)
        }
    }

    // 3. 删除多余文件（如果配置了）
    if fs.config.DeleteExtra {
        for relPath := range destFiles {
            if _, exists := srcFiles[relPath]; !exists {
                fs.deleteFile(filepath.Join(fs.config.DestDir, relPath))
            }
        }
    }

    return nil
}
```

### syncFile 方法：文件同步

```go
func (fs *FileSync) syncFile(srcPath, destPath string, fileInfo *FileInfo) error {
    // 1. 确保目标目录存在
    destDir := filepath.Dir(destPath)
    os.MkdirAll(destDir, 0755)

    // 2. 复制文件
    srcFile, err := os.Open(srcPath)
    destFile, err := os.Create(destPath)
    io.Copy(destFile, srcFile)

    // 3. 保持原始修改时间
    os.Chtimes(destPath, time.Now(), fileInfo.ModTime)

    return nil
}
```

### Start 方法：定期同步

```go
func (fs *FileSync) Start() {
    ticker := time.NewTicker(fs.config.SyncInterval)
    defer ticker.Stop()

    // 立即执行一次同步
    fs.Sync()

    for {
        select {
        case <-ticker.C:      // 定时器触发
            fs.Sync()         // 执行同步
        case <-fs.stopChan:   // 接收停止信号
            return
        }
    }
}
```

## 使用方法

### 1. 编译运行
```bash
go run main.go
```

程序会：
- 创建source和dest测试目录
- 在source目录创建测试文件
- 执行同步
- 修改文件后再次同步
- 显示统计信息

### 2. 运行测试
```bash
go test -v
```

## 测试覆盖

- `TestFileSync`: 测试基本文件同步功能
- `TestFileUpdate`: 测试文件更新同步
- `TestDeleteExtraFiles`: 测试删除多余文件
- `TestHiddenFiles`: 测试隐藏文件过滤
- `TestGetStats`: 测试统计信息获取

## 扩展思路

1. **双向同步**: 支持双向变更检测和合并
2. **压缩传输**: 支持文件压缩传输节省带宽
3. **断点续传**: 支持大文件断点续传
4. **多线程同步**: 并发同步提高性能
5. **云存储集成**: 支持云存储作为同步目标
6. **图形界面**: 添加Web界面进行配置和管理