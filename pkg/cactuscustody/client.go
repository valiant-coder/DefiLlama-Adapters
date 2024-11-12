package cactuscustody

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CactusClient struct {
	BaseURL    string
	AKId       string
	PrivateKey *ecdsa.PrivateKey
	ApiKey     string
}

func NewCactusClient(baseURL, akId string, privateKey *ecdsa.PrivateKey, apiKey string) *CactusClient {
	return &CactusClient{
		BaseURL:    baseURL,
		AKId:       akId,
		PrivateKey: privateKey,
		ApiKey:     apiKey,
	}
}


func (c *CactusClient) generateSignature(contentToSign string) (string, error) {
	hash := sha256.Sum256([]byte(contentToSign))
	r, s, err := ecdsa.Sign(rand.Reader, c.PrivateKey, hash[:])
	if err != nil {
		return "", err
	}
	signature := append(r.Bytes(), s.Bytes()...)
	return base64.StdEncoding.EncodeToString(signature), nil
}


func (c *CactusClient) constructContentToSign(method, uri string, params map[string]string, body []byte, nonce string) string {
	accept := "application/json"
	contentType := "application/json"
	date := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")

	var contentSHA256 string
	if method == "POST" || method == "PUT" || method == "PATCH" {
		hash := sha256.Sum256(body)
		contentSHA256 = base64.StdEncoding.EncodeToString(hash[:])
	}

	var paramPairs []string
	for k, v := range params {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=[%s]", k, v))
	}
	sort.Strings(paramPairs)
	paramString := strings.Join(paramPairs, ", ")

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\nx-api-key:%s\nx-api-nonce:%s\n%s?{%s}",
		method, accept, contentSHA256, contentType, date, c.ApiKey, nonce, uri, paramString)
}


func (c *CactusClient) SendRequest(method, uri string, params map[string]string, body []byte) (*http.Response, error) {
	nonce := uuid.New().String()
	contentToSign := c.constructContentToSign(method, uri, params, body,nonce)
	signature, err := c.generateSignature(contentToSign)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, c.BaseURL+uri, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", c.ApiKey)
	req.Header.Set("x-api-nonce", nonce)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	req.Header.Set("Authorization", fmt.Sprintf("api %s:%s", c.AKId, signature))

	if method == "POST" || method == "PUT" || method == "PATCH" {
		hash := sha256.Sum256(body)
		req.Header.Set("Content-SHA256", base64.StdEncoding.EncodeToString(hash[:]))
	}

	client := &http.Client{}
	return client.Do(req)
}


