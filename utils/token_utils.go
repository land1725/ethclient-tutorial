package utils

import (
	"math/big"
)

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

// WeiToToken 将wei单位转换为代币单位（考虑精度）
func WeiToToken(weiAmount *big.Int, decimals int) *big.Float {
	if decimals == 0 {
		return new(big.Float).SetInt(weiAmount)
	}

	// 创建除数 (10^decimals)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	divisorFloat := new(big.Float).SetInt(divisor)

	// 将wei转换为big.Float并除以除数
	weiFloat := new(big.Float).SetInt(weiAmount)
	result := new(big.Float).Quo(weiFloat, divisorFloat)

	return result
}

// FormatTokenAmount 格式化代币数量显示
func FormatTokenAmount(weiAmount *big.Int, decimals int, precision int) string {
	tokenAmount := WeiToToken(weiAmount, decimals)
	return tokenAmount.Text('f', precision)
}
