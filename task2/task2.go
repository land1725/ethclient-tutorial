package task2

import (
	"ethclient_tutorial/contracts"
	"ethclient_tutorial/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetTotalSupply 查询总供应量
// 该函数用于查询指定地址的总供应量，使用EIP-1559
func GetTotalSupply(client *ethclient.Client, tokenAddress common.Address) (*big.Float, error) {
	//根据tokenAddress获取合约实例
	instance, err := contracts.NewMYERC20(tokenAddress, client)
	if err != nil {
		return big.NewFloat(0.0), fmt.Errorf("创建合约实例失败: %v", err)
	}
	//调用合约的TotalSupply方法获取总供应量
	totalSupply, err := instance.TotalSupply(&bind.CallOpts{})
	if err != nil {
		return big.NewFloat(0.0), fmt.Errorf("获取总供应量失败: %v", err)
	}
	fmt.Printf("✓ 总供应量: %s\n", totalSupply.String())
	//将总供应量转换为Wei格式
	totalSupplyWei := utils.WeiToToken(totalSupply, 18) // 假设代币精度为18
	fmt.Printf("✓ 总供应量 (Wei): %s\n", totalSupplyWei.String())
	//返回总供应量
	return totalSupplyWei, nil
}
