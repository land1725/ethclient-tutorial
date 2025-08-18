package receipt_query

import (
	"context"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetTransactionReceipt 获取交易收据
func GetTransactionReceipt(client *ethclient.Client, txHash common.Hash) *types.Receipt {
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatalf("Failed to retrieve receipt for tx %s: %v", txHash.Hex(), err)
	}
	return receipt
}

// CheckTransactionStatus 检查交易状态
func CheckTransactionStatus(client *ethclient.Client, txHash common.Hash) string {
	receipt := GetTransactionReceipt(client, txHash)

	switch receipt.Status {
	case types.ReceiptStatusSuccessful:
		return "Successful"
	case types.ReceiptStatusFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}
