#!/bin/bash

# 连接到私有开发链的控制台脚本

GETH_PATH="/Users/temp/go-ethereum/build/bin/geth"
# 连接到项目中的 Geth 实例
PROJECT_DATA_DIR="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/dev-data"

echo "🔗 连接到私有开发链控制台..."
echo "📂 连接数据目录: $PROJECT_DATA_DIR"
echo "💡 使用以下命令可以查看账户信息:"
echo "   eth.accounts"
echo "   eth.getBalance(eth.accounts[0])"
echo "   miner.start()"
echo "   miner.stop()"

# 尝试连接到项目中运行的实例
if [ -S "$PROJECT_DATA_DIR/geth.ipc" ]; then
    echo "✅ 找到运行中的 Geth 实例，正在连接..."
    $GETH_PATH attach --datadir "$PROJECT_DATA_DIR"
else
    echo "❌ 未找到运行中的 Geth 实例"
    echo "IPC 文件路径: $PROJECT_DATA_DIR/geth.ipc"
    echo "请先运行以下命令启动私有开发链："
    echo "  ./start-dev-simple.sh 或 ./start-dev-chain.sh"
fi
