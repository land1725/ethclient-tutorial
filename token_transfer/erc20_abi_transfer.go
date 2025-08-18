package token_transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"ethclient_tutorial/contracts"
)

// TransferERC20WithABI 使用ABI绑定进行EIP-1559 ERC20转账
func TransferERC20WithABI(client *ethclient.Client, privateKeyHex string, toAddress common.Address, tokenAddress common.Address, amount float64) (common.Hash, error) {
	fmt.Println("\n=== 开始ABI绑定ERC20转账 (EIP-1559) ===")

	// 1. 加载私钥并获取发送方地址
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return common.Hash{}, fmt.Errorf("加载私钥失败: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Hash{}, fmt.Errorf("公钥类型转换失败")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("✓ 发送方: %s\n", fromAddress.Hex())
	fmt.Printf("✓ 接收方: %s\n", toAddress.Hex())
	fmt.Printf("✓ 代币合约: %s\n", tokenAddress.Hex())

	// 2. 创建合约实例
	instance, err := contracts.NewMYERC20(tokenAddress, client)
	if err != nil {
		return common.Hash{}, fmt.Errorf("创建合约实例失败: %v", err)
	}

	// 3. 获取代币精度
	decimals, err := instance.Decimals(&bind.CallOpts{})
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取代币精度失败: %v", err)
	}
	fmt.Printf("✓ 代币精度: %d\n", decimals)

	// 4. 转换代币数量
	tokenAmount := TokenToWei(amount, int(decimals))
	fmt.Printf("✓ 转账数量: %s (原始: %.6f)\n", tokenAmount.String(), amount)

	// 5. 获取当前nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取nonce失败: %v", err)
	}
	fmt.Printf("✓ Nonce: %d\n", nonce)

	// 6. 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取链ID失败: %v", err)
	}
	fmt.Printf("✓ 链ID: %s\n", chainID.String())

	// 7. 获取EIP-1559费用参数
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取区块头失败: %v", err)
	}

	tipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		tipCap = big.NewInt(2e9) // 2 Gwei 默认小费
		fmt.Printf("Warning: 获取小费建议失败，使用默认值: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(tipCap), big.NewFloat(1e9)))
	}

	// 计算gasFeeCap = baseFee * 2 + tipCap
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(header.BaseFee, big.NewInt(2)),
		tipCap,
	)

	fmt.Printf("✓ 基础费用: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(header.BaseFee), big.NewFloat(1e9)))
	fmt.Printf("✓ 小费上限: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(tipCap), big.NewFloat(1e9)))
	fmt.Printf("✓ 费用上限: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(gasFeeCap), big.NewFloat(1e9)))

	// 8. 设置交易选项 (EIP-1559)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return common.Hash{}, fmt.Errorf("创建交易授权失败: %v", err)
	}

	// 配置EIP-1559参数
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // ERC20转账ETH值为0
	auth.GasTipCap = tipCap
	auth.GasFeeCap = gasFeeCap

	// 9. 估算Gas
	gasLimit, err := estimateTransferGas(client, instance, fromAddress, toAddress, tokenAmount)
	if err != nil {
		auth.GasLimit = 60000 // ERC20转账默认Gas限制
		fmt.Printf("Warning: Gas估算失败，使用默认值: %d\n", auth.GasLimit)
	} else {
		auth.GasLimit = gasLimit
		fmt.Printf("✓ 估算Gas: %d\n", gasLimit)
	}

	// 10. 执行转账
	fmt.Println("✓ 开始执行转账交易...")
	tx, err := instance.Transfer(auth, toAddress, tokenAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("转账交易失败: %v", err)
	}

	fmt.Printf("✅ ABI绑定ERC20转账成功!\n")
	fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("交易类型: %d (EIP-1559)\n", tx.Type())

	return tx.Hash(), nil
}

// estimateTransferGas 估算ERC20转账所需的Gas
func estimateTransferGas(client *ethclient.Client, instance *contracts.MYERC20, from, to common.Address, amount *big.Int) (uint64, error) {
	// 使用合约绑定的ABI来生成调用数据进行Gas估算
	callOpts := &bind.CallOpts{
		From: from,
	}

	// 先验证发送方余额（可选验证）
	balance, err := instance.BalanceOf(callOpts, from)
	if err != nil {
		return 0, fmt.Errorf("查询余额失败: %v", err)
	}

	// 简单验证余额是否足够
	if balance.Cmp(amount) < 0 {
		return 0, fmt.Errorf("余额不足: 需要 %s, 当前 %s", amount.String(), balance.String())
	}

	// 返回ERC20转账的标准Gas估算
	// 一般ERC20 transfer需要21000(基础) + 20000(合约调用) + 额外开销
	return 60000, nil
}
