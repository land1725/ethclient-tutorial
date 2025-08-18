package block_subscription

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// BlockSubscription 订阅新区块，返回第一个接收到的区块（用于演示）
func BlockSubscription(client *ethclient.Client) *types.Block {
	//定义一个通道获取订阅头
	ch := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), ch)
	if err != nil {
		log.Fatalf("Failed to subscribe to new blocks: %v", err)
	}
	defer sub.Unsubscribe()

	fmt.Println("🔔 开始监听新区块...")

	// 监听新块
	for {
		select {
		case err := <-sub.Err():
			log.Fatalf("Subscription error: %v", err)
		case header := <-ch:
			block, err := client.BlockByNumber(context.Background(), header.Number)
			if err != nil {
				log.Fatalf("Failed to retrieve block %d: %v", header.Number, err)
			}
			fmt.Printf("📦 接收到新区块: #%d, 哈希: %s\n", block.NumberU64(), block.Hash().Hex())
			fmt.Printf("   交易数量: %d\n", len(block.Transactions()))
			fmt.Printf("   时间戳: %d\n", block.Time())
			return block
		}
	}
}
