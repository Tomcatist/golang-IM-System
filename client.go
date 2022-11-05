package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp string
	ServerPort int
	Name string
	conn net.Conn
	flag int  // 当前客户端的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp: serverIp,
		ServerPort: serverPort,
		flag: 999,
	}
	// 链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}
	client.conn = conn
	// 返回对象
	return client
}

// 处理server返回的消息，直接返回到标准输出中
func (client *Client) DealResponse() {
	// 一旦client,conn有消息，就直接copy的std标准输出上，永久监听
	io.Copy(os.Stdout, client.conn)
}

// 菜单函数
func (client *Client) menu() bool {

	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >=0 && flag <=3 {
		client.flag = flag
		return true
	}else {
		fmt.Println(">>>>> 请输入合法范围内的数字 <<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		// 根据不同模式，处理不同业务
		switch client.flag {
		case 1:
			// 公聊模式
			fmt.Println("公聊模式选择...")
			break
		case 2:
			// 私聊模式
			fmt.Println("私聊模式选择...")
			break
		case 3:
			// 更新用户名
			fmt.Println("更新用户名选择选择...")
			client.updateName()
			break
		}
	}
}


var serverIp string
var serverPort int
// 初始化client 读取命令行ip和port 格式 ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认为127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认是8888）")
}

// 更新用户名方法
func (client *Client) updateName() bool {
	fmt.Println(">>>>>请输入用户名")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}


func main() {
	// 命令行解析
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>>>>> 服务器链接失败")
		return
	}
	fmt.Println(">>>>>>>> 服务器链接成功")
	// 启动server消息监听
	go client.DealResponse()

	// 启动客户端的业务
	client.Run()

}
