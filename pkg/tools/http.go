package tools

import (
	"bytes"
	"fmt"
	json "github.com/json-iterator/go"
	"io"
	"net/http"
	"strings"
	"time"
)

type HttpClient struct {
	Host    string
	Timeout time.Duration
}

func NewHttpClient(host string) *HttpClient {
	
	return &HttpClient{
		Host:    host,
		Timeout: 35 * time.Second,
	}
}

func (c *HttpClient) SendBytes(url, method string, body []byte, headers ...map[string]string) (data []byte, code int, err error) {
	
	return c.Send(url, method, bytes.NewReader(body), headers...)
}

func (c *HttpClient) SendStr(url, method, body string, headers ...map[string]string) (data []byte, code int, err error) {
	
	return c.Send(url, method, strings.NewReader(body), headers...)
}

// SendValue 发送struct请求
func (c *HttpClient) SendValue(url, method string, body interface{}, headers ...map[string]string) (data []byte, code int, err error) {
	
	if body == nil {
		
		return c.Send(url, method, nil, headers...)
	}
	
	marshal, err := json.Marshal(body)
	if err != nil {
		
		return
	}
	
	return c.SendBytes(url, method, marshal, headers...)
}

// Send 发送请求
func (c *HttpClient) Send(url, method string, body io.Reader, headers ...map[string]string) (data []byte, code int, err error) {
	
	// 拼接完整地址
	if !strings.HasPrefix(url, "http") {
		
		url = fmt.Sprintf("%s%s", c.Host, url)
	}
	
	// 创建请求
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		
		return
	}
	
	req.Header.Add("content-type", "application/json")
	
	// 处理自定义header
	if len(headers) > 0 {
		
		for k, v := range headers[0] {
			req.Header.Add(k, v)
		}
	}
	
	// 发送请求
	httpClient := http.DefaultClient
	httpClient.Timeout = c.Timeout
	resp, err := httpClient.Do(req)
	if err != nil {
		
		return
	}
	
	code = resp.StatusCode
	
	// 读取
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		
		return
	}
	
	data = respBody
	return
}
