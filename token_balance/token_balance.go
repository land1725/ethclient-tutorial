package token_balance

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC20 ABI for balanceOf function
const erc20ABI = `[
	{
		"constant": true,
		"inputs": [
			{
				"name": "_owner",
				"type": "address"
			}
		],
		"name": "balanceOf",
		"outputs": [
			{
				"name": "balance",
				"type": "uint256"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "decimals",
		"outputs": [
			{
				"name": "",
				"type": "uint8"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "symbol",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	},
	{
		"constant": true,
		"inputs": [],
		"name": "name",
		"outputs": [
			{
				"name": "",
				"type": "string"
			}
		],
		"payable": false,
		"stateMutability": "view",
		"type": "function"
	}
]`

// GetTokenBalance 查询ERC20代币余额
func GetTokenBalance(client *ethclient.Client, tokenContract, walletAddress common.Address) (*big.Int, error) {
	//// 方法1: 使用ABI调用
	//balance, err := getTokenBalanceWithABI(client, tokenContract, walletAddress)
	//if err != nil {
	//	// 如果ABI方法失败，尝试手动构造调用
	//	fmt.Printf("ABI方法失败，尝试手动构造: %v\n", err)
	//	return getTokenBalanceManual(client, tokenContract, walletAddress)
	//}
	//return balance, nil
	return getTokenBalanceManual(client, tokenContract, walletAddress)

}

// getTokenBalanceWithABI 使用ABI查询代币余额
func getTokenBalanceWithABI(client *ethclient.Client, tokenContract, walletAddress common.Address) (*big.Int, error) {
	// 解析ABI
	contractABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// 构造balanceOf调用数据
	data, err := contractABI.Pack("balanceOf", walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to pack balanceOf data: %v", err)
	}

	// 调用合约
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenContract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	// 解包结果
	var balance *big.Int
	err = contractABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %v", err)
	}

	return balance, nil
}

// getTokenBalanceManual 手动构造balanceOf调用
func getTokenBalanceManual(client *ethclient.Client, tokenContract, walletAddress common.Address) (*big.Int, error) {
	// balanceOf(address) 函数签名
	balanceOfSignature := []byte("balanceOf(address)")
	balanceOfHash := crypto.Keccak256Hash(balanceOfSignature)

	// 构造调用数据: 4字节函数选择器 + 32字节地址参数
	data := append(balanceOfHash[:4], common.LeftPadBytes(walletAddress.Bytes(), 32)...)

	// 调用合约
	result, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenContract,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract manually: %v", err)
	}

	// 检查返回数据长度
	if len(result) != 32 {
		return nil, fmt.Errorf("unexpected result length: got %d, expected 32", len(result))
	}

	// 将结果转换为big.Int
	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// GetTokenInfo 获取代币基本信息
func GetTokenInfo(client *ethclient.Client, tokenContract common.Address) (name, symbol string, decimals uint8, err error) {
	contractABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// 获取代币名称
	nameData, err := contractABI.Pack("name")
	if err == nil {
		result, err := client.CallContract(context.Background(), ethereum.CallMsg{
			To:   &tokenContract,
			Data: nameData,
		}, nil)
		if err == nil {
			var tokenName string
			contractABI.UnpackIntoInterface(&tokenName, "name", result)
			name = tokenName
		}
	}

	// 获取代币符号
	symbolData, err := contractABI.Pack("symbol")
	if err == nil {
		result, err := client.CallContract(context.Background(), ethereum.CallMsg{
			To:   &tokenContract,
			Data: symbolData,
		}, nil)
		if err == nil {
			var tokenSymbol string
			contractABI.UnpackIntoInterface(&tokenSymbol, "symbol", result)
			symbol = tokenSymbol
		}
	}

	// 获取小数位数
	decimalsData, err := contractABI.Pack("decimals")
	if err == nil {
		result, err := client.CallContract(context.Background(), ethereum.CallMsg{
			To:   &tokenContract,
			Data: decimalsData,
		}, nil)
		if err == nil {
			var tokenDecimals uint8
			contractABI.UnpackIntoInterface(&tokenDecimals, "decimals", result)
			decimals = tokenDecimals
		}
	}

	return name, symbol, decimals, nil
}

// FormatTokenBalance 格式化代币余额显示
func FormatTokenBalance(balance *big.Int, decimals uint8) *big.Float {
	if balance == nil {
		return big.NewFloat(0)
	}

	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	balanceFloat := new(big.Float).SetInt(balance)
	divisorFloat := new(big.Float).SetInt(divisor)

	return new(big.Float).Quo(balanceFloat, divisorFloat)
}

// CheckTokenBalance 完整的代币余额查询函数
func CheckTokenBalance(client *ethclient.Client, tokenContract, walletAddress common.Address) {
	fmt.Printf("\n=== 查询代币余额 ===\n")
	fmt.Printf("代币合约地址: %s\n", tokenContract.Hex())
	fmt.Printf("钱包地址: %s\n", walletAddress.Hex())

	// 获取代币信息
	name, symbol, decimals, err := GetTokenInfo(client, tokenContract)
	if err != nil {
		fmt.Printf("Warning: 无法获取代币信息: %v\n", err)
		decimals = 18 // 使用默认值
		symbol = "TOKEN"
	} else {
		fmt.Printf("代币名称: %s (%s)\n", name, symbol)
		fmt.Printf("小数位数: %d\n", decimals)
	}

	// 查询余额
	balance, err := GetTokenBalance(client, tokenContract, walletAddress)
	if err != nil {
		fmt.Printf("❌ 查询余额失败: %v\n", err)
		return
	}

	// 格式化并显示余额
	formattedBalance := FormatTokenBalance(balance, decimals)
	fmt.Printf("✓ 原始余额: %s wei\n", balance.String())
	fmt.Printf("✓ 格式化余额: %s %s\n", formattedBalance.Text('f', 6), symbol)

	// 检查余额是否为0
	if balance.Cmp(big.NewInt(0)) == 0 {
		fmt.Printf("⚠️  余额为0，可能的原因：\n")
		fmt.Printf("   1. 地址确实没有该代币\n")
		fmt.Printf("   2. 代币合约地址错误\n")
		fmt.Printf("   3. 网络连接问题\n")
		fmt.Printf("   4. 合约不是标准ERC20代币\n")
	}
}
