package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"ethclient_tutorial/block_query"
	"ethclient_tutorial/config"
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
	if cfg.InfuraProjectID == "" {
		fmt.Println("\n注意：区块查询、交易查询等功能需要真实的以太坊网络连接")
		fmt.Println("请在 .env 文件中设置 INFURA_PROJECT_ID 来启用网络功能。")
		fmt.Println("示例:")
		fmt.Println("  INFURA_PROJECT_ID=your_project_id_here")
		return
	}

	// 初始化以太坊客户端
	fmt.Printf("\n2. 连接到以太坊网络 (%s)...\n", cfg.EthereumNetwork)
	client, err := ethclient.Dial(cfg.GetEthereumURL())
	if err != nil {
		log.Printf("Failed to connect to Ethereum network: %v", err)
		fmt.Println("网络连接失败，仅演示离线功能")
		return
	}
	defer client.Close()

	fmt.Println("✅ 成功连接到以太坊网络!")

	// 网络功能演示
	fmt.Println("\n3. 网络功能演示:")
	blockQueryDemo(client)
	transactionQueryDemo(client)
	receiptQueryDemo(client)

	// 只有配置了私钥才演示转账功能
	if cfg.TestPrivateKey != "" {
		//fmt.Println("\n4. 转账功能演示:")
		ethTransferDemo(client, cfg)
		erc20TransferDemo(client, cfg)

		// 查询接收地址的代币余额
		fmt.Println("\n5. 代币余额查询:")
		checkRecipientTokenBalance(client, cfg)
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

func erc20TransferDemo(client *ethclient.Client, cfg *config.Config) {
	toAddress := common.HexToAddress(cfg.TestRecipientAddress)
	//打印接受地址
	fmt.Printf("erc20TransferDemo 接收地址: %s\n", toAddress.Hex())

	// 获取ERC20合约地址
	var erc20Address common.Address

	erc20Address = common.HexToAddress(cfg.ContractAddress)
	fmt.Printf("✓ 使用配置的ERC20合约地址: %s\n", erc20Address.Hex())

	fmt.Printf("准备从配置的私钥地址转账 0.001 erc20 到 %s\n", cfg.TestRecipientAddress)
	fmt.Println("注意：这只是演示，请确保使用测试网络和测试代币!")

	// 使用原来的手动构造交易方式
	fmt.Println("\n=== 方式1: 手动构造ERC20转账交易 ===")
	txHash1, err := token_transfer.TransferERC20WithAmount(client, cfg.TestPrivateKey, toAddress, erc20Address, 10, 18) // 添加decimals参数
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
func checkRecipientTokenBalance(client *ethclient.Client, cfg *config.Config) {
	recipientAddress := common.HexToAddress(cfg.TestRecipientAddress)

	// 获取ERC20合约地址
	var erc20Address common.Address
	if cfg.ContractAddress == "" {
		// 使用交易中实际使用的GLD代币合约地址
		erc20Address = common.HexToAddress("0x38a62fbf3373325D2FCE9692749ae6Bc35ac31A2") // GLD合约地址
		fmt.Printf("使用默认GLD合约地址: %s\n", erc20Address.Hex())
	} else {
		erc20Address = common.HexToAddress(cfg.ContractAddress)
	}

	fmt.Printf("查询地址: %s\n", recipientAddress.Hex())
	fmt.Printf("代币合约: %s\n", erc20Address.Hex())

	// 使用完整的代币余额查询功能
	token_balance.CheckTokenBalance(client, erc20Address, recipientAddress)
}

// WeiToEther 将Wei转换为ETH单位
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
}
