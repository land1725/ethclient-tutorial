#!/bin/bash

# 私有开发链启动脚本（持久化模式，非开发者模式）
# 使用最新编译的 Geth 创建持久化的私有链

set -e

# 配置变量
GETH_PATH="/Users/temp/go-ethereum/build/bin/geth"
DATA_DIR="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/data"
GENESIS_FILE="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/genesis.json"
HTTP_PORT=8545
WS_PORT=8546
P2P_PORT=30303
NETWORK_ID=1337  # 私有链网络ID（自定义）
ACCOUNT="0xBc694bc8E249956958dBc2529d39bBc94647712F"  # 你的账户地址

# 创建数据目录
mkdir -p "$DATA_DIR"

echo "🚀 启动私有以太坊链（持久化模式）..."
echo "📂 数据目录: $DATA_DIR"
echo "🔗 HTTP RPC: http://localhost:$HTTP_PORT"
echo "📡 WebSocket: ws://localhost:$WS_PORT"
echo "🌐 网络ID: $NETWORK_ID"

# 检查是否需要初始化创世区块
if [ ! -d "$DATA_DIR/geth" ]; then
    echo "🔧 初始化创世区块..."
    $GETH_PATH --datadir "$DATA_DIR" init "$GENESIS_FILE"
fi

# 启动 Geth 节点（非开发者模式）
echo "⚡ 启动 Geth 节点..."
$GETH_PATH \
    --datadir "$DATA_DIR" \
    --networkid $NETWORK_ID \
    --http \
    --http.addr "0.0.0.0" \
    --http.port $HTTP_PORT \
    --http.api "admin,debug,eth,miner,net,personal,txpool,web3" \
    --http.corsdomain "*" \
    --ws \
    --ws.addr "127.0.0.1" \
    --ws.port $WS_PORT \
    --ws.api "admin,debug,eth,miner,net,personal,txpool,web3" \
    --ws.origins "*" \
    --port $P2P_PORT \
    --unlock "$ACCOUNT" \
    --password /dev/null \
    --allow-insecure-unlock \
    --mine \
    --miner.etherbase "$ACCOUNT" \
    --gcmode archive \
    --nodiscover \
    --maxpeers 0 \
    --verbosity 3 \
    --log.file "$DATA_DIR/geth.log"
