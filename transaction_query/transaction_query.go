package transaction_query

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetTransaction 通过交易哈希查询交易详情
func GetTransaction(client *ethclient.Client, txHash common.Hash) *types.Transaction {
	tx, pending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		log.Fatalf("Failed to retrieve transaction %s: %v", txHash.Hex(), err)
	}
	if pending {
		log.Println("Transaction is pending in mempool")
	}
	return tx
}

// GetTransactionInfo 获取交易详细信息
func GetTransactionInfo(client *ethclient.Client, txHash common.Hash) *TxInfo {
	tx := GetTransaction(client, txHash)

	// 获取交易收据来获取发送者信息
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		log.Fatal("Failed to get transaction receipt: ", err)
	}

	var fromAddress string
	if receipt != nil {
		// 通过区块和交易索引获取完整交易信息
		block, err := client.BlockByHash(context.Background(), receipt.BlockHash)
		if err == nil && int(receipt.TransactionIndex) < len(block.Transactions()) {
			blockTx := block.Transactions()[receipt.TransactionIndex]
			signer := types.LatestSignerForChainID(blockTx.ChainId())
			sender, err := types.Sender(signer, blockTx)
			if err == nil {
				fromAddress = sender.Hex()
			}
		}
	}

	toAddress := ""
	if tx.To() != nil {
		toAddress = tx.To().Hex()
	}

	return &TxInfo{
		Hash:     tx.Hash().Hex(),
		From:     fromAddress,
		To:       toAddress,
		Value:    tx.Value(),
		GasPrice: tx.GasPrice(),
		GasLimit: tx.Gas(),
		Nonce:    tx.Nonce(),
		Data:     tx.Data(),
		ChainID:  tx.ChainId(),
	}
}

// TxInfo 交易信息结构体
type TxInfo struct {
	Hash     string   `json:"hash"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Value    *big.Int `json:"value"`
	GasPrice *big.Int `json:"gasPrice"`
	GasLimit uint64   `json:"gasLimit"`
	Nonce    uint64   `json:"nonce"`
	Data     []byte   `json:"data"`
	ChainID  *big.Int `json:"chainId"`
}
