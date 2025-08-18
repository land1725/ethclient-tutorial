package token_transfer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ModernERC20Transfer 现代化的ERC20转账 - 手动构造哈希，使用EIP-1559
func ERC20Transfer(client *ethclient.Client, privateKeyHex string, toAddress common.Address, tokenAddress common.Address, amount *big.Int) (common.Hash, error) {
	fmt.Println("\n=== 开始现代化ERC20转账 (手动哈希 + EIP-1559) ===")

	// 1. 加载私钥并获取发送方地址
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("✓ 发送方: %s\n", fromAddress.Hex())
	fmt.Printf("✓ 接收方: %s\n", toAddress.Hex())
	fmt.Printf("✓ 代币合约: %s\n", tokenAddress.Hex())

	// 2. 获取当前nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Nonce: %d\n", nonce)

	// 3. 手动构造transfer(address,uint256)函数调用数据
	transferFnSignature := []byte("transfer(address,uint256)")
	methodID := crypto.Keccak256(transferFnSignature)[:4]
	fmt.Printf("✓ 方法ID: %s\n", hexutil.Encode(methodID))

	// 4. 构造函数参数
	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	// 组装完整的调用数据
	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	fmt.Printf("✓ 调用数据: %s\n", hexutil.Encode(data))

	// 5. 估算Gas
	gasLimit, err := client.EstimateGas(context.Background(), ethereum.CallMsg{
		From: fromAddress,
		To:   &tokenAddress,
		Data: data,
	})
	if err != nil {
		gasLimit = 60000 // ERC20转账默认Gas
		fmt.Printf("Warning: Gas估算失败，使用默认: %d\n", gasLimit)
	} else {
		fmt.Printf("✓ 估算Gas: %d\n", gasLimit)
	}

	// 6. 获取链ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ 链ID: %s\n", chainID.String())

	// 7. 获取EIP-1559费用参数
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	tipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		tipCap = big.NewInt(2e9) // 2 Gwei
	}

	// 计算gasFeeCap = baseFee * 2 + tipCap
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(header.BaseFee, big.NewInt(2)),
		tipCap,
	)

	fmt.Printf("✓ 基础费用: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(header.BaseFee), big.NewFloat(1e9)))
	fmt.Printf("✓ 小费上限: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(tipCap), big.NewFloat(1e9)))
	fmt.Printf("✓ 费用上限: %s Gwei\n", new(big.Float).Quo(new(big.Float).SetInt(gasFeeCap), big.NewFloat(1e9)))

	// 8. 创建EIP-1559交易
	tx := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &tokenAddress,
		Value:     big.NewInt(0), // ERC20转账ETH值为0
		Data:      data,
	}

	// 9. 签名交易
	signedTx, err := types.SignTx(types.NewTx(tx), types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ 交易已签名: %s\n", signedTx.Hash().Hex())

	// 10. 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✅ 交易发送成功: %s\n", signedTx.Hash().Hex())
	return signedTx.Hash(), nil
}

// TokenToWei 将代币数量转换为最小单位
func TokenToWei(amount float64, decimals int) *big.Int {
	factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	amountFloat := new(big.Float).Mul(big.NewFloat(amount), new(big.Float).SetInt(factor))
	weiInt := new(big.Int)
	amountFloat.Int(weiInt)
	return weiInt
}

// TransferERC20WithAmount 使用浮点数金额的便捷函数
func TransferERC20WithAmount(client *ethclient.Client, privateKeyHex string, toAddress common.Address, tokenAddress common.Address, amount float64, decimals int) (common.Hash, error) {
	tokenAmount := TokenToWei(amount, decimals)
	return ERC20Transfer(client, privateKeyHex, toAddress, tokenAddress, tokenAmount)
}
