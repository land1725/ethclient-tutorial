#!/bin/bash

# 私有开发链网络测试脚本

echo "🔍 测试私有开发链网络连接..."

# 测试 HTTP RPC 接口
echo "📡 测试 HTTP RPC (localhost:8545)..."
CHAIN_ID=$(curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result')

if [ "$CHAIN_ID" != "null" ] && [ "$CHAIN_ID" != "" ]; then
    echo "✅ HTTP RPC 连接成功！Chain ID: $CHAIN_ID"
else
    echo "❌ HTTP RPC 连接失败"
    exit 1
fi

# 获取区块号
BLOCK_NUMBER=$(curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result')

echo "📦 当前区块号: $BLOCK_NUMBER"

# 获取账户列表
ACCOUNTS=$(curl -s -X POST -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","method":"eth_accounts","params":[],"id":1}' \
  http://localhost:8545 | jq -r '.result[]')

echo "👤 开发账户: $ACCOUNTS"

# 获取账户余额
if [ "$ACCOUNTS" != "" ]; then
    BALANCE=$(curl -s -X POST -H "Content-Type: application/json" \
      --data "{\"jsonrpc\":\"2.0\",\"method\":\"eth_getBalance\",\"params\":[\"$ACCOUNTS\",\"latest\"],\"id\":1}" \
      http://localhost:8545 | jq -r '.result')

    echo "💰 账户余额: $BALANCE wei"

    # 转换为 ETH (简化显示)
    if command -v python3 &> /dev/null; then
        ETH_BALANCE=$(python3 -c "print(int('$BALANCE', 16) / 10**18)")
        echo "💰 账户余额: $ETH_BALANCE ETH"
    fi
fi

echo ""
echo "🎉 私有开发链测试完成！"
echo "🔗 HTTP RPC: http://localhost:8545"
echo "📡 WebSocket: ws://localhost:8546"
echo "📝 可以开始进行智能合约开发和测试了！"
