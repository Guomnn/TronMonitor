package transaction

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	client "go-tron/clients"
	"google.golang.org/protobuf/proto"
	"log"
	"time"
)

func CheckAndTransfer(client *client.Client, accountA, accountC, privateKeyB string) {

	// 配置参数
	const (
		thresholdTRX = 2 * 1e6 // 2 TRX，以 Sun 为单位 (1 TRX = 1e6 Sun)
	)

	// 监控 A 账户的 TRX 余额accountA
	account, err := client.GetTrxBalance(accountA)
	if err != nil {
		log.Printf("\n----->>>>获取账户余额失败: %v", err)
	}

	//获取账户余额
	balanceTRX := account.Balance

	// 打印 A 的当前余额
	fmt.Printf("\n----->>>>A 账户余额: %.6f TRX", float64(balanceTRX)/1e6)

	// 检查余额是否超过阈值
	if balanceTRX > thresholdTRX {
		// 如果余额超过阈值，则触发转账
		fmt.Println("\n----->>>>余额超过阈值。正在触发转账...")

		// 执行多重签名转账
		accountNum := balanceTRX - thresholdTRX

		fmt.Println("\n----->>>>开始时间：", time.Now().Format(time.DateTime))
		err := triggerMultiSigTransfer(client, accountA, accountC, accountNum, privateKeyB)
		fmt.Println("\n----->>>>结束时间：", time.Now().Format(time.DateTime))

		if err != nil {
			log.Printf("\n----->>>>执行转账失败: %v", err)
		} else {
			fmt.Printf("\n=============转账执行成功！TRX: %.6f \n===============", float64(accountNum)/1e6)
		}
	} else {
		fmt.Println("\n----->>>>余额不超过2TRX,不做处理....")
	}

}

// SignTransaction 使用提供的私钥对交易进行签名
func signTransaction(transaction *core.Transaction, privateKey string) (*core.Transaction, error) {
	// 从十六进制格式解码私钥
	privateBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("十六进制解码私钥错误: %v", err)
	}
	// 将私钥字节转换为 ECDSA 私钥
	priv := crypto.ToECDSAUnsafe(privateBytes)
	// 确保在使用后将私钥在内存中清零
	defer zeroKey(priv)
	// 序列化交易的原始数据
	rawData, err := proto.Marshal(transaction.GetRawData())
	if err != nil {
		return nil, fmt.Errorf("proto 序列化交易原始数据错误: %v", err)
	}
	// 计算原始数据的 SHA256 哈希值
	h256h := sha256.New()
	h256h.Write(rawData)
	hash := h256h.Sum(nil)
	// 使用私钥对哈希值进行签名
	signature, err := crypto.Sign(hash, priv)
	if err != nil {
		return nil, fmt.Errorf("签名错误: %v", err)
	}
	// 将签名附加到交易中
	transaction.Signature = append(transaction.Signature, signature)
	return transaction, nil
}

// zeroKey 清零内存中的私钥以确保安全
func zeroKey(k *ecdsa.PrivateKey) {
	b := k.D.Bits()
	for i := range b {
		b[i] = 0
	}
}

// triggerMultiSigTransfer 处理多重签名转账逻辑
func triggerMultiSigTransfer(client *client.Client, from, to string, amount int64, privateKey string) error {
	// 准备 TRX 转账交易
	txExt, err := client.Transfer(from, to, amount)
	if err != nil {
		return fmt.Errorf("交易准备失败: %v", err)
	}
	// 使用私钥对准备好的交易进行签名
	signedTx, err := signTransaction(txExt.Transaction, privateKey)
	if err != nil {
		return fmt.Errorf("交易签名失败: %v", err)
	}
	// 将已签名的交易广播到网络
	err = client.BroadcastTransaction(signedTx)
	if err != nil {
		return fmt.Errorf("广播交易失败: %v", err)
	}
	// 成功广播后返回 nil
	return nil
}
