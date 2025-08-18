package block_query

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetBlockByNumber 通过区块号查询区块信息
func GetBlockByNumber(client *ethclient.Client, blockNumber uint64) *types.Block {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Fatalf("Failed to retrieve block %d: %v", blockNumber, err)
	}
	return block
}

// GetLatestBlock 获取最新区块
func GetLatestBlock(client *ethclient.Client) *types.Block {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal("Failed to get latest header: ", err)
	}

	block, err := client.BlockByNumber(context.Background(), header.Number)
	if err != nil {
		log.Fatalf("Failed to retrieve block %d: %v", header.Number, err)
	}
	return block
}
