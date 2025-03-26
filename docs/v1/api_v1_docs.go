// Package v1 Code generated by swaggo/swag. DO NOT EDIT
package v1

import "github.com/swaggo/swag"

const docTemplateapi_v1 = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/balances": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get user balances",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get user balances",
                "responses": {
                    "200": {
                        "description": "user balances",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.SubAccountBalance"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/depth": {
            "get": {
                "description": "Get order book by pool id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "depth"
                ],
                "summary": "Get depth",
                "parameters": [
                    {
                        "type": "string",
                        "description": "pool_id",
                        "name": "pool_id",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "0.00000001 ~ 10000",
                        "name": "precision",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "limit",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "order depth",
                        "schema": {
                            "$ref": "#/definitions/entity.Depth"
                        }
                    }
                }
            }
        },
        "/api/v1/history-orders": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get history orders",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "order"
                ],
                "summary": "Get history orders",
                "parameters": [
                    {
                        "type": "string",
                        "description": "pool_id",
                        "name": "pool_id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "0 buy 1 sell",
                        "name": "side",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "0 market 1limit",
                        "name": "type",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "status",
                        "name": "status",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "history orders",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.Order"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/info": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get user info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get user info",
                "responses": {
                    "200": {
                        "description": "user info",
                        "schema": {
                            "$ref": "#/definitions/entity.RespV1UserInfo"
                        }
                    }
                }
            }
        },
        "/api/v1/klines": {
            "get": {
                "description": "Get kline data by pool id and interval",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "kline"
                ],
                "summary": "Get kline data",
                "parameters": [
                    {
                        "type": "string",
                        "description": "pool id",
                        "name": "pool_id",
                        "in": "query",
                        "required": true
                    },
                    {
                        "enum": [
                            "1m",
                            "5m",
                            "15m",
                            "30m",
                            "1h",
                            "4h",
                            "1d",
                            "1w",
                            "1M"
                        ],
                        "type": "string",
                        "description": "interval",
                        "name": "interval",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "start timestamp",
                        "name": "start",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "end timestamp",
                        "name": "end",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "kline data",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.Kline"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/latest-trades": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get latest trades",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "trade"
                ],
                "summary": "Get latest trades",
                "parameters": [
                    {
                        "type": "string",
                        "description": "pool_id",
                        "name": "pool_id",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "limit count",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "trade list",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.Trade"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/open-orders": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get open orders",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "order"
                ],
                "summary": "Get open orders",
                "parameters": [
                    {
                        "type": "string",
                        "description": "pool_id",
                        "name": "pool_id",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "0 buy 1 sell",
                        "name": "side",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "open order list",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.OpenOrder"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/orders/{id}": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get history order detail",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "order"
                ],
                "summary": "Get history order detail",
                "parameters": [
                    {
                        "type": "string",
                        "description": "pool_id+order_id+side,ps:0-1-0 pool_id = 0,order_id = 1,side = buy",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "history order detail",
                        "schema": {
                            "$ref": "#/definitions/entity.OrderDetail"
                        }
                    }
                }
            }
        },
        "/api/v1/ping": {
            "get": {
                "description": "Ping",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "Ping",
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/api/v1/pools": {
            "get": {
                "description": "Get a list of all trading pools",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pools"
                ],
                "summary": "List all trading pools",
                "parameters": [
                    {
                        "type": "string",
                        "description": "base coin",
                        "name": "base_coin",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "quote coin",
                        "name": "quote_coin",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "pool info",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.PoolStats"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/pools/{symbolOrId}": {
            "get": {
                "description": "Get detailed information about a specific trading pool",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "pools"
                ],
                "summary": "Get pool details",
                "parameters": [
                    {
                        "type": "string",
                        "description": "pool symbol or pool id",
                        "name": "symbolOrId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entity.Pool"
                        }
                    }
                }
            }
        },
        "/api/v1/system-info": {
            "get": {
                "description": "Get system information",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "Get system information",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entity.RespSystemInfo"
                        }
                    }
                }
            }
        },
        "/api/v1/token/{symbol}": {
            "get": {
                "description": "Get token info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "token"
                ],
                "summary": "Get token info",
                "parameters": [
                    {
                        "type": "string",
                        "description": "coin symbol,ps BTC",
                        "name": "symbol",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "token info",
                        "schema": {
                            "$ref": "#/definitions/entity.Token"
                        }
                    }
                }
            }
        },
        "/api/v1/tokens": {
            "get": {
                "description": "Get support tokens",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "token"
                ],
                "summary": "Get support tokens",
                "responses": {
                    "200": {
                        "description": "token list",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.Token"
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/tx": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "send  tx",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "tx"
                ],
                "summary": "send tx",
                "parameters": [
                    {
                        "description": "signed tx",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/entity.ReqSendTx"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "txid",
                        "schema": {
                            "$ref": "#/definitions/entity.RespSendTx"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "entity.Chain": {
            "type": "object",
            "properties": {
                "chain_id": {
                    "type": "integer"
                },
                "chain_name": {
                    "type": "string"
                },
                "exsat_token_address": {
                    "type": "string"
                },
                "exsat_withdraw_fee": {
                    "type": "string"
                },
                "min_deposit_amount": {
                    "type": "string"
                },
                "min_withdraw_amount": {
                    "type": "string"
                },
                "withdraw_fee": {
                    "type": "string"
                }
            }
        },
        "entity.Depth": {
            "type": "object",
            "properties": {
                "asks": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        }
                    }
                },
                "bids": {
                    "type": "array",
                    "items": {
                        "type": "array",
                        "items": {
                            "type": "string"
                        }
                    }
                },
                "pool_id": {
                    "type": "integer"
                },
                "precision": {
                    "type": "string"
                },
                "timestamp": {
                    "type": "integer"
                }
            }
        },
        "entity.Kline": {
            "type": "object",
            "properties": {
                "close": {
                    "type": "number"
                },
                "count": {
                    "type": "integer"
                },
                "high": {
                    "type": "number"
                },
                "interval": {
                    "type": "string"
                },
                "low": {
                    "type": "number"
                },
                "open": {
                    "type": "number"
                },
                "pool_id": {
                    "type": "integer"
                },
                "timestamp": {
                    "type": "string"
                },
                "turnover": {
                    "type": "number"
                },
                "volume": {
                    "type": "number"
                }
            }
        },
        "entity.LockBalance": {
            "type": "object",
            "properties": {
                "balance": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "pool_symbol": {
                    "type": "string"
                }
            }
        },
        "entity.OpenOrder": {
            "type": "object",
            "properties": {
                "avg_price": {
                    "type": "string"
                },
                "base_coin_precision": {
                    "type": "integer"
                },
                "executed_amount": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "order_amount": {
                    "type": "string"
                },
                "order_cid": {
                    "type": "string"
                },
                "order_id": {
                    "type": "integer"
                },
                "order_price": {
                    "type": "string"
                },
                "order_time": {
                    "type": "string"
                },
                "order_total": {
                    "type": "string"
                },
                "pool_base_coin": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "pool_quote_coin": {
                    "type": "string"
                },
                "pool_symbol": {
                    "type": "string"
                },
                "quote_coin_precision": {
                    "type": "integer"
                },
                "side": {
                    "description": "0 buy 1 sell",
                    "type": "integer"
                },
                "trader": {
                    "type": "string"
                },
                "type": {
                    "description": "0 market 1 limit",
                    "type": "integer"
                }
            }
        },
        "entity.Order": {
            "type": "object",
            "properties": {
                "avg_price": {
                    "type": "string"
                },
                "base_coin_precision": {
                    "type": "integer"
                },
                "executed_amount": {
                    "type": "string"
                },
                "filled_total": {
                    "type": "string"
                },
                "history": {
                    "type": "boolean"
                },
                "id": {
                    "type": "string"
                },
                "order_amount": {
                    "type": "string"
                },
                "order_cid": {
                    "type": "string"
                },
                "order_id": {
                    "type": "integer"
                },
                "order_price": {
                    "type": "string"
                },
                "order_time": {
                    "type": "string"
                },
                "order_total": {
                    "type": "string"
                },
                "permission": {
                    "type": "string"
                },
                "pool_base_coin": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "pool_quote_coin": {
                    "type": "string"
                },
                "pool_symbol": {
                    "type": "string"
                },
                "quote_coin_precision": {
                    "type": "integer"
                },
                "side": {
                    "description": "0 buy 1 sell",
                    "type": "integer"
                },
                "status": {
                    "description": "0 open 1partially_filled 2full_filled 3.canceled",
                    "type": "integer"
                },
                "trader": {
                    "type": "string"
                },
                "type": {
                    "description": "0 market 1 limit",
                    "type": "integer"
                },
                "unread": {
                    "type": "boolean"
                }
            }
        },
        "entity.OrderDetail": {
            "type": "object",
            "properties": {
                "avg_price": {
                    "type": "string"
                },
                "base_coin_precision": {
                    "type": "integer"
                },
                "executed_amount": {
                    "type": "string"
                },
                "filled_total": {
                    "type": "string"
                },
                "history": {
                    "type": "boolean"
                },
                "id": {
                    "type": "string"
                },
                "order_amount": {
                    "type": "string"
                },
                "order_cid": {
                    "type": "string"
                },
                "order_id": {
                    "type": "integer"
                },
                "order_price": {
                    "type": "string"
                },
                "order_time": {
                    "type": "string"
                },
                "order_total": {
                    "type": "string"
                },
                "permission": {
                    "type": "string"
                },
                "pool_base_coin": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "pool_quote_coin": {
                    "type": "string"
                },
                "pool_symbol": {
                    "type": "string"
                },
                "quote_coin_precision": {
                    "type": "integer"
                },
                "side": {
                    "description": "0 buy 1 sell",
                    "type": "integer"
                },
                "status": {
                    "description": "0 open 1partially_filled 2full_filled 3.canceled",
                    "type": "integer"
                },
                "trader": {
                    "type": "string"
                },
                "trades": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/entity.TradeDetail"
                    }
                },
                "type": {
                    "description": "0 market 1 limit",
                    "type": "integer"
                },
                "unread": {
                    "type": "boolean"
                }
            }
        },
        "entity.Pool": {
            "type": "object",
            "properties": {
                "asking_time": {
                    "type": "string"
                },
                "base_coin": {
                    "type": "string"
                },
                "base_coin_precision": {
                    "type": "integer"
                },
                "base_contract": {
                    "type": "string"
                },
                "base_symbol": {
                    "type": "string"
                },
                "maker_fee_rate": {
                    "type": "number"
                },
                "max_flct": {
                    "type": "integer"
                },
                "min_amount": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "pool_stats": {
                    "$ref": "#/definitions/entity.PoolStats"
                },
                "price_precision": {
                    "type": "integer"
                },
                "quote_coin": {
                    "type": "string"
                },
                "quote_coin_precision": {
                    "type": "integer"
                },
                "quote_contract": {
                    "type": "string"
                },
                "quote_symbol": {
                    "type": "string"
                },
                "status": {
                    "type": "integer"
                },
                "symbol": {
                    "type": "string"
                },
                "taker_fee_rate": {
                    "type": "number"
                },
                "trading_time": {
                    "type": "string"
                }
            }
        },
        "entity.PoolStats": {
            "type": "object",
            "properties": {
                "base_coin": {
                    "type": "string"
                },
                "change": {
                    "type": "string"
                },
                "change_rate": {
                    "type": "number"
                },
                "high": {
                    "type": "string"
                },
                "last_price": {
                    "type": "string"
                },
                "low": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "quote_coin": {
                    "type": "string"
                },
                "symbol": {
                    "type": "string"
                },
                "trades": {
                    "type": "integer"
                },
                "turnover": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                },
                "volume": {
                    "type": "string"
                }
            }
        },
        "entity.ReqSendTx": {
            "type": "object",
            "required": [
                "single_signed_transaction"
            ],
            "properties": {
                "single_signed_transaction": {
                    "type": "string"
                }
            }
        },
        "entity.RespSendTx": {
            "type": "object",
            "properties": {
                "transaction_id": {
                    "type": "string"
                }
            }
        },
        "entity.RespSystemInfo": {
            "type": "object",
            "properties": {
                "app_contract": {
                    "type": "string"
                },
                "pay_eos_account": {
                    "type": "string"
                },
                "token_contract": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "entity.RespV1UserInfo": {
            "type": "object",
            "properties": {
                "eos_account": {
                    "type": "string"
                },
                "permission": {
                    "type": "string"
                }
            }
        },
        "entity.SubAccountBalance": {
            "type": "object",
            "properties": {
                "balance": {
                    "type": "string"
                },
                "coin": {
                    "type": "string"
                },
                "locked": {
                    "type": "string"
                },
                "locks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/entity.LockBalance"
                    }
                },
                "usdt_price": {
                    "type": "string"
                }
            }
        },
        "entity.Token": {
            "type": "object",
            "properties": {
                "decimals": {
                    "type": "integer"
                },
                "eos_contract": {
                    "type": "string"
                },
                "icon_url": {
                    "type": "string"
                },
                "info": {
                    "$ref": "#/definitions/entity.TokenInfo"
                },
                "name": {
                    "type": "string"
                },
                "support_chain": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/entity.Chain"
                    }
                },
                "symbol": {
                    "type": "string"
                }
            }
        },
        "entity.TokenInfo": {
            "type": "object",
            "properties": {
                "circulating_supply": {
                    "type": "string"
                },
                "fully_diluted_market_cap": {
                    "type": "string"
                },
                "historical_high": {
                    "type": "string"
                },
                "historical_high_date": {
                    "type": "string"
                },
                "historical_low": {
                    "type": "string"
                },
                "historical_low_date": {
                    "type": "string"
                },
                "intro": {
                    "type": "string"
                },
                "issue_date": {
                    "type": "string"
                },
                "links": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/entity.TokenLink"
                    }
                },
                "market_capitalization": {
                    "type": "string"
                },
                "market_dominance": {
                    "type": "string"
                },
                "maximum_supply": {
                    "type": "string"
                },
                "rank": {
                    "type": "string"
                },
                "total_supply": {
                    "type": "string"
                },
                "volume": {
                    "type": "string"
                },
                "volume_div_market_cap": {
                    "type": "string"
                }
            }
        },
        "entity.TokenLink": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "entity.Trade": {
            "type": "object",
            "properties": {
                "buyer": {
                    "type": "string"
                },
                "buyer_permission": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "price": {
                    "type": "string"
                },
                "quantity": {
                    "type": "string"
                },
                "seller": {
                    "type": "string"
                },
                "seller_permission": {
                    "type": "string"
                },
                "side": {
                    "$ref": "#/definitions/entity.TradeSide"
                },
                "traded_at": {
                    "type": "string"
                }
            }
        },
        "entity.TradeDetail": {
            "type": "object",
            "properties": {
                "base_quantity": {
                    "type": "string"
                },
                "maker": {
                    "type": "string"
                },
                "maker_app_fee": {
                    "type": "string"
                },
                "maker_fee": {
                    "type": "string"
                },
                "maker_order_cid": {
                    "type": "string"
                },
                "maker_order_id": {
                    "type": "integer"
                },
                "pool_id": {
                    "type": "integer"
                },
                "price": {
                    "type": "string"
                },
                "quote_quantity": {
                    "type": "string"
                },
                "taker": {
                    "type": "string"
                },
                "taker_app_fee": {
                    "type": "string"
                },
                "taker_fee": {
                    "type": "string"
                },
                "taker_is_bid": {
                    "type": "boolean"
                },
                "taker_order_cid": {
                    "type": "string"
                },
                "taker_order_id": {
                    "type": "integer"
                },
                "timestamp": {
                    "type": "string"
                },
                "tx_id": {
                    "type": "string"
                }
            }
        },
        "entity.TradeSide": {
            "type": "string",
            "enum": [
                "buy",
                "sell"
            ],
            "x-enum-varnames": [
                "TradeSideBuy",
                "TradeSideSell"
            ]
        }
    },
    "securityDefinitions": {
        "ApiKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfoapi_v1 holds exported Swagger Info so clients can modify it
var SwaggerInfoapi_v1 = &swag.Spec{
	Version:          "1.0",
	Host:             "127.0.0.1:8080",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "exapp api v1",
	Description:      "",
	InfoInstanceName: "api_v1",
	SwaggerTemplate:  docTemplateapi_v1,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfoapi_v1.InstanceName(), SwaggerInfoapi_v1)
}
