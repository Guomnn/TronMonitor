package clients

import (
	"encoding/json"
	"fmt"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"google.golang.org/grpc"
	"strings"
)

// Client 结构体，用于管理 GRPC 连接
type Client struct {
	node string
	GRPC *client.GrpcClient
}

// NewClient 初始化一个新的 GRPC 客户端
func NewClient(node string) (*Client, error) {
	c := &Client{
		node: node,
		GRPC: client.NewGrpcClient(node),
	}
	c.GRPC.SetAPIKey("229451d4-b95f-489b-bf2a-d53337b22c30")
	err := c.GRPC.Start(grpc.WithInsecure())

	if err != nil {
		return nil, fmt.Errorf("grpc 客户端启动失败: %v", err)
	}
	return c, nil
}

// keepConnect 保持连接活跃，并在需要时重新连接
func (c *Client) keepConnect() error {

	// 尝试获取节点信息以检查连接性
	_, err := c.GRPC.GetNodeInfo()
	if err != nil {
		// 如果是 "no such host" 错误，则尝试重新连接
		if strings.Contains(err.Error(), "no such host") {
			return c.GRPC.Reconnect(c.node)
		}
		return fmt.Errorf("节点连接错误: %v", err)
	}
	return nil
}

// GetTrxBalance 获取账户的 TRX 余额
func (c *Client) GetTrxBalance(addr string) (*core.Account, error) {
	// 确保连接处于活跃状态
	err := c.keepConnect()
	if err != nil {
		return nil, err
	}
	// 使用 GRPC 客户端获取账户详细信息
	return c.GRPC.GetAccount(addr)
}

func (c *Client) Transfer(from, to string, amount int64) (*api.TransactionExtention, error) {
	// 确保连接处于活跃状态
	err := c.keepConnect()
	if err != nil {
		return nil, err
	}
	// 使用 GRPC 客户端发起 TRX 转账
	return c.GRPC.Transfer(from, to, amount)
}

// BroadcastTransaction 广播已签名的交易
func (c *Client) BroadcastTransaction(transaction *core.Transaction) error {
	// 确保连接处于活跃状态
	err := c.keepConnect()
	if err != nil {
		return err
	}
	// 使用 GRPC 客户端广播已签名的交易
	result, err := c.GRPC.Broadcast(transaction)
	if err != nil {
		return fmt.Errorf("交易广播失败: %v", err)
	}
	// 检查交易结果代码
	if result.Code != 0 {
		return fmt.Errorf("错误的交易: %v", string(result.GetMessage()))
	}
	// 验证交易结果
	if result.Result == true {
		return nil
	}
	// 为错误日志记录序列化结果
	d, _ := json.Marshal(result)
	return fmt.Errorf("交易发送失败: %s", string(d))
}
