package token_transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"ethclient_tutorial/utils"
)

// TransferERC20WithABIFile 使用ABI文件进行EIP-1559 ERC20转账
func TransferERC20WithABIFile(client *ethclient.Client, privateKeyHex string, toAddress common.Address, tokenAddress common.Address, amount float64) (common.Hash, error) {
	fmt.Println("\n=== 开始ABI文件ERC20转账 (EIP-1559) ===")

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

	// 2. 读取并解析ABI文件
	abiData, err := os.ReadFile("contracts/compiled/MyToken.abi")
	if err != nil {
		return common.Hash{}, fmt.Errorf("读取ABI文件失败: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiData)))
	if err != nil {
		return common.Hash{}, fmt.Errorf("解析ABI失败: %v", err)
	}
	fmt.Printf("✓ ABI文件解析成功\n")

	// 3. 获取代币精度
	decimals, err := getTokenDecimals(client, tokenAddress, parsedABI)
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取代币精度失败: %v", err)
	}
	fmt.Printf("✓ 代币精度: %d\n", decimals)

	// 4. 转换代币数量
	tokenAmount := utils.TokenToWei(amount, int(decimals))
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

	// 8. 构造transfer函数调用数据
	callData, err := parsedABI.Pack("transfer", toAddress, tokenAmount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("构造调用数据失败: %v", err)
	}
	fmt.Printf("✓ 调用数据构造成功，长度: %d bytes\n", len(callData))

	// 9. 估算Gas
	gasLimit, err := estimateGasForABITransfer(client, fromAddress, tokenAddress, callData)
	if err != nil {
		gasLimit = 60000 // ERC20转账默认Gas限制
		fmt.Printf("Warning: Gas估算失败，使用默认值: %d\n", gasLimit)
	} else {
		fmt.Printf("✓ 估算Gas: %d\n", gasLimit)
	}

	// 10. 创建EIP-1559交易
	tx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &tokenAddress,
		Value:     big.NewInt(0), // ERC20转账ETH值为0
		Data:      callData,
	}

	// 包装为通用交易类型
	newTx := types.NewTx(tx)

	// 11. 签名交易
	fmt.Println("✓ 开始签名交易...")
	signedTx, err := types.SignTx(newTx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return common.Hash{}, fmt.Errorf("交易签名失败: %v", err)
	}
	fmt.Printf("✓ 交易签名成功\n")

	// 12. 发送交易
	fmt.Println("✓ 开始发送交易...")
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("发送交易失败: %v", err)
	}

	fmt.Printf("✅ ERC20转账交易提交成功!\n")
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())
	fmt.Printf("交易类型: %d (EIP-1559)\n", signedTx.Type())

	// 13. 等待交易确认
	fmt.Println("\n--- 等待转账确认 ---")
	status, err := utils.WaitForTransactionQuick(client, signedTx.Hash())
	if err != nil {
		return common.Hash{}, fmt.Errorf("等待转账确认失败: %v", err)
	}

	if !status.Success {
		// 添加详细的失败诊断
		fmt.Printf("❌ 转账交易执行失败!\n")
		fmt.Printf("   交易哈希: %s\n", signedTx.Hash().Hex())
		fmt.Printf("   区块号: #%d\n", status.BlockNumber)
		fmt.Printf("   Gas使用: %d\n", status.GasUsed)

		// 检查可能的失败原因
		fmt.Println("\n--- 失败原因诊断 ---")

		// 1. 检查发送方余额
		senderBalance, err := getTokenBalance(client, tokenAddress, fromAddress, parsedABI)
		if err == nil {
			fmt.Printf("发送方代币余额: %s\n", senderBalance.String())
			if senderBalance.Cmp(tokenAmount) < 0 {
				fmt.Printf("❌ 余额不足! 需要: %s, 当前: %s\n", tokenAmount.String(), senderBalance.String())
			} else {
				fmt.Printf("✓ 余额充足: %s >= %s\n", senderBalance.String(), tokenAmount.String())
			}
		}

		// 2. 检查合约状态
		paused, err := getTokenPaused(client, tokenAddress, parsedABI)
		if err == nil {
			if paused {
				fmt.Printf("❌ 合约已暂停!\n")
			} else {
				fmt.Printf("✓ 合约未暂停\n")
			}
		}

		return common.Hash{}, fmt.Errorf("转账交易执行失败 - 请查看上述诊断信息")
	}

	fmt.Printf("✅ 转账已确认!\n")
	fmt.Printf("   区块号: #%d\n", status.BlockNumber)
	fmt.Printf("   Gas使用: %d\n", status.GasUsed)

	// 14. 验证转账结果
	fmt.Println("\n--- 验证转账结果 ---")

	// 查询转账后余额
	senderBalance, err := getTokenBalance(client, tokenAddress, fromAddress, parsedABI)
	if err == nil {
		fmt.Printf("✓ 发送方余额: %s\n", senderBalance.String())
	}

	receiverBalance, err := getTokenBalance(client, tokenAddress, toAddress, parsedABI)
	if err == nil {
		fmt.Printf("✓ 接收方余额: %s\n", receiverBalance.String())
	}

	return signedTx.Hash(), nil
}

// getTokenDecimals 获取代币精度
func getTokenDecimals(client *ethclient.Client, tokenAddress common.Address, parsedABI abi.ABI) (uint8, error) {
	callData, err := parsedABI.Pack("decimals")
	if err != nil {
		return 0, fmt.Errorf("构造decimals调用数据失败: %v", err)
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return 0, fmt.Errorf("调用decimals失败: %v", err)
	}

	var decimals uint8
	err = parsedABI.UnpackIntoInterface(&decimals, "decimals", result)
	if err != nil {
		return 0, fmt.Errorf("解析decimals结果失败: %v", err)
	}

	return decimals, nil
}

// getTokenBalance 获取代币余额
func getTokenBalance(client *ethclient.Client, tokenAddress common.Address, account common.Address, parsedABI abi.ABI) (*big.Int, error) {
	callData, err := parsedABI.Pack("balanceOf", account)
	if err != nil {
		return nil, fmt.Errorf("构造balanceOf调用数据失败: %v", err)
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("调用balanceOf失败: %v", err)
	}

	var balance *big.Int
	err = parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("解析balanceOf结果失败: %v", err)
	}

	return balance, nil
}

// getTokenPaused 检查合约是否暂停
func getTokenPaused(client *ethclient.Client, tokenAddress common.Address, parsedABI abi.ABI) (bool, error) {
	callData, err := parsedABI.Pack("paused")
	if err != nil {
		return false, fmt.Errorf("构造paused调用数据失败: %v", err)
	}

	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: callData,
	}, nil)
	if err != nil {
		return false, fmt.Errorf("调用paused失败: %v", err)
	}

	var paused bool
	err = parsedABI.UnpackIntoInterface(&paused, "paused", result)
	if err != nil {
		return false, fmt.Errorf("解析paused结果失败: %v", err)
	}

	return paused, nil
}

// estimateGasForABITransfer 估算ABI转账所需的Gas
func estimateGasForABITransfer(client *ethclient.Client, from, to common.Address, data []byte) (uint64, error) {
	// 使用CallMsg估算Gas
	msg := ethereum.CallMsg{
		From: from,
		To:   &to,
		Data: data,
	}

	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, fmt.Errorf("Gas估算失败: %v", err)
	}

	// 为估算的Gas添加20%的缓冲
	gasWithBuffer := gasLimit + (gasLimit * 20 / 100)
	return gasWithBuffer, nil
}
