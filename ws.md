# WebSocket API Documentation

## Overview

We use Socket.IO v4 protocol for WebSocket communication.

## Connection Information

- **WebSocket URL**: `ws://domain/socket.io`
- **Protocol**: Socket.IO v4
- **Ping Interval**: 10 seconds
- **Ping Timeout**: 5 seconds
- **Connection Timeout**: 5 seconds

## Authentication Flow

### 1. User Authentication
Users can authenticate by sending the `authenticate` event:

```javascript 
socket.emit('authenticate', 'account_id');

socket.on('authenticated', (response) => {
    console.log(response);
    // response:
    // {
    //     "status": "success",
    //     "account": "account_id"
    // }
});
```

### 2. Unauthenticate
```javascript
socket.emit('unauthenticate', 'account_id');
```

## Subscription Interface

### 1. Kline Data Subscription
```javascript
socket.emit('subscribe_kline', poolId, interval);

socket.emit('subscribe_kline', 1, '1m');
```

### 2. Depth Data Subscription
```javascript
socket.emit('subscribe_depth', poolId, precision);

socket.emit('subscribe_depth', 1, '0.00000001');
```

### 3. Trades Data Subscription
```javascript
socket.emit('subscribe_trades', poolId);

socket.emit('subscribe_trades', 1);

### 4. Pool Stats Data Subscription
```javascript
socket.emit('subscribe_pool_stats', poolId);

socket.emit('subscribe_pool_stats', 1);
```

### 5. Unsubscribe
```javascript
socket.emit('unsubscribe', subscriptionType, poolId, interval, precision);

socket.emit('unsubscribe', 'kline', 1, '1m');

socket.emit('unsubscribe', 'depth', 1, '0.00000001');

socket.emit('unsubscribe', 'trades', 1);

socket.emit('unsubscribe', 'pool_stats', 1);
```

## Subscription Success Response

When successfully subscribing to a topic, the `subscribed` event is received:

```javascript
socket.on('subscribed', (response) => {
    console.log(response);
    // {
    //     "type": "kline",
    //     "poolID": 1,
    //     "interval": "1m"
    // }
    
    // {
    //     "type": "depth",
    //     "poolID": 1
    //     "precision": "0.00000001"
    // }
    
    // {
    //     "type": "trades",
    //     "poolID": 1
    // }

    // {
    //     "type": "pool_stats",
    //     "poolID": 1
    // }
});
```

## Data Format Description

### 1. Kline Data Format
```javascript
socket.on('kline', (data) => {
    // data:
    // {
    //     "pool_id": 1,
    //     "interval": "1m",
    //     "timestamp": 1234567890,
    //     "open": 100.5,
    //     "high": 101.5,
    //     "low": 99.5,
    //     "close": 100.8,
    //     "volume": 1000.5,
    //     "turnover": 100500.5,
    //     "count": 100
    // }
});
```

### 2. Depth Data Format
```javascript
socket.on('depth', (data) => {
    // data:
    // {
    //     "pool_id": 1,
    //     "timestamp": 1234567890,
    //     "bids": [
    //         ["100.5", "10.5"],  // [price, amount]
    //         ["100.4", "15.2"]
    //     ],
    //     "asks": [
    //         ["100.6", "5.5"],
    //         ["100.7", "8.2"]
    //     ]
    //     "precision": "0.00000001"
    // }
});
```

### 3. Trades Data Format
```javascript
socket.on('trade', (data) => {
    // data:
    // {
    //     "pool_id": 1,
    //     "buyer": "account1",
    //     "seller": "account2",
    //     "quantity": "10.5",
    //     "price": "100.5",
    //     "traded_at": 1234567890,
    //     "side": "buy"  // "buy" or "sell"
    // }
});
```

### 4. Order Update Data Format
```javascript
socket.on('order_update', (data) => {
    // data:
    // {
    //     "account": "account1",
    //     "id": "1-1000-0"  // poolId-orderId-side
    // }
});
```

### 5. Pool Stats Data Format
```javascript
socket.on('pool_stats', (data) => {
    // data:
    // {
    //     "pool_id": 1,
    //     "base_coin": "BTC",
    //     "quote_coin": "USDT",
    //     "symbol": "BTC/USDT",
    //     "last_price": "100.5",
    //     "change": "1.0",
    //     "change_rate": 0.01,
    //     "high": "101.5",
    //     "low": "99.5",
    //     "volume": "1000.5",
    //     "turnover": "100500.5",
    //     "trades": 100,
    //     "updated_at": 1234567890
    // }
});
```

## Example Code

```javascript
const socket = io('ws://domain/socket.io', {
    transports: ['websocket'],
    reconnection: true,
    reconnectionDelay: 1000,
    reconnectionDelayMax: 5000,
    reconnectionAttempts: Infinity
});

socket.on('connect', () => {
    console.log('Connected to WebSocket server');
    
    socket.emit('authenticate', 'account_id');
});

socket.on('authenticated', (response) => {
    console.log('Authentication response:', response);
    
    socket.emit('subscribe_kline', 1, '1m');
    socket.emit('subscribe_depth', 1, '0.00000001');
    socket.emit('subscribe_trades', 1);
});

socket.on('subscribed', (response) => {
    console.log('Subscription response:', response);
});

socket.on('kline', (data) => {
    console.log('Kline update:', data);
});

socket.on('depth', (data) => {
    console.log('Depth update:', data);
});

socket.on('trades', (data) => {
    console.log('Trades update:', data);
});

socket.on('pool_stats', (data) => {
    console.log('Pool stats update:', data);
});

socket.on('order_update', (data) => {
    console.log('Order update:', data);
});

socket.on('balance_update', (data) => {
    console.log('Balance update:', data);
});

socket.on('disconnect', () => {
    console.log('Disconnected from WebSocket server');
});


```
