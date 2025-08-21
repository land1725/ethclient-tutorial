package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 连接到本地私有开发链
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal("Failed to connect to the Ethereum client: ", err)
	}
	defer client.Close()

	fmt.Println("🎉 成功连接到私有开发链!")

	// 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal("Failed to get chain ID: ", err)
	}
	fmt.Printf("🆔 Chain ID: %d\n", chainID)

	// 获取最新区块号
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatal("Failed to get block number: ", err)
	}
	fmt.Printf("📦 最新区块号: %d\n", blockNumber)

	// 获取最新区块信息
	block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumber)))
	if err != nil {
		log.Fatal("Failed to get block: ", err)
	}
	fmt.Printf("📅 区块时间: %d\n", block.Time())
	fmt.Printf("⛽ Gas限制: %d\n", block.GasLimit())
	fmt.Printf("📊 交易数量: %d\n", len(block.Transactions()))

	// 获取网络版本
	networkID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal("Failed to get network ID: ", err)
	}
	fmt.Printf("🌐 网络ID: %d\n", networkID)

	fmt.Println("\n✅ 私有开发链运行正常，可以开始开发了！")
	fmt.Println("🔗 HTTP RPC: http://localhost:8545")
	fmt.Println("📡 WebSocket: ws://localhost:8546")
}
