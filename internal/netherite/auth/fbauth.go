// Package auth 实现网易租赁服认证客户端。
// 对接 FastBuilder 认证服务器，支持 FB Token 登录。
package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// 默认认证服务器
const DefaultAuthServer = "https://user.fastbuilder.pro"

// 允许的认证服务器域名
var allowedHosts = map[string]bool{
	"user.fastbuilder.pro": true,
	"liliya233.uk":        true,
	"localhost":            true,
	"127.0.0.1":           true,
}

// Client FB 认证客户端。
type Client struct {
	authServer string
	secret     string
	httpClient *http.Client
}

// Options 连接选项。
type Options struct {
	AuthServer string // 认证服务器 URL
}

// DefaultOptions 默认选项。
func DefaultOptions() *Options {
	return &Options{
		AuthServer: DefaultAuthServer,
	}
}

// NewClient 创建认证客户端。
// 内部流程：
//  1. GET /api/new → 拿到 secret bearer token
//  2. 用 secret 请求登录接口
func NewClient(opts *Options) (*Client, error) {
	if opts.AuthServer == "" {
		opts.AuthServer = DefaultAuthServer
	}

	// 校验服务器域名
	parsedURL, err := url.Parse(opts.AuthServer)
	if err != nil {
		return nil, fmt.Errorf("invalid auth server url: %w", err)
	}
	if !allowedHosts[parsedURL.Hostname()] {
		return nil, fmt.Errorf("auth server hostname not allowed: %s (allowed: user.fastbuilder.pro, liliya233.uk, localhost)", parsedURL.Hostname())
	}

	client := &Client{
		authServer: opts.AuthServer,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Step 1: 获取 secret token
	secret, err := client.fetchSecret()
	if err != nil {
		return nil, fmt.Errorf("fetch secret: %w", err)
	}
	client.secret = secret

	return client, nil
}

// fetchSecret GET /api/new 拿 secret
func (c *Client) fetchSecret() (string, error) {
	resp, err := c.httpClient.Get(c.authServer + "/api/new")
	if err != nil {
		return "", fmt.Errorf("request /api/new: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == 503 {
		return "", errors.New("auth server is temporarily unavailable (503)")
	}
	if resp.StatusCode != 200 {
		return "", c.parseError(string(body))
	}

	return string(bytes.TrimSpace(body)), nil
}

// LoginRequest 登录请求。
type LoginRequest struct {
	// 登录凭据（二选一）
	LoginToken string // FB Token（推荐）
	Username   string // 账号
	Password   string // 密码（不推荐）

	// 服务器信息
	ServerCode    string // 房间号/服务器码
	ServerPasscode string // 房间密码（可空）

	// 客户端公钥（ECDH P-384）
	ClientPublicKey []byte
}

// LoginResponse 登录响应。
type LoginResponse struct {
	// 通用
	Success bool
	Message string
	IPAddress string `json:"ip_address"` // Raknet 连接地址

	// 身份信息
	ChainInfo     string `json:"chain_info"`
	IdentityData  string `json:"identity_data"`  // JSON，base64 编码的 NeteaseIdentity
	NeteaseSid    string `json:"netease_sid"`  // 用于连接
	UID           int64  `json:"uid"`           // 用户 ID

	// 外观信息
	OutfitInfo map[string]any `json:"outfit_info"` // mod 列表
}

// Login 登录租赁服。
// POST /api/phoenix/login
func (c *Client) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// 构造请求体
	body := make(map[string]any)
	if req.LoginToken != "" {
		body["login_token"] = req.LoginToken
	} else if req.Username != "" {
		body["username"] = req.Username
		body["password"] = req.Password
	} else {
		return nil, errors.New("must provide login_token or username/password")
	}

	body["server_code"] = req.ServerCode
	body["server_passcode"] = req.ServerPasscode
	if len(req.ClientPublicKey) > 0 {
		body["client_public_key"] = req.ClientPublicKey
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.authServer+"/api/phoenix/login",
		bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.secret))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request /api/phoenix/login: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode == 503 {
		return nil, errors.New("auth server is temporarily unavailable (503)")
	}
	if resp.StatusCode != 200 {
		return nil, c.parseError(string(bodyBytes))
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return nil, fmt.Errorf("login failed: %s", msg)
	}

	// 提取字段
	loginResp := &LoginResponse{
		Success: true,
	}

	if v, ok := result["ip_address"].(string); ok {
		loginResp.IPAddress = v
	}
	if v, ok := result["chain_info"].(string); ok {
		loginResp.ChainInfo = v
	}
	if v, ok := result["identity_data"].(string); ok {
		loginResp.IdentityData = v
	}
	if v, ok := result["netease_sid"].(string); ok {
		loginResp.NeteaseSid = v
	}
	if v, ok := result["uid"].(float64); ok {
		loginResp.UID = int64(v)
	}
	if v, ok := result["outfit_info"].(map[string]any); ok {
		loginResp.OutfitInfo = v
	}

	return loginResp, nil
}

// TransferStartType 获取进服验证题目
// GET /api/phoenix/transfer_start_type?content=xxx
func (c *Client) TransferStartType(ctx context.Context, content string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/phoenix/transfer_start_type?content=%s", c.authServer, content),
		nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.secret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", c.parseError(string(body))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return "", fmt.Errorf("transfer_start_type: %s", msg)
	}

	data, _ := result["data"].(string)
	return data, nil
}

// TransferCheckNum 提交进服验证答案
// POST /api/phoenix/transfer_check_num
func (c *Client) TransferCheckNum(ctx context.Context, data string) (string, error) {
	body := map[string]string{"data": data}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.authServer+"/api/phoenix/transfer_check_num",
		bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.secret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", c.parseError(string(bodyBytes))
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	success, _ := result["success"].(bool)
	if !success {
		msg, _ := result["message"].(string)
		return "", fmt.Errorf("transfer_check_num: %s", msg)
	}

	val, _ := result["value"].(string)
	return val, nil
}

// parseError 解析错误信息
func (c *Client) parseError(body string) error {
	// 格式: "XXX Description\nDetails"
	re := regexp.MustCompile(`^(HTTP \d{3} [A-Za-z ]+)\n?(.*)`)
	matches := re.FindStringSubmatch(body)
	if len(matches) >= 2 {
		return fmt.Errorf("%s: %s", matches[1], trimLine(body, 2))
	}
	return fmt.Errorf("auth error: %s", trimLine(body, 0))
}

func trimLine(s string, n int) string {
	lines := regexp.MustCompile(`\n`).Split(s, -1)
	if n < len(lines) {
		return lines[n]
	}
	return s
}

// AuthServer 返回认证服务器地址
func (c *Client) AuthServer() string {
	return c.authServer
}
