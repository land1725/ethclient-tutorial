# Ethereum Client Tutorial

这是一个完整的以太坊客户端教程项目，展示了如何使用Go语言与以太坊区块链进行交互。

## 项目结构

```
ethclient_tutorial/
├── main.go                     # 主程序入口
├── go.mod                      # Go模块配置
├── .env                        # 环境变量配置（敏感信息）
├── .env.example               # 环境变量模板
├── .gitignore                 # Git忽略文件
├── config/                    # 配置管理
│   └── config.go
├── account_balance/            # 账户余额查询
├── block_query/                # 区块查询功能
├── block_subscription/         # 区块订阅功能
├── contract_deployment/        # 智能合约部署
├── contract_events/            # 合约事件监听
├── contract_execution/         # 合约执行
├── contract_loader/            # 合约加载
├── eth_transfer/               # ETH转账功能
├── receipt_query/              # 交易收据查询
├── token_balance/              # Token余额查询
├── token_transfer/             # ERC20 Token转账
├── transaction_query/          # 交易查询功能
└── wallet_management/          # 钱包管理功能
```

## 环境配置

### 1. 配置环境变量

首先复制环境变量模板：

```bash
cp .env.example .env
```

然后编辑 `.env` 文件，填入您的配置信息：

```bash
# 必需配置
INFURA_PROJECT_ID=your_infura_project_id_here
ETHEREUM_NETWORK=sepolia

# 可选配置（用于转账功能测试）
TEST_PRIVATE_KEY=your_test_private_key_here
TEST_RECIPIENT_ADDRESS=0x742d35Cc6634C0532925a3b8D4C9db96C5C7F4C1
```

### 2. 获取Infura Project ID

1. 访问 [Infura](https://infura.io/) 并注册账户
2. 创建新项目
3. 复制项目ID到 `.env` 文件中的 `INFURA_PROJECT_ID`

### 3. 网络选择

支持的网络：
- `mainnet` - 以太坊主网（生产环境）
- `sepolia` - Sepolia测试网（推荐用于测试）
- `goerli` - Goerli测试网

**⚠️ 强烈建议在测试时使用测试网络！**

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入您的配置
```

### 3. 运行演示程序

```bash
go run main.go
```

## 功能特性

### 🔐 安全特性
- ✅ 环境变量管理敏感信息
- ✅ .gitignore保护配置文件  
- ✅ 生产/测试环境分离
- ✅ 私钥安全提醒

### 💼 钱包功能
- ✅ 创建新钱包
- ✅ 地址验证
- ✅ 私钥管理

### 🌐 网络功能（需要API密钥）
- ✅ 区块查询
- ✅ 交易查询
- ✅ 交易收据查询
- ✅ ETH转账
- ✅ 智能合约交互

## 环境变量说明

| 变量名 | 必需 | 说明 | 示例 |
|-------|------|------|------|
| `INFURA_PROJECT_ID` | ✅ | Infura项目ID | `abc123...` |
| `ETHEREUM_NETWORK` | ✅ | 以太坊网络 | `sepolia` |
| `TEST_PRIVATE_KEY` | ❌ | 测试私钥 | `0x123...` |
| `TEST_RECIPIENT_ADDRESS` | ❌ | 测试接收地址 | `0x742d35...` |
| `DEFAULT_GAS_LIMIT` | ❌ | 默认Gas限制 | `21000` |
| `GAS_PRICE_MULTIPLIER` | ❌ | Gas价格倍数 | `1.1` |

## 使用示例

### 基本功能（无需网络）
```bash
# 运行钱包创建演示
go run main.go
```

### 网络功能（需要配置API密钥）
```bash
# 确保已配置 INFURA_PROJECT_ID
echo "INFURA_PROJECT_ID=your_project_id" >> .env
go run main.go
```

## 安全注意事项

⚠️ **重要安全提醒**：

1. **私钥保护**
   - 永远不要在代码中硬编码私钥
   - 使用 `.env` 文件管理敏感信息
   - `.env` 文件已添加到 `.gitignore`

2. **网络选择**
   - 测试时使用测试网络（Sepolia、Goerli）
   - 生产环境谨慎操作主网

3. **版本控制**
   - 不要提交 `.env` 文件到Git
   - 使用 `.env.example` 作为模板

4. **权限管理**
   - 测试私钥仅用于测试目的
   - 生产环境使用硬件钱包或HSM

## 示例输出

### 配置API密钥前：
```
Ethereum Client Tutorial
========================

1. 演示钱包创建:
Address: 0x3f5D372D47054209ac09349e3B1dEC2176b26C1F
Private Key: b168d089f36c5ede23125aa1cb76232a0838f5896a1fe4c5ce00a41e76b651df

注意：区块查询、交易查询等功能需要真实的以太坊网络连接
请在 .env 文件中设置 INFURA_PROJECT_ID 来启用网络功能。
示例:
  INFURA_PROJECT_ID=your_project_id_here
```

### 配置API密钥后：
```
Ethereum Client Tutorial
========================

1. 演示钱包创建:
Address: 0x7890...
Private Key: abc123...

2. 连接到以太坊网络 (sepolia)...
✅ 成功连接到以太坊网络!

3. 网络功能演示:
Block #12345: 0xabc123...
Tx 0x456def... => Value: 0.1000 ETH
Receipt: Status=1, GasUsed=21000
```

## 依赖包

- `github.com/ethereum/go-ethereum`: 以太坊Go客户端库
- `github.com/joho/godotenv`: 环境变量管理

## 贡献

欢迎提交Issue和Pull Request来改进这个教程项目。

## 许可证

MIT License
