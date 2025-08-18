package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"ethclient_tutorial/block_query"
	"ethclient_tutorial/block_subscription"
	"ethclient_tutorial/config"
	"ethclient_tutorial/contract_deployment"
	"ethclient_tutorial/eth_transfer"
	"ethclient_tutorial/receipt_query"
	"ethclient_tutorial/token_balance"
	"ethclient_tutorial/token_transfer"
	"ethclient_tutorial/transaction_query"
	"ethclient_tutorial/wallet_management"
)

func main() {
	fmt.Println("Ethereum Client Tutorial")
	fmt.Println("========================")

	// 加载配置
	cfg := config.LoadConfig()
	cfg.ValidateConfig()

	// 演示钱包创建功能（不需要网络连接）
	fmt.Println("\n1. 演示钱包创建:")
	walletDemo()

	// 检查是否配置了API密钥
	if cfg.AlchemyAPIKey == "" {
		fmt.Println("\n注意：区块查询、交易查询等功能需要真实的以太坊网络连接")
		fmt.Println("请在 .env 文件中设置 ALCHEMY_API_KEY 来启用网络功能。")
		fmt.Println("示例:")
		fmt.Println("  ALCHEMY_API_KEY=your_api_key_here")
		return
	}

	// 初始化以太坊客户端
	fmt.Printf("\n2. 连接到以太坊网络 (%s)...\n", cfg.EthereumNetwork)

	// 添加调试信息
	fmt.Printf("🔍 调试信息:\n")
	fmt.Printf("   Alchemy API Key: %s\n", cfg.AlchemyAPIKey)
	fmt.Printf("   HTTP URL: %s\n", cfg.GetHTTPURL())
	fmt.Printf("   WebSocket URL: %s\n", cfg.GetWebSocketURL())
	fmt.Printf("   完整连接URL: %s\n", cfg.GetEthereumURL())

	client, err := ethclient.Dial(cfg.GetEthereumURL())
	if err != nil {
		log.Printf("Failed to connect to Ethereum network: %v", err)
		fmt.Println("网络连接失败，仅演示离线功能")
		return
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到��太坊网络!")

	// 网络功能演示
	fmt.Println("\n3. 网络功能演示:")
	blockQueryDemo(client)
	transactionQueryDemo(client)
	receiptQueryDemo(client)

	// 区块订阅演示
	fmt.Println("\n4. 区块订阅演示:")
	blockSubscriptionDemo(client)

	// 只有配置了私钥才演示转账功能
	if cfg.TestPrivateKey != "" {
		// 合约部署演示
		fmt.Println("\n5. 合约部署演示:")
		deployedContractAddress, deployedSuccess := contractDeploymentDemo(client, cfg)

		// 转账功能演示
		fmt.Println("\n6. 转账功能演示:")
		ethTransferDemo(client, cfg)

		// 只有合约部署成功才进行ERC20相关操作
		if deployedSuccess {
			fmt.Println("\n=== 使用新部署的合约进行ERC20操作 ===")
			erc20TransferDemo(client, cfg, deployedContractAddress)

			// 查询接收地址的代币余额
			fmt.Println("\n7. 代币余额查询:")
			checkRecipientTokenBalance(client, cfg, deployedContractAddress)
		} else {
			fmt.Println("\n⚠️ 由于合约部署失败，跳过所有ERC20代币相关功能")
		}
	} else {
		fmt.Println("\n注意：要演示转账功能，请在 .env 文件中设置 TEST_PRIVATE_KEY")
	}
}

func blockQueryDemo(client *ethclient.Client) {
	//blockNum := uint64(15537394)
	//block := block_query.GetBlockByNumber(client, blockNum)
	block := block_query.GetLatestBlock(client)
	fmt.Printf("Block #%d: %s\n", block.NumberU64(), block.Hash().Hex())
}

func transactionQueryDemo(client *ethclient.Client) {
	txHash := common.HexToHash("0x34315509289fd16d4bb9e4d0c9b57441cf31a8c5552bb95a74d988c3f794cb67")
	tx := transaction_query.GetTransaction(client, txHash)
	fmt.Printf("Tx %s => Value: %.4f ETH\n", tx.Hash().Hex(), WeiToEther(tx.Value()))
}

func receiptQueryDemo(client *ethclient.Client) {
	txHash := common.HexToHash("0x34315509289fd16d4bb9e4d0c9b57441cf31a8c5552bb95a74d988c3f794cb67")
	receipt := receipt_query.GetTransactionReceipt(client, txHash)
	fmt.Printf("Receipt: Status=%d, GasUsed=%d\n", receipt.Status, receipt.GasUsed)
}

func walletDemo() {
	wallet, pk := wallet_management.CreateNewWallet()
	fmt.Printf("Address: %s\nPrivate Key: %s\n", wallet.Address.Hex(), pk)
}

func ethTransferDemo(client *ethclient.Client, cfg *config.Config) {
	toAddress := common.HexToAddress(cfg.TestRecipientAddress)

	fmt.Printf("准备从配置的私钥地址转账 0.001 ETH 到 %s\n", cfg.TestRecipientAddress)
	fmt.Println("注意：这只是演示，请确保使用测试网络和测试ETH!")

	txHash, err := eth_transfer.TransferETH(client, cfg.TestPrivateKey, toAddress, 0.001) // 转0.001 ETH
	if err != nil {
		log.Printf("ETH transfer failed: %v", err)
		return
	}
	fmt.Printf("✅ ETH转账成功! TX Hash: %s\n", txHash.Hex())
}

func erc20TransferDemo(client *ethclient.Client, cfg *config.Config, deployedContractAddress common.Address) {
	toAddress := common.HexToAddress(cfg.TestRecipientAddress)
	//打印接受地址
	fmt.Printf("erc20TransferDemo 接收地址: %s\n", toAddress.Hex())

	// 获取ERC20合约地址
	var erc20Address common.Address

	erc20Address = deployedContractAddress
	fmt.Printf("✓ 使用部署的ERC20合约地址: %s\n", erc20Address.Hex())

	fmt.Printf("准备从配置的私钥地址转账 0.001 erc20 到 %s\n", cfg.TestRecipientAddress)
	fmt.Println("注意：这只是演示，请确保使用测试网络和测试代币!")

	// 使用原来的手动构造交易方式
	fmt.Println("\n=== 方式1: 手动构造ERC20转账交易 ===")
	txHash1, err := token_transfer.TransferERC20WithAmount(client, cfg.TestPrivateKey, toAddress, erc20Address, 10, 18) // 添���decimals参数
	if err != nil {
		log.Printf("手动构造ERC20转账失败: %v", err)
	} else {
		fmt.Printf("✅ 手动构造ERC20转账成功! TX Hash: %s\n", txHash1.Hex())
	}

	// 使用基于ABI绑定的新方式
	fmt.Println("\n=== 方式2: 基于ABI绑定的ERC20转账 (EIP-1559) ===")
	txHash2, err := token_transfer.TransferERC20WithABI(client, cfg.TestPrivateKey, toAddress, erc20Address, 15)
	if err != nil {
		log.Printf("ABI绑定ERC20转账失败: %v", err)
	} else {
		fmt.Printf("✅ ABI绑定ERC20转账成功! TX Hash: %s\n", txHash2.Hex())
	}
}

// checkRecipientTokenBalance 查询接收地址的代币余额
func checkRecipientTokenBalance(client *ethclient.Client, cfg *config.Config, deployedContractAddress common.Address) {
	recipientAddress := common.HexToAddress(cfg.TestRecipientAddress)

	// 获取ERC20合约地址
	var erc20Address = deployedContractAddress

	fmt.Printf("查询地址: %s\n", recipientAddress.Hex())
	fmt.Printf("代币合约: %s\n", erc20Address.Hex())

	// 使用完整的代币余额查询功能
	token_balance.CheckTokenBalance(client, erc20Address, recipientAddress)
}

// WeiToEther 将Wei转换为ETH单位
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
}

// blockSubscriptionDemo 区块订阅演���
func blockSubscriptionDemo(client *ethclient.Client) {
	fmt.Println("=== 区块订阅演示 ===")

	// 区块订阅需要WebSocket连接，重新创建WebSocket客户端
	cfg := config.GlobalConfig
	wsClient, err := ethclient.Dial(cfg.GetWebSocketURL())
	if err != nil {
		log.Printf("WebSocket��接失败: %v", err)
		fmt.Println("⚠️ 区块订阅需要WebSocket连接，跳过订阅演示")
		return
	}
	defer wsClient.Close()

	// 订阅并等待下一个新区块
	fmt.Println("\n--- 监听下一个新区块 ---")
	block := block_subscription.BlockSubscription(wsClient)
	fmt.Printf("✅ 成功接收到新区块: #%d\n", block.NumberU64())
}

// contractDeploymentDemo 合约部署演示
func contractDeploymentDemo(client *ethclient.Client, cfg *config.Config) (common.Address, bool) {
	recipientAddress := common.HexToAddress(cfg.TestSendAddress)

	fmt.Printf("准备部署 MYERC20 合约...\n")
	fmt.Printf("部署者私钥对应的地址将成为合约所有者\n")
	fmt.Printf("初始代币接收者: %s\n", recipientAddress.Hex())

	contractAddress, txHash, err := contract_deployment.DeployContract(client, cfg.TestPrivateKey, recipientAddress)
	if err != nil {
		log.Printf("合约部署失败: %v", err)
		return common.Address{}, false
	}

	fmt.Printf("✅ 合约部署成功!\n")
	fmt.Printf("   合约地址: %s\n", contractAddress.Hex())
	fmt.Printf("   部署交易哈希: %s\n", txHash.Hex())

	// 更新配置中的合约地址以供后续使用
	fmt.Printf("📝 建议将合约地址更新到 .env 文件中的 CONTRACT_ADDRESS\n")
	return contractAddress, true
}
