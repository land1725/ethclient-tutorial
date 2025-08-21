#!/bin/bash

# 快速开发模式启动脚本
# 使用 --dev 模式，自动挖矿，预分配账户

set -e

GETH_PATH="/Users/temp/go-ethereum/build/bin/geth"
DATA_DIR="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/dev-data"

echo "🚀 启动快速开发模式..."
echo "🔗 HTTP RPC: http://localhost:8545"
echo "📡 WebSocket: ws://localhost:8546"
echo "💰 预分配账户将自动创建并资助"

# 创建数据目录
mkdir -p "$DATA_DIR"

# 启动开发模式
$GETH_PATH \
    --datadir "$DATA_DIR" \
    --dev \
    --dev.period 2 \
    --http \
    --http.addr "0.0.0.0" \
    --http.port 8545 \
    --http.api "eth,net,web3,personal,admin,miner,debug,txpool" \
    --http.corsdomain "*" \
    --ws \
    --ws.addr "0.0.0.0" \
    --ws.port 8546 \
    --ws.api "eth,net,web3,personal,admin,miner,debug,txpool" \
    --ws.origins "*" \
    --allow-insecure-unlock \
    --verbosity 3
