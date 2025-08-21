#!/bin/bash

# 连接到私有开发链的控制台脚本

GETH_PATH="/Users/temp/go-ethereum/build/bin/geth"
DATA_DIR="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/data"

echo "🔗 连接到私有开发链控制台..."
echo "💡 使用以下命令可以查看账户信息:"
echo "   eth.accounts"
echo "   eth.getBalance(eth.accounts[0])"
echo "   miner.start()"
echo "   miner.stop()"

$GETH_PATH attach --datadir "$DATA_DIR"
