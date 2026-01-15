package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// Transaction 交易记录
type Transaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"` // debit, credit
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// Account 账户信息
type Account struct {
	UserID      string  `json:"user_id"`
	Balance     float64 `json:"balance"`
	FrozenAmount float64 `json:"frozen_amount"`
	Version     int64   `json:"version"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SettlementResult 结算结果
type SettlementResult struct {
	TransactionID string    `json:"transaction_id"`
	Success       bool      `json:"success"`
	NewBalance    float64   `json:"new_balance"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// SettlementEngine 结算引擎
type SettlementEngine struct {
	accounts   map[string]*Account
	transactions []Transaction
	mutex      sync.RWMutex
	settlementChan chan *Transaction
	stopChan   chan bool
	batchSize  int
	batchTimeout time.Duration
}

// NewSettlementEngine 创建结算引擎
func NewSettlementEngine() *SettlementEngine {
	return &SettlementEngine{
		accounts:       make(map[string]*Account),
		transactions:   make([]Transaction, 0),
		settlementChan: make(chan *Transaction, 1000),
		stopChan:       make(chan bool),
		batchSize:      100,
		batchTimeout:   5 * time.Second,
	}
}

// CreateAccount 创建账户
func (se *SettlementEngine) CreateAccount(userID string, initialBalance float64) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	if _, exists := se.accounts[userID]; exists {
		return fmt.Errorf("账户 %s 已存在", userID)
	}

	se.accounts[userID] = &Account{
		UserID:      userID,
		Balance:     initialBalance,
		FrozenAmount: 0,
		Version:     1,
		UpdatedAt:   time.Now(),
	}

	fmt.Printf("创建账户: %s, 初始余额: %.2f\n", userID, initialBalance)
	return nil
}

// SubmitTransaction 提交交易
func (se *SettlementEngine) SubmitTransaction(tx *Transaction) error {
	if tx.UserID == "" || tx.Amount <= 0 {
		return fmt.Errorf("无效的交易参数")
	}

	tx.ID = fmt.Sprintf("tx_%d", time.Now().UnixNano())
	tx.Timestamp = time.Now()
	tx.Status = "pending"

	se.mutex.Lock()
	se.transactions = append(se.transactions, *tx)
	se.mutex.Unlock()

	select {
	case se.settlementChan <- tx:
		fmt.Printf("交易已提交: %s, 用户: %s, 金额: %.2f\n", tx.ID, tx.UserID, tx.Amount)
		return nil
	default:
		return fmt.Errorf("结算队列已满")
	}
}

// processTransaction 处理单个交易
func (se *SettlementEngine) processTransaction(tx *Transaction) *SettlementResult {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	account, exists := se.accounts[tx.UserID]
	if !exists {
		return &SettlementResult{
			TransactionID: tx.ID,
			Success:       false,
			ErrorMessage:  "账户不存在",
			Timestamp:     time.Now(),
		}
	}

	var newBalance float64
	var success bool
	var errorMsg string

	switch tx.Type {
	case "credit": // 入账
		newBalance = account.Balance + tx.Amount
		success = true
	case "debit": // 出账
		if account.Balance >= tx.Amount {
			newBalance = account.Balance - tx.Amount
			success = true
		} else {
			success = false
			errorMsg = "余额不足"
			newBalance = account.Balance
		}
	default:
		success = false
		errorMsg = "无效的交易类型"
		newBalance = account.Balance
	}

	if success {
		account.Balance = newBalance
		account.Version++
		account.UpdatedAt = time.Now()
	}

	return &SettlementResult{
		TransactionID: tx.ID,
		Success:       success,
		NewBalance:    newBalance,
		ErrorMessage:  errorMsg,
		Timestamp:     time.Now(),
	}
}

// batchProcessTransactions 批量处理交易
func (se *SettlementEngine) batchProcessTransactions(txs []*Transaction) []*SettlementResult {
	results := make([]*SettlementResult, len(txs))

	se.mutex.Lock()
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

		var newBalance float64
		var success bool
		var errorMsg string

		switch tx.Type {
		case "credit":
			newBalance = account.Balance + tx.Amount
			success = true
		case "debit":
			if account.Balance >= tx.Amount {
				newBalance = account.Balance - tx.Amount
				success = true
			} else {
				success = false
				errorMsg = "余额不足"
				newBalance = account.Balance
			}
		default:
			success = false
			errorMsg = "无效的交易类型"
			newBalance = account.Balance
		}

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

// Start 启动结算引擎
func (se *SettlementEngine) Start() {
	fmt.Println("结算引擎已启动")

	go se.processSettlementQueue()
}

// processSettlementQueue 处理结算队列
func (se *SettlementEngine) processSettlementQueue() {
	batch := make([]*Transaction, 0, se.batchSize)
	timer := time.NewTimer(se.batchTimeout)
	defer timer.Stop()

	for {
		select {
		case tx := <-se.settlementChan:
			batch = append(batch, tx)

			// 达到批处理大小时立即处理
			if len(batch) >= se.batchSize {
				se.processBatch(batch)
				batch = batch[:0] // 清空批次
				timer.Reset(se.batchTimeout)
			}

		case <-timer.C:
			// 超时处理当前批次
			if len(batch) > 0 {
				se.processBatch(batch)
				batch = batch[:0]
			}
			timer.Reset(se.batchTimeout)

		case <-se.stopChan:
			// 处理剩余的批次
			if len(batch) > 0 {
				se.processBatch(batch)
			}
			fmt.Println("结算队列处理已停止")
			return
		}
	}
}

// processBatch 处理批次交易
func (se *SettlementEngine) processBatch(batch []*Transaction) {
	fmt.Printf("开始处理批次交易，数量: %d\n", len(batch))

	results := se.batchProcessTransactions(batch)

	successCount := 0
	failCount := 0

	for _, result := range results {
		if result.Success {
			successCount++
			fmt.Printf("交易成功: %s, 新余额: %.2f\n", result.TransactionID, result.NewBalance)
		} else {
			failCount++
			fmt.Printf("交易失败: %s, 原因: %s\n", result.TransactionID, result.ErrorMessage)
		}
	}

	fmt.Printf("批次处理完成，成功: %d, 失败: %d\n", successCount, failCount)
}

// Stop 停止结算引擎
func (se *SettlementEngine) Stop() {
	close(se.stopChan)
}

// GetAccount 获取账户信息
func (se *SettlementEngine) GetAccount(userID string) (*Account, error) {
	se.mutex.RLock()
	defer se.mutex.RUnlock()

	account, exists := se.accounts[userID]
	if !exists {
		return nil, fmt.Errorf("账户 %s 不存在", userID)
	}

	return account, nil
}

// GetTransactionStats 获取交易统计
func (se *SettlementEngine) GetTransactionStats() map[string]int {
	se.mutex.RLock()
	defer se.mutex.RUnlock()

	stats := map[string]int{
		"total_accounts": len(se.accounts),
		"total_transactions": len(se.transactions),
		"pending_transactions": 0,
		"processed_transactions": 0,
	}

	for _, tx := range se.transactions {
		if tx.Status == "pending" {
			stats["pending_transactions"]++
		} else {
			stats["processed_transactions"]++
		}
	}

	return stats
}

// FreezeAmount 冻结金额
func (se *SettlementEngine) FreezeAmount(userID string, amount float64) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	account, exists := se.accounts[userID]
	if !exists {
		return fmt.Errorf("账户 %s 不存在", userID)
	}

	if account.Balance < amount {
		return fmt.Errorf("余额不足，无法冻结")
	}

	account.Balance -= amount
	account.FrozenAmount += amount
	account.Version++
	account.UpdatedAt = time.Now()

	fmt.Printf("冻结金额: 用户%s, 金额%.2f\n", userID, amount)
	return nil
}

// UnfreezeAmount 解冻金额
func (se *SettlementEngine) UnfreezeAmount(userID string, amount float64) error {
	se.mutex.Lock()
	defer se.mutex.Unlock()

	account, exists := se.accounts[userID]
	if !exists {
		return fmt.Errorf("账户 %s 不存在", userID)
	}

	if account.FrozenAmount < amount {
		return fmt.Errorf("冻结金额不足")
	}

	account.Balance += amount
	account.FrozenAmount -= amount
	account.Version++
	account.UpdatedAt = time.Now()

	fmt.Printf("解冻金额: 用户%s, 金额%.2f\n", userID, amount)
	return nil
}

func main() {
	// 创建结算引擎
	engine := NewSettlementEngine()

	// 创建账户
	engine.CreateAccount("user1", 1000.0)
	engine.CreateAccount("user2", 500.0)

	// 启动结算引擎
	engine.Start()

	// 提交交易
	transactions := []*Transaction{
		{UserID: "user1", Amount: 100.0, Type: "debit", Description: "购买商品"},
		{UserID: "user2", Amount: 50.0, Type: "credit", Description: "充值"},
		{UserID: "user1", Amount: 200.0, Type: "debit", Description: "转账"},
		{UserID: "user1", Amount: 1500.0, Type: "debit", Description: "大额消费"}, // 这笔会失败
		{UserID: "user2", Amount: 30.0, Type: "credit", Description: "奖金"},
	}

	for _, tx := range transactions {
		if err := engine.SubmitTransaction(tx); err != nil {
			log.Printf("提交交易失败: %v", err)
		}
	}

	// 等待处理完成
	time.Sleep(2 * time.Second)

	// 显示账户信息
	user1Account, _ := engine.GetAccount("user1")
	user2Account, _ := engine.GetAccount("user2")

	fmt.Printf("\n=== 账户信息 ===\n")
	fmt.Printf("用户1余额: %.2f, 冻结金额: %.2f\n", user1Account.Balance, user1Account.FrozenAmount)
	fmt.Printf("用户2余额: %.2f, 冻结金额: %.2f\n", user2Account.Balance, user2Account.FrozenAmount)

	// 显示统计信息
	stats := engine.GetTransactionStats()
	fmt.Printf("\n=== 统计信息 ===\n")
	fmt.Printf("账户总数: %d\n", stats["total_accounts"])
	fmt.Printf("交易总数: %d\n", stats["total_transactions"])
	fmt.Printf("待处理交易: %d\n", stats["pending_transactions"])
	fmt.Printf("已处理交易: %d\n", stats["processed_transactions"])

	// 演示冻结功能
	engine.FreezeAmount("user1", 100.0)
	time.Sleep(1 * time.Second)
	engine.UnfreezeAmount("user1", 50.0)

	// 最终账户状态
	finalAccount, _ := engine.GetAccount("user1")
	fmt.Printf("\n最终用户1余额: %.2f, 冻结金额: %.2f\n", finalAccount.Balance, finalAccount.FrozenAmount)

	engine.Stop()
}