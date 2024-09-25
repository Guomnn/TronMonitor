package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	client "go-tron/clients"
	transaction "go-tron/transaction"
	"log"
	"strings"
	"time"
)

func main() {

	//// 从用户输入获取账户和私钥
	//reader := bufio.NewReader(os.Stdin)
	//fmt.Print("请输入 监听 账户地址: ")
	//accountA, _ := reader.ReadString('\n')
	//accountA = strings.TrimSpace(accountA)
	//
	//fmt.Print("请输入 归集的 账户地址: ")
	//accountC, _ := reader.ReadString('\n')
	//accountC = strings.TrimSpace(accountC)
	//
	//fmt.Print("请输入 所有权 账户的私钥: ")
	//privateKeyB, _ := reader.ReadString('\n')
	//privateKeyB = strings.TrimSpace(privateKeyB)

	accountA := "TYx3fs5c8qeSGN28xHtSzJTBg56awKWqmw"
	accountC := "TSYLrqCkdfHCMoEco28uRZxJrTvnN196Fr"
	privateKeyB := ""

	fmt.Println("=================开始监听事件...=================")
	//创建动态bar
	createBar()

	client, err := client.NewClient("grpc.trongrid.io:50051")

	//运行定时任务
	runCron(client, err, accountA, accountC, privateKeyB)

}

var printing bool = true // 全局变量，控制打印状态

func runCron(client *client.Client, errClient error, accountA, accountC, privateKeyB string) {

	c := cron.New()

	// 初始化 GRPC 客户端
	if errClient != nil {
		fmt.Printf("error: %v", errClient)
		log.Fatalf("创建 GRPC 客户端失败: %v", errClient)
	}

	// 添加一个定时任务，这里使用的是 cron 表达式
	_, err := c.AddFunc("@every 5s", func() {
		printing = false // 停止打印 "监听中"
		transaction.CheckAndTransfer(client, accountA, accountC, privateKeyB)
		printing = true // 重新开始打印 "监听中"
	})
	if err != nil {
		fmt.Println("添加任务失败:", err)
		return
	}

	// 启动 cron 调度器
	c.Start()

	// 阻止程序退出
	select {}

}

func createBar() {
	// 添加动态日志
	go func() {

		for {
			if printing {
				for i := 0; i < 10; i++ {
					fmt.Printf("\r监听中%s", strings.Repeat(">", i))
					time.Sleep(100 * time.Millisecond)
				}
			} else {
				time.Sleep(100 * time.Millisecond) // 避免 CPU 空转
			}
		}
	}()
}
