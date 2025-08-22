#!/bin/bash

# 连接到私有开发链的控制台脚本

GETH_PATH="/Users/temp/go-ethereum/build/bin/geth"
# 连接到项目中的 Geth 实例 - 修复数据目录路径
PROJECT_DATA_DIR="/Users/temp/GolandProjects/ethclient_tutorial/private-chain/data"
IPC_PATH="$PROJECT_DATA_DIR/geth.ipc"

echo "🔗 连接到私有开发链控制台..."
echo "📂 连接数据目录: $PROJECT_DATA_DIR"
echo "🔌 IPC 文件路径: $IPC_PATH"
echo "💡 使用以下命令可以查看账户信息:"
echo "   eth.accounts"
echo "   eth.getBalance(eth.accounts[0])"
echo "   miner.start()"
echo "   miner.stop()"

# 尝试连接到项目中运行的实例
if [ -S "$IPC_PATH" ]; then
    echo "✅ 找到运行中的 Geth 实例，正在连接..."
    $GETH_PATH attach ipc://$IPC_PATH
else
    echo "❌ 未找到运行中的 Geth 实例"
    echo "请先运行 ./start-dev-chain.sh 启动开发链"
    exit 1
fi
