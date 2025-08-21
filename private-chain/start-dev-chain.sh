#!/bin/bash

# 私有开发链启动脚本
# 使用最新编译的 Geth 创建持久化的私有开发链

set -e

# 配置变量
GETH_PATH="/Users/temp/go-ethereum/build/bin/geth"
DATA_DIR="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/data"
GENESIS_FILE="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/genesis.json"
HTTP_PORT=8545
WS_PORT=8546
P2P_PORT=30303

# 创建数据目录
mkdir -p "$DATA_DIR"

echo "🚀 启动私有以太坊开发链..."
echo "📂 数据目录: $DATA_DIR"
echo "🔗 HTTP RPC: http://localhost:$HTTP_PORT"
echo "📡 WebSocket: ws://localhost:$WS_PORT"

# 检查是否需要初始化创世区块
if [ ! -d "$DATA_DIR/geth" ]; then
    echo "🔧 初始化创世区块..."
    $GETH_PATH --datadir "$DATA_DIR" init "$GENESIS_FILE"
fi

# 启动 Geth 节点
echo "⚡ 启动 Geth 节点..."
$GETH_PATH \
    --datadir "$DATA_DIR" \
    --http \
    --http.addr "0.0.0.0" \
    --http.port $HTTP_PORT \
    --http.api "eth,net,web3,personal,admin,miner,debug,txpool" \
    --http.corsdomain "*" \
    --ws \
    --ws.addr "0.0.0.0" \
    --ws.port $WS_PORT \
    --ws.api "eth,net,web3,personal,admin,miner,debug,txpool" \
    --ws.origins "*" \
    --port $P2P_PORT \
    --miner.etherbase "0xBc694bc8E249956958dBc2529d39bBc94647712F" \
    --unlock "0xBc694bc8E249956958dBc2529d39bBc94647712F" \
    --password /dev/null \
    --allow-insecure-unlock \
    --dev \
    --dev.period 5 \
    --gcmode archive \
    --verbosity 3 \
    --log.file "$DATA_DIR/geth.log"
