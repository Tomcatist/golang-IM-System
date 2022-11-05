package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex

	// 消息广播的channel
	Message chan string
}

// 创建一个serber的对外函数接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}

	return server
}

// 监听Message广播消息channel的goroutine，一旦有消息就发送给全部在线的User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		// 将message发送给全部在线user
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播消息的方法
func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

// 消息的handler
func (this *Server) Handler(conn net.Conn) {
	// ...当前链接的业务
	//fmt.Println("链接建立成功")
	user := NewUser(conn, this)

	// 用户上线，将用户加入到OnlineMap中
	user.Online()

	// 监听用户是否活跃的channel
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 提取用户的消息
			msg := string(buf[:n-1])

			// 用户针对消息进行处理
			user.DoMessage(msg)

			// 用户的任意操作，代表用户是活跃的
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
			case <-isLive:
				// 当前用户是活跃的，需要重置定时器
				// 不做任何事情，为了激活select，更新下面的定时器
			case <- time.After(time.Second * 100):
				// 已经超时了,将当前的user强制关闭
				user.SendMessage("你被踢了")
				// 销毁用户的资源
				close(user.C)
				user.conn.Close()
				// 退出当前handler
				return
		}
	}
}



// 启动服务器的接口
func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go this.Handler(conn)
	}
}

