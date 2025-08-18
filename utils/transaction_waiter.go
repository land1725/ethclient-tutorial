package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TransactionStatus 交易状态
type TransactionStatus struct {
	Success     bool
	BlockNumber uint64
	GasUsed     uint64
	TxHash      common.Hash
	Receipt     *types.Receipt
}

// WaitForTransaction 等待交易确认
func WaitForTransaction(client *ethclient.Client, txHash common.Hash, confirmations uint64, timeout time.Duration) (*TransactionStatus, error) {
	fmt.Printf("⏳ 等待交易确认: %s\n", txHash.Hex())
	fmt.Printf("   需要确认数: %d, 超时时间: %v\n", confirmations, timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second) // 每3秒检查一次
	defer ticker.Stop()

	var receipt *types.Receipt
	var err error

	// 首先等待交易被包含在区块中
	fmt.Print("等待交易被挖矿")
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("等待交易超时: %v", ctx.Err())
		case <-ticker.C:
			receipt, err = client.TransactionReceipt(context.Background(), txHash)
			if err == nil {
				fmt.Printf("\n✓ 交易已被包含在区块 #%d 中\n", receipt.BlockNumber.Uint64())
				goto receiptFound // 使用 goto 跳出外层循环
			}
			fmt.Print(".")
		}
	}

receiptFound:
	// 检查交易状态
	status := &TransactionStatus{
		Success:     receipt.Status == types.ReceiptStatusSuccessful,
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
		TxHash:      txHash,
		Receipt:     receipt,
	}

	if !status.Success {
		return status, fmt.Errorf("交易执行失败")
	}

	fmt.Printf("✓ 交易执行成功, Gas使用: %d\n", status.GasUsed)

	// 如果只需要1个确认，直接返回
	if confirmations <= 1 {
		return status, nil
	}

	// 如果需要等待更多确认数，继续等待
	fmt.Printf("⏳ 等待额外的 %d 个确认...\n", confirmations-1)
	targetBlockNumber := status.BlockNumber + confirmations - 1

	for {
		select {
		case <-ctx.Done():
			// 即使超时，如果交易已经成功，也返回状态而不是错误
			fmt.Printf("⚠️ 等待额外确认超时，但交易已执行成功\n")
			return status, nil
		case <-ticker.C:
			currentBlock, err := client.BlockNumber(context.Background())
			if err != nil {
				fmt.Printf("获取当前区块号失败: %v\n", err)
				continue
			}

			if currentBlock >= targetBlockNumber {
				fmt.Printf("✅ 已获得 %d 个确认 (当前区块: #%d)\n", confirmations, currentBlock)
				return status, nil
			}

			confirmedBlocks := currentBlock - status.BlockNumber + 1
			fmt.Printf("   确认进度: %d/%d (当前区块: #%d)\n", confirmedBlocks, confirmations, currentBlock)
		}
	}
}

// WaitForTransactionQuick 快速等待交易（只等待1个确认）
func WaitForTransactionQuick(client *ethclient.Client, txHash common.Hash) (*TransactionStatus, error) {
	return WaitForTransaction(client, txHash, 1, 3*time.Minute) // 增加到3分钟
}

// WaitForTransactionSafe 安全等待交易（等待3个确认）
func WaitForTransactionSafe(client *ethclient.Client, txHash common.Hash) (*TransactionStatus, error) {
	return WaitForTransaction(client, txHash, 3, 10*time.Minute) // 增加到10分钟
}

// WaitForTransactionDeploy 专门用于合约部署的等待函数（更长的超时时间）
func WaitForTransactionDeploy(client *ethclient.Client, txHash common.Hash) (*TransactionStatus, error) {
	return WaitForTransaction(client, txHash, 2, 8*time.Minute) // 等待2个确认，8分钟超时
}
