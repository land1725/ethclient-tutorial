const WebSocket = require('ws');

console.log('🔍 正在测试 Geth WebSocket 连接...');

const ws = new WebSocket('ws://127.0.0.1:8546', [], {
    headers: {
        'Sec-WebSocket-Protocol': 'eth'
    }
});

const testPayload = '{"jsonrpc":"2.0","method":"web3_clientVersion","id":1}';

ws.on('open', function open() {
    console.log('✅ WebSocket 连接成功');
    console.log('📝 发送请求:', testPayload);
    ws.send(testPayload);
});

ws.on('message', function message(data) {
    console.log('📋 响应:', data.toString());
    console.log('🎉 Geth WebSocket 服务正常运行');
    ws.close();
    process.exit(0);
});

ws.on('error', function error(err) {
    console.log('❌ WebSocket 连接失败:', err.message);
    process.exit(1);
});

// 5秒超时
setTimeout(() => {
    console.log('⚠️ WebSocket 连接超时');
    ws.close();
    process.exit(1);
}, 5000);
