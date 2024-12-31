// Package marketplace Code generated by swaggo/swag. DO NOT EDIT
package marketplace

import "github.com/swaggo/swag"

const docTemplatemarketplace = `{
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
                "parameters": [
                    {
                        "type": "string",
                        "description": "eos account name",
                        "name": "account",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "user balances",
                        "schema": {
                            "$ref": "#/definitions/entity.UserBalance"
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
                        "description": "eos account name",
                        "name": "trader",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "0market 1limit",
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
                                "$ref": "#/definitions/entity.HistoryOrder"
                            }
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
                        "description": "eos account name",
                        "name": "trader",
                        "in": "query",
                        "required": true
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
                        "type": "integer",
                        "description": "Order ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "history order detail",
                        "schema": {
                            "$ref": "#/definitions/entity.HistoryOrderDetail"
                        }
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
        "/credentials": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Get user credentials",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get user credentials",
                "responses": {
                    "200": {
                        "description": "user credentials",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/entity.UserCredential"
                            }
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Create user credentials",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Create user credentials",
                "parameters": [
                    {
                        "description": "create user credential params",
                        "name": "req",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/entity.UserCredential"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/eos/pay-cpu": {
            "post": {
                "description": "pay cpu for user tx",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "eos"
                ],
                "summary": "pay cpu",
                "parameters": [
                    {
                        "description": "signed tx",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/entity.ReqPayCPU"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "txid",
                        "schema": {
                            "$ref": "#/definitions/entity.RespPayCPU"
                        }
                    }
                }
            }
        },
        "/login": {
            "post": {
                "description": "Login",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Login",
                "parameters": [
                    {
                        "description": "login params",
                        "name": "req",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/entity.ReqUserLogin"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        }
    },
    "definitions": {
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
                "timestamp": {
                    "type": "string"
                }
            }
        },
        "entity.HistoryOrder": {
            "type": "object",
            "properties": {
                "avg_price": {
                    "type": "string"
                },
                "executed_amount": {
                    "type": "string"
                },
                "filled_total": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "order_amount": {
                    "type": "string"
                },
                "order_cid": {
                    "type": "string"
                },
                "order_price": {
                    "type": "string"
                },
                "order_time": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "side": {
                    "description": "0 buy 1 sell",
                    "type": "integer"
                },
                "status": {
                    "description": "1partially_filled 2full_filled 3.canceled",
                    "type": "integer"
                },
                "type": {
                    "description": "0 market 1 limit",
                    "type": "integer"
                }
            }
        },
        "entity.HistoryOrderDetail": {
            "type": "object",
            "properties": {
                "avg_price": {
                    "type": "string"
                },
                "executed_amount": {
                    "type": "string"
                },
                "filled_total": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "order_amount": {
                    "type": "string"
                },
                "order_cid": {
                    "type": "string"
                },
                "order_price": {
                    "type": "string"
                },
                "order_time": {
                    "type": "string"
                },
                "pool_id": {
                    "type": "integer"
                },
                "side": {
                    "description": "0 buy 1 sell",
                    "type": "integer"
                },
                "status": {
                    "description": "1partially_filled 2full_filled 3.canceled",
                    "type": "integer"
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
        "entity.OpenOrder": {
            "type": "object",
            "properties": {
                "avg_price": {
                    "type": "string"
                },
                "executed_amount": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "order_amount": {
                    "type": "string"
                },
                "order_cid": {
                    "type": "string"
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
                "pool_id": {
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
                    "type": "integer"
                },
                "max_flct": {
                    "type": "integer"
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
                    "type": "integer"
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
        "entity.ReqPayCPU": {
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
        "entity.ReqUserLogin": {
            "type": "object",
            "properties": {
                "id_token": {
                    "type": "string"
                },
                "method": {
                    "description": "google,apple",
                    "type": "string"
                }
            }
        },
        "entity.RespPayCPU": {
            "type": "object",
            "properties": {
                "transaction_id": {
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
        },
        "entity.UserBalance": {
            "type": "object",
            "properties": {
                "pool_balances": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/entity.UserPoolBalance"
                    }
                },
                "token_balances": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/entity.UserTokenBalance"
                    }
                }
            }
        },
        "entity.UserCredential": {
            "type": "object",
            "properties": {
                "credential_id": {
                    "type": "string"
                },
                "public_key": {
                    "type": "string"
                }
            }
        },
        "entity.UserPoolBalance": {
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
                },
                "token_contract": {
                    "type": "string"
                },
                "token_symbol": {
                    "type": "string"
                }
            }
        },
        "entity.UserTokenBalance": {
            "type": "object",
            "properties": {
                "balance": {
                    "type": "string"
                },
                "contract": {
                    "type": "string"
                },
                "symbol": {
                    "type": "string"
                }
            }
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

// SwaggerInfomarketplace holds exported Swagger Info so clients can modify it
var SwaggerInfomarketplace = &swag.Spec{
	Version:          "1.0",
	Host:             "127.0.0.1:8080",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "exapp-go marketplace api",
	Description:      "",
	InfoInstanceName: "marketplace",
	SwaggerTemplate:  docTemplatemarketplace,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfomarketplace.InstanceName(), SwaggerInfomarketplace)
}
