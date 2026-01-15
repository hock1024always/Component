# 实时结算系统

这是一个高性能的实时结算系统，支持账户管理、交易处理、批量结算和资金冻结等功能。

## 架构设计

### 核心组件

1. **Transaction** - 交易记录
   - ID: 交易唯一标识
   - UserID: 用户ID
   - Amount: 交易金额
   - Type: 交易类型 (debit入账, credit出账)
   - Status: 交易状态
   - Timestamp: 交易时间

2. **Account** - 账户信息
   - UserID: 用户ID
   - Balance: 可用余额
   - FrozenAmount: 冻结金额
   - Version: 版本号（乐观锁）
   - UpdatedAt: 更新时间

3. **SettlementResult** - 结算结果
   - TransactionID: 交易ID
   - Success: 是否成功
   - NewBalance: 新余额
   - ErrorMessage: 错误信息
   - Timestamp: 处理时间

4. **SettlementEngine** - 结算引擎核心
   - 账户管理
   - 交易队列处理
   - 批量结算
   - 并发安全

## 核心特性

1. **实时结算**: 支持实时交易处理和余额更新
2. **批量处理**: 支持批量交易处理提高吞吐量
3. **并发安全**: 基于读写锁保证数据一致性
4. **资金冻结**: 支持金额冻结和解冻操作
5. **交易记录**: 完整的交易历史记录
6. **错误处理**: 完善的交易失败处理机制

## 结算策略

1. **队列处理**: 使用channel实现交易队列，避免并发冲突
2. **批量优化**: 达到批次大小时立即处理，或超时后处理
3. **乐观锁**: 使用版本号防止并发更新冲突
4. **事务一致性**: 保证账户余额计算的原子性

## 代码结构解析

### Transaction 交易结构体详解

```go
type Transaction struct {
    ID          string    // 交易ID，自动生成
    UserID      string    // 用户ID
    Amount      float64   // 交易金额
    Type        string    // 交易类型: "debit"(出账), "credit"(入账)
    Status      string    // 状态: pending, processed
    Timestamp   time.Time // 交易时间戳
    Description string    // 交易描述
}
```

### Account 账户结构体详解

```go
type Account struct {
    UserID       string    // 用户ID
    Balance      float64   // 可用余额
    FrozenAmount float64   // 冻结金额
    Version      int64     // 版本号，用于乐观锁
    UpdatedAt    time.Time // 最后更新时间
}
```

### SettlementEngine 结算引擎详解

```go
type SettlementEngine struct {
    accounts       map[string]*Account    // 账户存储
    transactions   []Transaction          // 交易历史
    mutex          sync.RWMutex           // 读写锁
    settlementChan chan *Transaction      // 结算队列
    stopChan       chan bool              // 停止信号
    batchSize      int                    // 批处理大小
    batchTimeout   time.Duration          // 批处理超时
}
```

### SubmitTransaction 方法：提交交易

```go
func (se *SettlementEngine) SubmitTransaction(tx *Transaction) error {
    // 1. 参数验证
    if tx.UserID == "" || tx.Amount <= 0 {
        return fmt.Errorf("无效的交易参数")
    }

    // 2. 生成交易ID和时间戳
    tx.ID = fmt.Sprintf("tx_%d", time.Now().UnixNano())
    tx.Timestamp = time.Now()
    tx.Status = "pending"

    // 3. 记录交易到历史
    se.mutex.Lock()
    se.transactions = append(se.transactions, *tx)
    se.mutex.Unlock()

    // 4. 提交到结算队列
    select {
    case se.settlementChan <- tx:
        return nil
    default:
        return fmt.Errorf("结算队列已满")
    }
}
```

### processSettlementQueue 方法：队列处理核心

```go
func (se *SettlementEngine) processSettlementQueue() {
    batch := make([]*Transaction, 0, se.batchSize)  // 当前批次
    timer := time.NewTimer(se.batchTimeout)         // 超时定时器
    defer timer.Stop()

    for {
        select {
        case tx := <-se.settlementChan:
            batch = append(batch, tx)

            // 达到批处理大小时立即处理
            if len(batch) >= se.batchSize {
                se.processBatch(batch)      // 处理当前批次
                batch = batch[:0]            // 清空批次
                timer.Reset(se.batchTimeout) // 重置定时器
            }

        case <-timer.C:
            // 超时处理当前批次
            if len(batch) > 0 {
                se.processBatch(batch)
                batch = batch[:0]
            }
            timer.Reset(se.batchTimeout)

        case <-se.stopChan:
            // 程序退出时处理剩余批次
            if len(batch) > 0 {
                se.processBatch(batch)
            }
            return
        }
    }
}
```

### batchProcessTransactions 方法：批量处理交易

```go
func (se *SettlementEngine) batchProcessTransactions(txs []*Transaction) []*SettlementResult {
    results := make([]*SettlementResult, len(txs))

    se.mutex.Lock()  // 批处理需要独占锁
    defer se.mutex.Unlock()

    for i, tx := range txs {
        account, exists := se.accounts[tx.UserID]
        if !exists {
            results[i] = &SettlementResult{
                TransactionID: tx.ID,
                Success:       false,
                ErrorMessage:  "账户不存在",
                Timestamp:     time.Now(),
            }
            continue
        }

        // 处理交易逻辑
        var newBalance float64
        var success bool
        var errorMsg string

        switch tx.Type {
        case "credit":  // 入账
            newBalance = account.Balance + tx.Amount
            success = true
        case "debit":   // 出账
            if account.Balance >= tx.Amount {
                newBalance = account.Balance - tx.Amount
                success = true
            } else {
                success = false
                errorMsg = "余额不足"
                newBalance = account.Balance
            }
        }

        // 更新账户
        if success {
            account.Balance = newBalance
            account.Version++
            account.UpdatedAt = time.Now()
        }

        results[i] = &SettlementResult{
            TransactionID: tx.ID,
            Success:       success,
            NewBalance:    newBalance,
            ErrorMessage:  errorMsg,
            Timestamp:     time.Now(),
        }
    }

    return results
}
```

### FreezeAmount 方法：资金冻结

```go
func (se *SettlementEngine) FreezeAmount(userID string, amount float64) error {
    se.mutex.Lock()
    defer se.mutex.Unlock()

    account, exists := se.accounts[userID]
    if !exists {
        return fmt.Errorf("账户 %s 不存在", userID)
    }

    // 检查余额是否足够
    if account.Balance < amount {
        return fmt.Errorf("余额不足，无法冻结")
    }

    // 执行冻结操作
    account.Balance -= amount        // 减少可用余额
    account.FrozenAmount += amount   // 增加冻结金额
    account.Version++                // 版本递增
    account.UpdatedAt = time.Now()   // 更新时间

    return nil
}
```

## 使用方法

### 1. 编译运行
```bash
go run main.go
```

程序会演示：
- 创建账户
- 提交各种类型的交易
- 批量处理结算
- 显示账户余额和统计信息
- 演示资金冻结功能

### 2. 运行测试
```bash
go test -v
```

## 测试覆盖

- `TestCreateAccount`: 测试账户创建
- `TestSubmitTransaction`: 测试交易提交
- `TestCreditTransaction`: 测试入账交易
- `TestDebitTransaction`: 测试出账交易
- `TestBatchProcessing`: 测试批量处理
- `TestFreezeUnfreeze`: 测试资金冻结解冻
- `TestGetAccount`: 测试账户查询
- `TestTransactionStats`: 测试统计信息

## 性能优化

1. **批量处理**: 减少锁竞争，提高吞吐量
2. **异步队列**: 解耦交易提交和处理
3. **读写锁**: 区分读写操作，提高并发度
4. **乐观锁**: 使用版本号避免长事务

## 扩展思路

1. **持久化**: 支持数据库持久化存储
2. **分布式**: 支持多节点分布式结算
3. **事务**: 支持跨账户转账事务
4. **限流**: 添加交易频率限制
5. **监控**: 添加性能监控和告警
6. **撤销**: 支持交易撤销和冲正