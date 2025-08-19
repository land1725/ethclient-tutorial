package contract_events

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EventWatcher 事件监听器结构体
type EventWatcher struct {
	client          *ethclient.Client
	contractAddress common.Address
	contractABI     abi.ABI
	subscription    ethereum.Subscription
	logChan         chan types.Log
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewEventWatcher 创建新的事件监听器
func NewEventWatcher(client *ethclient.Client, contractAddress common.Address) (*EventWatcher, error) {
	// 读取并解析ABI文件
	abiData, err := os.ReadFile("contracts/compiled/MyToken.abi")
	if err != nil {
		return nil, fmt.Errorf("读取ABI文件失败: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiData)))
	if err != nil {
		return nil, fmt.Errorf("解析ABI失败: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &EventWatcher{
		client:          client,
		contractAddress: contractAddress,
		contractABI:     parsedABI,
		logChan:         make(chan types.Log),
		ctx:             ctx,
		cancel:          cancel,
	}, nil
}

// StartWatching 开始监听合约事件
func (w *EventWatcher) StartWatching() error {
	fmt.Printf("🔍 开始监听合约事件: %s\n", w.contractAddress.Hex())

	// 创建过滤器查询，监听指定合约的所有事件
	query := ethereum.FilterQuery{
		Addresses: []common.Address{w.contractAddress},
	}

	// 订阅日志
	sub, err := w.client.SubscribeFilterLogs(w.ctx, query, w.logChan)
	if err != nil {
		return fmt.Errorf("订阅日志失败: %v", err)
	}

	w.subscription = sub

	fmt.Println("✅ 事件订阅成功，开始监听...")

	// 启动事件处理协程
	go w.handleEvents()

	return nil
}

// handleEvents 处理接收到的事件
func (w *EventWatcher) handleEvents() {
	for {
		select {
		case err := <-w.subscription.Err():
			log.Printf("❌ 事件订阅错误: %v", err)
			return

		case vLog := <-w.logChan:
			w.processEvent(vLog)

		case <-w.ctx.Done():
			fmt.Println("🛑 事件监听已停止")
			return
		}
	}
}

// processEvent 处理单个事件
func (w *EventWatcher) processEvent(vLog types.Log) {
	fmt.Println("\n📧 收到新事件:")
	fmt.Printf("   区块号: #%d\n", vLog.BlockNumber)
	fmt.Printf("   交易哈希: %s\n", vLog.TxHash.Hex())
	fmt.Printf("   合约地址: %s\n", vLog.Address.Hex())
	fmt.Printf("   事件索引: %d\n", vLog.Index)

	// 解析事件
	if len(vLog.Topics) == 0 {
		fmt.Println("   ⚠️ 事件没有主题")
		return
	}

	eventSig := vLog.Topics[0]
	fmt.Printf("   事件签名: %s\n", eventSig.Hex())

	// 根据事件签名解析不同类型的事件
	switch eventSig {
	case w.contractABI.Events["Transfer"].ID:
		w.parseTransferEvent(vLog)
	case w.contractABI.Events["Approval"].ID:
		w.parseApprovalEvent(vLog)
	case w.contractABI.Events["Paused"].ID:
		w.parsePausedEvent(vLog)
	case w.contractABI.Events["Unpaused"].ID:
		w.parseUnpausedEvent(vLog)
	case w.contractABI.Events["OwnershipTransferred"].ID:
		w.parseOwnershipTransferredEvent(vLog)
	default:
		fmt.Printf("   ⚠️ 未知事件类型: %s\n", eventSig.Hex())
		w.parseUnknownEvent(vLog)
	}
}

// parseTransferEvent 解析Transfer事件
func (w *EventWatcher) parseTransferEvent(vLog types.Log) {
	fmt.Println("   📤 Transfer 事件:")

	type TransferEvent struct {
		From  common.Address
		To    common.Address
		Value *big.Int
	}

	var transferEvent TransferEvent
	err := w.contractABI.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
	if err != nil {
		fmt.Printf("      ❌ 解析Transfer事件失败: %v\n", err)
		return
	}

	// Topics[1] = from, Topics[2] = to (indexed parameters)
	if len(vLog.Topics) >= 3 {
		transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
		transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
	}

	fmt.Printf("      发送方: %s\n", transferEvent.From.Hex())
	fmt.Printf("      接收方: %s\n", transferEvent.To.Hex())
	fmt.Printf("      金额: %s\n", transferEvent.Value.String())

	// 转换为可读的代币数量（假设18位小数）
	tokenAmount := new(big.Float).Quo(new(big.Float).SetInt(transferEvent.Value), big.NewFloat(1e18))
	fmt.Printf("      金额(代币): %s\n", tokenAmount.String())
}

// parseApprovalEvent 解析Approval事件
func (w *EventWatcher) parseApprovalEvent(vLog types.Log) {
	fmt.Println("   ✅ Approval 事件:")

	type ApprovalEvent struct {
		Owner   common.Address
		Spender common.Address
		Value   *big.Int
	}

	var approvalEvent ApprovalEvent
	err := w.contractABI.UnpackIntoInterface(&approvalEvent, "Approval", vLog.Data)
	if err != nil {
		fmt.Printf("      ❌ 解析Approval事件失败: %v\n", err)
		return
	}

	// Topics[1] = owner, Topics[2] = spender (indexed parameters)
	if len(vLog.Topics) >= 3 {
		approvalEvent.Owner = common.HexToAddress(vLog.Topics[1].Hex())
		approvalEvent.Spender = common.HexToAddress(vLog.Topics[2].Hex())
	}

	fmt.Printf("      所有者: %s\n", approvalEvent.Owner.Hex())
	fmt.Printf("      被授权者: %s\n", approvalEvent.Spender.Hex())
	fmt.Printf("      授权金额: %s\n", approvalEvent.Value.String())
}

// parsePausedEvent 解析Paused事件
func (w *EventWatcher) parsePausedEvent(vLog types.Log) {
	fmt.Println("   ⏸️ Paused 事件:")

	type PausedEvent struct {
		Account common.Address
	}

	var pausedEvent PausedEvent
	err := w.contractABI.UnpackIntoInterface(&pausedEvent, "Paused", vLog.Data)
	if err != nil {
		fmt.Printf("      ❌ 解析Paused事件失败: %v\n", err)
		return
	}

	fmt.Printf("      暂停者: %s\n", pausedEvent.Account.Hex())
}

// parseUnpausedEvent 解析Unpaused事件
func (w *EventWatcher) parseUnpausedEvent(vLog types.Log) {
	fmt.Println("   ▶️ Unpaused 事件:")

	type UnpausedEvent struct {
		Account common.Address
	}

	var unpausedEvent UnpausedEvent
	err := w.contractABI.UnpackIntoInterface(&unpausedEvent, "Unpaused", vLog.Data)
	if err != nil {
		fmt.Printf("      ❌ 解析Unpaused事件失败: %v\n", err)
		return
	}

	fmt.Printf("      恢复者: %s\n", unpausedEvent.Account.Hex())
}

// parseOwnershipTransferredEvent 解析OwnershipTransferred事件
func (w *EventWatcher) parseOwnershipTransferredEvent(vLog types.Log) {
	fmt.Println("   👑 OwnershipTransferred 事件:")

	type OwnershipTransferredEvent struct {
		PreviousOwner common.Address
		NewOwner      common.Address
	}

	var ownershipEvent OwnershipTransferredEvent
	err := w.contractABI.UnpackIntoInterface(&ownershipEvent, "OwnershipTransferred", vLog.Data)
	if err != nil {
		fmt.Printf("      ❌ 解析OwnershipTransferred事件失败: %v\n", err)
		return
	}

	// Topics[1] = previousOwner, Topics[2] = newOwner (indexed parameters)
	if len(vLog.Topics) >= 3 {
		ownershipEvent.PreviousOwner = common.HexToAddress(vLog.Topics[1].Hex())
		ownershipEvent.NewOwner = common.HexToAddress(vLog.Topics[2].Hex())
	}

	fmt.Printf("      前任所有者: %s\n", ownershipEvent.PreviousOwner.Hex())
	fmt.Printf("      新所有者: %s\n", ownershipEvent.NewOwner.Hex())
}

// parseUnknownEvent 解析未知事件
func (w *EventWatcher) parseUnknownEvent(vLog types.Log) {
	fmt.Println("   ❓ 未知事件:")
	fmt.Printf("      主题数量: %d\n", len(vLog.Topics))
	for i, topic := range vLog.Topics {
		fmt.Printf("      主题[%d]: %s\n", i, topic.Hex())
	}
	fmt.Printf("      数据长度: %d bytes\n", len(vLog.Data))
	if len(vLog.Data) > 0 {
		fmt.Printf("      数据: %x\n", vLog.Data)
	}
}

// Stop 停止事件监听
func (w *EventWatcher) Stop() {
	fmt.Println("🛑 正在停止事件监听...")
	if w.subscription != nil {
		w.subscription.Unsubscribe()
	}
	if w.cancel != nil {
		w.cancel()
	}
	close(w.logChan)
	fmt.Println("✅ 事件监听已停止")
}

// WatchContractEvents 监听合约事件的便捷函数
func WatchContractEvents(client *ethclient.Client, contractAddress common.Address) (*EventWatcher, error) {
	watcher, err := NewEventWatcher(client, contractAddress)
	if err != nil {
		return nil, fmt.Errorf("创建事件监听器失败: %v", err)
	}

	err = watcher.StartWatching()
	if err != nil {
		return nil, fmt.Errorf("启动事件监听失败: %v", err)
	}

	return watcher, nil
}
