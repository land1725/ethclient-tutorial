package eth_transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"ethclient_tutorial/config"
)

// TransferETH 发送ETH转账
func TransferETH(client *ethclient.Client, privateKeyHex string, toAddress common.Address, amount float64) (common.Hash, error) {
	return TransferETHWithConfig(client, privateKeyHex, toAddress, amount, config.GlobalConfig)
}

// TransferETHWithConfig 使用配置发送ETH转账 - 支持EIP-1559
func TransferETHWithConfig(client *ethclient.Client, privateKeyHex string, toAddress common.Address, amount float64, cfg *config.Config) (common.Hash, error) {
	fmt.Println("\n=== 开始ETH转账流程 ===")

	// 1. 加载发送者私钥
	privateKeyECDSA, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return common.Hash{}, fmt.Errorf("invalid private key: %v", err)
	}
	fmt.Println("✓ 私钥加载成功")

	// 2. 获取发送者公钥和地址
	publicKey := privateKeyECDSA.Public()
	// 将公钥转换为ECDSA类型
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Hash{}, fmt.Errorf("failed to cast public key to ECDSA")
	}
	// 使用公钥生成地址
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("✓ 发送方地址: %s\n", address.Hex())
	fmt.Printf("✓ 接收方地址: %s\n", toAddress.Hex())

	// 3. 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get chain ID: %v", err)
	}
	fmt.Printf("✓ 链ID: %s\n", chainID.String())

	// 4. 根据当前地址获取nonce
	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get nonce: %v", err)
	}
	fmt.Printf("✓ Nonce: %d\n", nonce)

	// 5. 计算ETH转账金额(1 ETH = 10^18 wei)
	value := EtherToWei(amount)
	fmt.Printf("✓ 转账金额: %s ETH (%s Wei)\n",
		new(big.Float).SetFloat64(amount),
		value.String())

	// 6. 尝试使用EIP-1559动态费用交易
	fmt.Println("\n=== 创建EIP-1559动态费用交易 ===")

	// 获取当前网络的基础费用建议
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get latest header: %v", err)
	}

	// 获取建议的小费上限
	tipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		// 如果获取小费失败，使用默认值
		tipCap = big.NewInt(2e9) // 2 Gwei
		fmt.Printf("Warning: 无法获取建议小费，使用默认值: %s Gwei\n",
			new(big.Float).Quo(new(big.Float).SetInt(tipCap), big.NewFloat(1e9)))
	}

	// 正确计算GasFeeCap (基础费用的2倍 + 小费)
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(header.BaseFee, big.NewInt(2)), // 2倍基础费缓冲
		tipCap,
	)

	// 应用配置的gas价格倍数
	gasFeeCapFloat := new(big.Float).SetInt(gasFeeCap)
	multiplier := big.NewFloat(cfg.GasPriceMultiplier)
	adjustedGasFeeCap := new(big.Float).Mul(gasFeeCapFloat, multiplier)
	adjustedGasFeeCap.Int(gasFeeCap)

	fmt.Printf("✓ 基础费用: %s Gwei\n",
		new(big.Float).Quo(new(big.Float).SetInt(header.BaseFee), big.NewFloat(1e9)))
	fmt.Printf("✓ 小费上限: %s Gwei\n",
		new(big.Float).Quo(new(big.Float).SetInt(tipCap), big.NewFloat(1e9)))
	fmt.Printf("✓ 费用上限: %s Gwei\n",
		new(big.Float).Quo(new(big.Float).SetInt(gasFeeCap), big.NewFloat(1e9)))

	// 7. 动态估算GasLimit
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  address,
		To:    &toAddress,
		Value: value,
	})
	if err != nil {
		// 如果估算失败，使用配置的默认值
		gasLimit = cfg.DefaultGasLimit
		fmt.Printf("Warning: 无法估算Gas，使用默认值: %d\n", gasLimit)
	} else {
		fmt.Printf("✓ 估算Gas限制: %d\n", gasLimit)
	}

	// 8. 创建EIP-1559交易
	tx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      nil,
	}

	// 包装为通用交易类型
	newTx := types.NewTx(tx)
	fmt.Println("✓ EIP-1559交易创建成功")

	// 9. 签名交易
	fmt.Println("\n=== 签名并发送交易 ===")
	signedTx, err := types.SignTx(newTx, types.LatestSignerForChainID(chainID), privateKeyECDSA)
	if err != nil {
		// 如果EIP-1559签名失败，尝试Legacy交易
		fmt.Printf("EIP-1559交易签名失败: %v\n", err)
		fmt.Println("尝试使用Legacy交易...")

		return createLegacyTransaction(client, privateKeyECDSA, address, toAddress, value, nonce, chainID, gasLimit, cfg)
	}
	fmt.Println("✓ EIP-1559交易签名成功")

	// 10. 打印交易详情
	fmt.Println("\n=== 交易详情 ===")
	fmt.Printf("交易类型: %d (EIP-1559)\n", signedTx.Type())
	fmt.Printf("发送方: %s\n", address.Hex())
	fmt.Printf("接收方: %s\n", toAddress.Hex())
	fmt.Printf("金额: %s ETH\n", new(big.Float).Quo(new(big.Float).SetInt(value), big.NewFloat(1e18)))
	fmt.Printf("Gas限制: %d\n", gasLimit)
	fmt.Printf("Nonce: %d\n", nonce)
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())

	// 11. 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		// 如果EIP-1559交易发送失败，尝试Legacy交易
		fmt.Printf("EIP-1559交易发送失败: %v\n", err)
		fmt.Println("尝试使用Legacy交易...")

		return createLegacyTransaction(client, privateKeyECDSA, address, toAddress, value, nonce, chainID, gasLimit, cfg)
	}

	fmt.Println("✅ EIP-1559交易发送成功!")
	return signedTx.Hash(), nil
}

// createLegacyTransaction 创建Legacy交易作为后备方案
func createLegacyTransaction(client *ethclient.Client, privateKey *ecdsa.PrivateKey, fromAddress, toAddress common.Address, value *big.Int, nonce uint64, chainID *big.Int, gasLimit uint64, cfg *config.Config) (common.Hash, error) {
	fmt.Println("\n=== 创建Legacy交易 ===")

	// 获取建议gas价格
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get gas price: %v", err)
	}

	// 应用gas价格倍数
	gasPriceFloat := new(big.Float).SetInt(gasPrice)
	multiplier := big.NewFloat(cfg.GasPriceMultiplier)
	adjustedGasPrice := new(big.Float).Mul(gasPriceFloat, multiplier)

	finalGasPrice := new(big.Int)
	adjustedGasPrice.Int(finalGasPrice)

	fmt.Printf("✓ Gas价格: %s Gwei\n",
		new(big.Float).Quo(new(big.Float).SetInt(finalGasPrice), big.NewFloat(1e9)))

	// 创建Legacy交易
	txData := types.LegacyTx{
		Nonce:    nonce,
		To:       &toAddress,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: finalGasPrice,
	}

	tx := types.NewTx(&txData)

	// 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to sign legacy transaction: %v", err)
	}
	fmt.Println("✓ Legacy交易签名成功")

	// 打印交易详情
	fmt.Printf("交易类型: %d (Legacy)\n", signedTx.Type())
	fmt.Printf("Gas价格: %s Gwei\n",
		new(big.Float).Quo(new(big.Float).SetInt(finalGasPrice), big.NewFloat(1e9)))
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())

	// 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to send legacy transaction: %v", err)
	}

	fmt.Println("✅ Legacy交易发送成功!")
	return signedTx.Hash(), nil
}

// EtherToWei 将ETH转换为Wei
func EtherToWei(eth float64) *big.Int {
	wei := new(big.Float).Mul(big.NewFloat(eth), big.NewFloat(1e18))
	weiInt := new(big.Int)
	wei.Int(weiInt)
	return weiInt
}
