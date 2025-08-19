package task1

import (
	"context"
	"crypto/ecdsa"
	"ethclient_tutorial/config"
	"ethclient_tutorial/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// 查询指定区块号的区块信息，包括区块的哈希、时间戳、交易数量等
func QueryBlockByNum(client *ethclient.Client, num *big.Int) {
	var block *types.Block
	//如果传入的num为nil 直接查询最新的区块信息
	//否则查询指定区块号的区块信息
	if num == nil {
		block, _ = client.BlockByNumber(context.Background(), nil)
	} else {
		block, _ = client.BlockByNumber(context.Background(), num)
	}
	fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
	fmt.Printf("区块时间戳: %d\n", block.Time())
	fmt.Printf("区块交易数量: %d\n", len(block.Transactions()))
	//for _, tx := range block.Transactions() {
	//	fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
	//	fmt.Printf("交易接收方: %s\n", tx.To().Hex())
	//}

}

// TransferETH 发送ETH转账
func TransferETH(client *ethclient.Client, privateKeyHex string, toAddress common.Address, amount float64) (common.Hash, error) {
	return TransferETHWithConfig(client, privateKeyHex, toAddress, amount, config.GlobalConfig)
}

// TransferETHWithConfig 使用配置发送ETH转账 - 支持EIP-1559
func TransferETHWithConfig(client *ethclient.Client, privateKeyHex string, toAddress common.Address, amount float64, cfg *config.Config) (common.Hash, error) {
	fmt.Println("\n=== 开始ETH转账流程 ===")
	//根据私钥转换为ECDSA
	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return common.Hash{}, fmt.Errorf("invalid private key: %v", err)
	}
	//根据ECDSA获取公钥
	publicKey := privateKeyECDSA.Public()
	//将公钥转换为ECDSA类型
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Hash{}, fmt.Errorf("failed to cast public key to ECDSA")
	}
	//使用公钥生成地址
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("✓ 发送方地址: %s\n", fromAddress.Hex())
	fmt.Printf("✓ 接收方地址: %s\n", fromAddress.Hex())
	//使用client 获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce: %v", err)
	}
	fmt.Printf("✓ Nonce: %d\n", nonce)
	//使用client 获取最新header
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get latest header: %v", err)
	}
	//获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get chain ID: %v", err)
	}
	//获取BaseFee
	baseFee := header.BaseFee
	//获取SuggestGasTipCap
	tipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		fmt.Printf("Warning: 获取小费建议失败，使用默认值: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(tipCap), big.NewFloat(1e9)))
	}
	//计算gasFeeCap = baseFee * 2 + tipCap
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		tipCap,
	)
	//获取gasLimit
	gasLimit := uint64(21000) // ETH转账的标准Gas限制
	// 8. 创建EIP-1559交易
	tx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     utils.TokenToWei(amount, 18),
		// 转账金额转换为Wei
		Data: nil,
	}
	//使用私钥 对 交易和 chainId 签名
	signedTx, err := types.SignTx(types.NewTx(tx), types.LatestSignerForChainID(chainID), privateKeyECDSA)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign transaction: %v", err)
	}
	fmt.Printf("✓ 交易已签名: %s\n", signedTx.Hash().Hex())
	//发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send transaction: %v", err)
	}
	return signedTx.Hash(), nil
}
