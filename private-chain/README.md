# 私有以太坊开发链设置

本目录包含了一个完整的私有以太坊开发链设置，使用最新编译的 Geth 客户端。

## 文件说明

- `genesis.json` - 创世区块配置文件
- `start-dev-chain.sh` - 完整的私有链启动脚本
- `start-dev-simple.sh` - 简化的开发模式启动脚本 (推荐)
- `console.sh` - 连接到私有链的控制台脚本

## 快速开始

### 方法1: 简化开发模式 (推荐)

```bash
# 启动简化开发模式
./start-dev-simple.sh
```

这将启动一个：
- 自动挖矿的私有链
- 预分配资金的开发账户
- HTTP RPC 接口: http://localhost:8545
- WebSocket 接口: ws://localhost:8546
- 支持跨域访问

### 方法2: 完整私有链模式

```bash
# 启动完整私有链
./start-dev-chain.sh
```

这将创建一个更接近主网配置的私有链，使用自定义的创世区块。

## 连接方式

### Web3 连接
```javascript
// JavaScript
const Web3 = require('web3');
const web3 = new Web3('http://localhost:8545');
```

```go
// Go
import "github.com/ethereum/go-ethereum/ethclient"

client, err := ethclient.Dial("http://localhost:8545")
```

### Geth 控制台
```bash
# 连接到运行中的节点控制台
./console.sh
```

## 预配置账户

创世区块中预分配了以下账户 (仅用于开发测试):

- `0xBc694bc8E249956958dBc2529d39bBc94647712F` - 1,000,000 ETH
- `0x6DaEf20BC08855c2eb79b89026d353bd4759aD06` - 1,000,000 ETH

## 网络配置

- Chain ID: 1337
- 网络ID: 1337
- 区块时间: 2-5秒
- Gas Limit: 128,000,000
- 挖矿算法: Clique (PoA) 或 Dev模式

## 数据持久化

- 完整模式数据目录: `./data/`
- 简化模式数据目录: `./dev-data/`
- 日志文件: `./data/geth.log` (仅完整模式)

## 停止节点

使用 `Ctrl+C` 停止节点，数据会自动保存。

## 重新启动

重新运行启动脚本即可恢复之前的区块链状态。
