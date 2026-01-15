package main

import (
	"testing"
	"time"
)

func TestCreateAccount(t *testing.T) {
	engine := NewSettlementEngine()

	// 测试创建账户
	err := engine.CreateAccount("user1", 1000.0)
	if err != nil {
		t.Errorf("创建账户失败: %v", err)
	}

	// 测试重复创建
	err = engine.CreateAccount("user1", 500.0)
	if err == nil {
		t.Error("期望重复创建失败，但成功了")
	}
}

func TestSubmitTransaction(t *testing.T) {
	engine := NewSettlementEngine()
	engine.CreateAccount("user1", 1000.0)

	// 启动引擎
	engine.Start()
	defer engine.Stop()

	// 测试提交有效交易
	tx := &Transaction{
		UserID:      "user1",
		Amount:      100.0,
		Type:        "debit",
		Description: "测试交易",
	}

	err := engine.SubmitTransaction(tx)
	if err != nil {
		t.Errorf("提交交易失败: %v", err)
	}

	// 等待处理
	time.Sleep(100 * time.Millisecond)

	// 测试提交无效交易
	invalidTx := &Transaction{
		UserID: "",
		Amount: 50.0,
		Type:   "debit",
	}

	err = engine.SubmitTransaction(invalidTx)
	if err == nil {
		t.Error("期望无效交易提交失败，但成功了")
	}
}

func TestCreditTransaction(t *testing.T) {
	engine := NewSettlementEngine()
	engine.CreateAccount("user1", 1000.0)

	engine.Start()
	defer engine.Stop()

	// 入账交易
	tx := &Transaction{
		UserID: "user1",
		Amount: 500.0,
		Type:   "credit",
	}

	engine.SubmitTransaction(tx)
	time.Sleep(100 * time.Millisecond)

	account, _ := engine.GetAccount("user1")
	expectedBalance := 1500.0
	if account.Balance != expectedBalance {
		t.Errorf("期望余额%.2f，实际余额%.2f", expectedBalance, account.Balance)
	}
}

func TestDebitTransaction(t *testing.T) {
	engine := NewSettlementEngine()
	engine.CreateAccount("user1", 1000.0)

	engine.Start()
	defer engine.Stop()

	// 出账交易 - 成功
	tx1 := &Transaction{
		UserID: "user1",
		Amount: 300.0,
		Type:   "debit",
	}

	engine.SubmitTransaction(tx1)
	time.Sleep(100 * time.Millisecond)

	account, _ := engine.GetAccount("user1")
	if account.Balance != 700.0 {
		t.Errorf("期望余额700.0，实际余额%.2f", account.Balance)
	}

	// 出账交易 - 失败（余额不足）
	tx2 := &Transaction{
		UserID: "user1",
		Amount: 1000.0,
		Type:   "debit",
	}

	engine.SubmitTransaction(tx2)
	time.Sleep(100 * time.Millisecond)

	// 余额应该不变
	account, _ = engine.GetAccount("user1")
	if account.Balance != 700.0 {
		t.Errorf("余额不足时余额应该不变，实际余额%.2f", account.Balance)
	}
}

func TestBatchProcessing(t *testing.T) {
	engine := NewSettlementEngine()
	engine.CreateAccount("user1", 1000.0)
	engine.CreateAccount("user2", 500.0)

	// 设置小批量大小以便测试
	engine.batchSize = 3

	engine.Start()
	defer engine.Stop()

	// 提交多笔交易
	transactions := []*Transaction{
		{UserID: "user1", Amount: 100.0, Type: "debit"},
		{UserID: "user2", Amount: 50.0, Type: "credit"},
		{UserID: "user1", Amount: 50.0, Type: "debit"},
		{UserID: "user2", Amount: 30.0, Type: "credit"},
	}

	for _, tx := range transactions {
		engine.SubmitTransaction(tx)
	}

	// 等待批处理完成
	time.Sleep(200 * time.Millisecond)

	account1, _ := engine.GetAccount("user1")
	account2, _ := engine.GetAccount("user2")

	// user1: 1000 - 100 - 50 = 850
	if account1.Balance != 850.0 {
		t.Errorf("期望user1余额850.0，实际%.2f", account1.Balance)
	}

	// user2: 500 + 50 + 30 = 580
	if account2.Balance != 580.0 {
		t.Errorf("期望user2余额580.0，实际%.2f", account2.Balance)
	}
}

func TestFreezeUnfreeze(t *testing.T) {
	engine := NewSettlementEngine()
	engine.CreateAccount("user1", 1000.0)

	// 测试冻结
	err := engine.FreezeAmount("user1", 200.0)
	if err != nil {
		t.Errorf("冻结金额失败: %v", err)
	}

	account, _ := engine.GetAccount("user1")
	if account.Balance != 800.0 || account.FrozenAmount != 200.0 {
		t.Errorf("冻结后余额%.2f，冻结金额%.2f", account.Balance, account.FrozenAmount)
	}

	// 测试解冻
	err = engine.UnfreezeAmount("user1", 100.0)
	if err != nil {
		t.Errorf("解冻金额失败: %v", err)
	}

	account, _ = engine.GetAccount("user1")
	if account.Balance != 900.0 || account.FrozenAmount != 100.0 {
		t.Errorf("解冻后余额%.2f，冻结金额%.2f", account.Balance, account.FrozenAmount)
	}

	// 测试冻结不足
	err = engine.FreezeAmount("user1", 1000.0)
	if err == nil {
		t.Error("期望余额不足时冻结失败")
	}

	// 测试解冻过多
	err = engine.UnfreezeAmount("user1", 200.0)
	if err == nil {
		t.Error("期望解冻过多时失败")
	}
}

func TestGetAccount(t *testing.T) {
	engine := NewSettlementEngine()

	// 测试获取不存在的账户
	_, err := engine.GetAccount("nonexistent")
	if err == nil {
		t.Error("期望获取不存在账户失败")
	}

	// 创建账户后获取
	engine.CreateAccount("user1", 500.0)
	account, err := engine.GetAccount("user1")
	if err != nil {
		t.Errorf("获取账户失败: %v", err)
	}

	if account.Balance != 500.0 {
		t.Errorf("期望余额500.0，实际%.2f", account.Balance)
	}
}

func TestTransactionStats(t *testing.T) {
	engine := NewSettlementEngine()
	engine.CreateAccount("user1", 1000.0)
	engine.CreateAccount("user2", 500.0)

	engine.Start()
	defer engine.Stop()

	// 提交一些交易
	transactions := []*Transaction{
		{UserID: "user1", Amount: 100.0, Type: "debit"},
		{UserID: "user2", Amount: 50.0, Type: "credit"},
	}

	for _, tx := range transactions {
		engine.SubmitTransaction(tx)
	}

	time.Sleep(100 * time.Millisecond)

	stats := engine.GetTransactionStats()

	if stats["total_accounts"] != 2 {
		t.Errorf("期望2个账户，实际%d个", stats["total_accounts"])
	}

	if stats["total_transactions"] != 2 {
		t.Errorf("期望2笔交易，实际%d笔", stats["total_transactions"])
	}
}