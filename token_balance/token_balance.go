package token_balance

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"ethclient_tutorial/contracts"
)

// CheckTokenBalance 使用ABI绑定查询代币余额和基本信息
func CheckTokenBalance(client *ethclient.Client, tokenAddress common.Address, holderAddress common.Address) {
	fmt.Printf("\n=== 代币信息查询 (使用ABI绑定) ===\n")
	fmt.Printf("代币合约地址: %s\n", tokenAddress.Hex())
	fmt.Printf("持有者地址: %s\n", holderAddress.Hex())

	// 创建合约实例
	instance, err := contracts.NewMYERC20(tokenAddress, client)
	if err != nil {
		fmt.Printf("❌ 创建合约实例失败: %v\n", err)
		return
	}
	fmt.Printf("✓ 合约实例创建成功\n")

	// 设置调用选项
	callOpts := &bind.CallOpts{}

	// 1. 查询代币名称
	name, err := instance.Name(callOpts)
	if err != nil {
		fmt.Printf("❌ 查询代币名称失败: %v\n", err)
	} else {
		fmt.Printf("✓ 代币名称: %s\n", name)
	}

	// 2. 查询代币符号
	symbol, err := instance.Symbol(callOpts)
	if err != nil {
		fmt.Printf("❌ 查询代币符号失败: %v\n", err)
	} else {
		fmt.Printf("✓ 代币符号: %s\n", symbol)
	}

	// 3. 查询代币精度
	decimals, err := instance.Decimals(callOpts)
	if err != nil {
		fmt.Printf("❌ 查询代币精度失败: %v\n", err)
	} else {
		fmt.Printf("✓ 代币精度: %d\n", decimals)
	}

	// 4. 查询总供应量
	totalSupply, err := instance.TotalSupply(callOpts)
	if err != nil {
		fmt.Printf("❌ 查询总供应量失败: %v\n", err)
	} else {
		fmt.Printf("✓ 总供应量: %s (原始值)\n", totalSupply.String())
		if decimals > 0 {
			readable := TokenFromWei(totalSupply, int(decimals))
			fmt.Printf("✓ 总供应量: %s %s\n", readable, symbol)
		}
	}

	// 5. 查询指定地址的余额
	balance, err := instance.BalanceOf(callOpts, holderAddress)
	if err != nil {
		fmt.Printf("❌ 查询余额失败: %v\n", err)
	} else {
		fmt.Printf("✓ 原始余额: %s\n", balance.String())
		if decimals > 0 {
			readable := TokenFromWei(balance, int(decimals))
			fmt.Printf("✓ 可读余额: %s %s\n", readable, symbol)
		}
	}

	// 6. 查询合约所有者（如果合约支持）
	owner, err := instance.Owner(callOpts)
	if err != nil {
		fmt.Printf("⚠️  查询合约所有者失败: %v\n", err)
	} else {
		fmt.Printf("✓ 合约所有者: %s\n", owner.Hex())
	}

	// 7. 查询暂停状态（如果合约支持）
	paused, err := instance.Paused(callOpts)
	if err != nil {
		fmt.Printf("⚠️  查询暂停状态失败: %v\n", err)
	} else {
		if paused {
			fmt.Printf("⚠️  合约状态: 已暂停\n")
		} else {
			fmt.Printf("✓ 合约状态: 正常运行\n")
		}
	}
}

// GetTokenInfo 获取代币基本信息
func GetTokenInfo(client *ethclient.Client, tokenAddress common.Address) (*TokenInfo, error) {
	// 创建合约实例
	instance, err := contracts.NewMYERC20(tokenAddress, client)
	if err != nil {
		return nil, fmt.Errorf("创建合约实例失败: %v", err)
	}

	callOpts := &bind.CallOpts{}

	// 获取代币基本信息
	name, err := instance.Name(callOpts)
	if err != nil {
		return nil, fmt.Errorf("查询代币名称失败: %v", err)
	}

	symbol, err := instance.Symbol(callOpts)
	if err != nil {
		return nil, fmt.Errorf("查询代币符号失败: %v", err)
	}

	decimals, err := instance.Decimals(callOpts)
	if err != nil {
		return nil, fmt.Errorf("查询代币精度失败: %v", err)
	}

	totalSupply, err := instance.TotalSupply(callOpts)
	if err != nil {
		return nil, fmt.Errorf("查询总供应量失败: %v", err)
	}

	return &TokenInfo{
		Address:     tokenAddress,
		Name:        name,
		Symbol:      symbol,
		Decimals:    decimals,
		TotalSupply: totalSupply,
	}, nil
}

// GetTokenBalance 获取指定地址的代币余额
func GetTokenBalance(client *ethclient.Client, tokenAddress common.Address, holderAddress common.Address) (*big.Int, error) {
	// 创建合约实例
	instance, err := contracts.NewMYERC20(tokenAddress, client)
	if err != nil {
		return nil, fmt.Errorf("创建合约实例失败: %v", err)
	}

	callOpts := &bind.CallOpts{}

	// 查询余额
	balance, err := instance.BalanceOf(callOpts, holderAddress)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %v", err)
	}

	return balance, nil
}

// TokenInfo 代币信息结构体
type TokenInfo struct {
	Address     common.Address
	Name        string
	Symbol      string
	Decimals    uint8
	TotalSupply *big.Int
}

// TokenFromWei 将wei单位转换为代币单位（考虑精度）
func TokenFromWei(wei *big.Int, decimals int) string {
	if decimals == 0 {
		return wei.String()
	}

	// 创建除数 (10^decimals)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// 转换为浮点数进行除法运算
	weiFloat := new(big.Float).SetInt(wei)
	divisorFloat := new(big.Float).SetInt(divisor)
	result := new(big.Float).Quo(weiFloat, divisorFloat)

	// 格式化输出，保留6位小数
	return result.Text('f', 6)
}

// TokenToWei 将代币单位转换为wei单位（考虑精度）
func TokenToWei(amount float64, decimals int) *big.Int {
	if decimals == 0 {
		return big.NewInt(int64(amount))
	}

	// 创建乘数 (10^decimals)
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// 将浮点数转换为big.Float
	amountFloat := big.NewFloat(amount)
	multiplierFloat := new(big.Float).SetInt(multiplier)

	// 执行乘法
	result := new(big.Float).Mul(amountFloat, multiplierFloat)

	// 转换为big.Int
	wei, _ := result.Int(nil)
	return wei
}
