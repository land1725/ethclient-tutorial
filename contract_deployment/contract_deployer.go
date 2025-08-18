package contract_deployment

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
	"ethclient_tutorial/utils"
)

// DeployContract 使用ABI绑定进行EIP-1559 部署合约
func DeployContract(client *ethclient.Client, privateKeyHex string, recipientAddress common.Address) (common.Address, common.Hash, error) {
	fmt.Println("\n=== 开始部署 MYERC20 合约 (EIP-1559) ===")

	// 1. 加载私钥并�����取发送方地址
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("加载私钥失败: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, common.Hash{}, fmt.Errorf("公钥类型转换失败")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("✓ 部署者地址: %s\n", fromAddress.Hex())
	fmt.Printf("✓ 接收者地址: %s\n", recipientAddress.Hex())

	// 2. 获取当前nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("获取nonce失败: %v", err)
	}
	fmt.Printf("✓ Nonce: %d\n", nonce)

	// 3. 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("获取链ID失败: %v", err)
	}
	fmt.Printf("✓ 链ID: %s\n", chainID.String())

	// 4. 获取EIP-1559费用参数
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("获取区块头失败: %v", err)
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

	// 5. 设置交易选项 (EIP-1559)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("创建交易授权失败: %v", err)
	}

	// 配置EIP-1559参数
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0) // 合约部署ETH值为0
	auth.GasTipCap = tipCap
	auth.GasFeeCap = gasFeeCap

	// 6. 基于实际字节码估算部署Gas
	deployGasLimit, err := estimateDeploymentGas(client, fromAddress, recipientAddress)
	if err != nil {
		fmt.Printf("Warning: Gas估算失败，使用默认值: %v\n", err)
		deployGasLimit = 5000000 // 默认500万Gas
	}

	auth.GasLimit = deployGasLimit

	// 7. 部署合约
	fmt.Println("✓ 开始部署合约...")
	contractAddress, tx, instance, err := contracts.DeployMYERC20(auth, client, recipientAddress, fromAddress)
	if err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("合约部署失败: %v", err)
	}

	fmt.Printf("✅ 合约部署交易提��成��!\n")
	fmt.Printf("合约地址: %s\n", contractAddress.Hex())
	fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("交易类型: %d (EIP-1559)\n", tx.Type())

	// 8. 等待交易确认
	fmt.Println("\n--- 等待合约部署确认 ---")
	status, err := utils.WaitForTransactionDeploy(client, tx.Hash())
	if err != nil {
		return common.Address{}, common.Hash{}, fmt.Errorf("等待合约部署确认失败: %v", err)
	}

	if !status.Success {
		return common.Address{}, common.Hash{}, fmt.Errorf("合约部署交易执行失败")
	}

	fmt.Printf("✅ 合约部署已确认!\n")
	fmt.Printf("   区块号: #%d\n", status.BlockNumber)
	fmt.Printf("   Gas使用: %d\n", status.GasUsed)

	// 9. 验证合约部署
	if instance != nil {
		fmt.Println("\n--- 验证合约信息 ---")

		// 获取合约基本信息
		name, err := instance.Name(&bind.CallOpts{})
		if err == nil {
			fmt.Printf("✓ 代币名称: %s\n", name)
		}

		symbol, err := instance.Symbol(&bind.CallOpts{})
		if err == nil {
			fmt.Printf("✓ 代币符号: %s\n", symbol)
		}

		decimals, err := instance.Decimals(&bind.CallOpts{})
		if err == nil {
			fmt.Printf("✓ 代币精度: %d\n", decimals)
		}

		// 验证初始余额
		balance, err := instance.BalanceOf(&bind.CallOpts{}, recipientAddress)
		if err == nil {
			fmt.Printf("✓ 接收者初始余额: %s\n", balance.String())
		}
	}

	return contractAddress, tx.Hash(), nil
}

// estimateDeploymentGas 基于实际字节码估算合约部署所需Gas
func estimateDeploymentGas(client *ethclient.Client, fromAddress, recipientAddress common.Address) (uint64, error) {
	fmt.Println("\n--- Gas估算分析 ---")

	// 1. 读取合约字节码
	contractBin := contracts.MYERC20MetaData.Bin
	if contractBin == "" {
		return 0, fmt.Errorf("无法获取合约字节码")
	}

	// 2. 计算字节码长度
	// 去掉0x前缀（如果有的话）
	if len(contractBin) > 2 && contractBin[:2] == "0x" {
		contractBin = contractBin[2:]
	}

	bytecodeLength := len(contractBin) / 2 // hex字符串转字节长度
	fmt.Printf("✓ 合���字节码长度: %d bytes\n", bytecodeLength)

	// 3. 基于字节码长度计算基础Gas消耗
	// EVM部署合约的Gas计算公式：
	// - 创建操作基础消耗：32,000 Gas
	// - 每字节代码存储：200 Gas (EIP-2028后)
	// - 初始化代码执行：根据复杂度变化

	baseCreationGas := uint64(32000)               // 创建合约的基础Gas
	codeStorageGas := uint64(bytecodeLength) * 200 // 代码存储Gas

	// 4. 根据合约复杂度添加初始化Gas
	// 分析构造函数参数和初始化逻辑复杂度
	initializationGas := estimateInitializationGas(recipientAddress, fromAddress)

	// 5. 计算总Gas需求
	totalEstimatedGas := baseCreationGas + codeStorageGas + initializationGas

	// 6. 添加安全缓冲（30%）
	safetyBuffer := totalEstimatedGas * 30 / 100
	finalGasLimit := totalEstimatedGas + safetyBuffer

	// 7. 确保Gas不低于最小值和不超过区块Gas限制
	const minDeployGas = 1000000  // 最小100万Gas
	const maxDeployGas = 10000000 // 最大1000万Gas

	if finalGasLimit < minDeployGas {
		finalGasLimit = minDeployGas
	}
	if finalGasLimit > maxDeployGas {
		finalGasLimit = maxDeployGas
	}

	// 8. 输出详细分析
	fmt.Printf("✓ 基础创建Gas: %d\n", baseCreationGas)
	fmt.Printf("✓ 代码存储Gas: %d\n", codeStorageGas)
	fmt.Printf("✓ 初始化执行Gas: %d\n", initializationGas)
	fmt.Printf("✓ 估算总Gas: %d\n", totalEstimatedGas)
	fmt.Printf("✓ 安全缓冲: %d (30%%)\n", safetyBuffer)
	fmt.Printf("✓ 最终Gas限制: %d\n", finalGasLimit)

	return finalGasLimit, nil
}

// estimateInitializationGas 估算合约初始化所需的Gas
func estimateInitializationGas(recipientAddress, ownerAddress common.Address) uint64 {
	// 基础初始化Gas：设置owner、名称、符号等
	baseInitGas := uint64(100000)

	// ERC20初始化：铸造初始代币给接收者
	mintGas := uint64(50000) // mint操作大约需要50,000 Gas

	// Pausable和Ownable初始化
	pausableInitGas := uint64(20000)
	ownableInitGas := uint64(20000)

	// EIP712域分隔符设置
	eip712InitGas := uint64(30000)

	// 总初始化Gas
	totalInitGas := baseInitGas + mintGas + pausableInitGas + ownableInitGas + eip712InitGas

	return totalInitGas
}
