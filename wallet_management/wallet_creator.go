package wallet_management

import (
	"crypto/ecdsa"
	"encoding/hex"
	"log"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// CreateNewWallet 创建新钱包并返回地址和私钥
func CreateNewWallet() (accounts.Account, string) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal("Failed to generate private key: ", err)
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 创建账户结构
	account := accounts.Account{
		Address: address,
	}

	return account, privateKeyHex
}

// ValidateAddress 验证以太坊地址有效性
func ValidateAddress(address string) bool {
	return common.IsHexAddress(address)
}
