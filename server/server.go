package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"net"
	"os"
	"strings"
	"strconv"
	"time"
	"math"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	chat "ChatRoom4435/proto"
)

//客户端管理
type ClientManager struct {
	//客户端 map 储存并管理所有的长连接client，在线的为true，不在的为false
	clients map[*Client]bool
	//web端发送来的的message我们用broadcast来接收，并最后分发给所有的client
	broadcast chan []byte
	//新创建的长连接client
	register chan *Client
	//新注销的长连接client
	unregister chan *Client
}

type node struct {
	// Self information
	Name       string
	Addr       string
	NodeName   string
	NodeNumber int
	NodeMap   map[string]int64
	TimeStamp      int64
	Todos     []data
	// Consul related variables
	SDAddress string
	SDKV      api.KV
	// used to make requests
	Clients map[string]chat.ChatClient
}

var (
	nd node
)


//客户端 Client
type Client struct {
	//用户id
	id string
	//连接的socket
	socket *websocket.Conn
	//发送的消息
	send chan []byte
}

//会把Message格式化成json
type Message struct {
	//消息struct
	Name    string `json:"name,omitempty"`    
	Photo string `json:"photo,omitempty"` 
	Content   string `json:"content,omitempty"`   
	Event string `json:"event,omitempty"` 
	Img64 string `json:"img_64,omitempty"` 
}

type data struct {
	Name    string
	Photo string
	Content   string 
	Event string
	Img64 string 
	TimeStamp int64
	Number int
	UserName string
}

//创建客户端管理者
var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (n *node) SendMessage(ctx context.Context, in *chat.Message) (*chat.MessageReply, error) {
	log.Printf("Received: %v, %v, %v", in.Name, in.Event, in.Content)
	n.NodeMap[in.Name] = in.Timestamp
	maxTime := math.Max(float64(n.TimeStamp), float64(in.Timestamp)) + 1
	n.TimeStamp = int64(maxTime)
	n.NodeMap[n.Name] = n.TimeStamp
	var (
		cnt string
		e string
	)
	switch(in.Event){
	case "hello":
		cnt = "received"
		e = "hello back"
		break
	case "message":
		cnt = "received"
		e = "msg received"
		err := n.InsertOperation(in.Name, in.Timestamp, in.Content, in.Photo, in.Event, in.Img, in.Cid)
		if err != nil {
			return nil, err
		}
		fmt.Println(n.Todos)
		break
	case "release":
		fmt.Println(len(n.Todos))
		jsonMessage, _ := json.Marshal(&Message{Name: in.Cid, Content: in.Content, Photo: in.Photo, Event: "message", Img64: in.Img})
		n.Todos = n.Todos[1:]
		// fmt.Println("get Sender: ",string(jsonMessage))
		// TODO: process data= false
		manager.broadcast <- jsonMessage
		go n.checkAndRelease(in.Cid, in.Content, in.Photo, in.Event, in.Img, in.Name)
		break
	}

	return &chat.MessageReply{Name: n.Name, Content: cnt, Event: e, Timestamp: n.TimeStamp}, nil
}

func (manager *ClientManager) start() {
	for {
		select {
		//如果有新的连接接入,就通过channel把连接传递给conn
		case conn := <-manager.register:
			//把客户端的连接设置为true
			manager.clients[conn] = true
			//把返回连接成功的消息json格式化
			jsonMessage, _ := json.Marshal(&Message{Content: "new socket has connected."})
			//调用客户端的send方法，发送消息
			manager.send(jsonMessage, conn)
			//如果连接断开了
		case conn := <-manager.unregister:
			//判断连接的状态，如果是true,就关闭send，删除连接client的值
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "socket has disconnected."})
				manager.send(jsonMessage, conn)
			}
			//广播
		case message := <-manager.broadcast:
			//遍历已经连接的客户端，把消息发送给他们
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

//定义客户端管理的send方法
func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		//不给屏蔽的连接发送消息
		if conn != ignore {
			conn.send <- message
		}
	}
}

//定义客户端结构体的read方法
func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		//读取消息
		_, message, err := c.socket.ReadMessage()
		//如果有错误信息，就注销这个连接然后关闭
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		//如果没有错误信息就把信息放入broadcast
		data := make(map[string]string)
		err = json.Unmarshal(message, &data)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("get Data: ", data)
		nd.GreetAll(c.id, data)
		fmt.Println("id:", data["name"], ", ct: ",data["content"])
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		//从send里读消息
		case message, ok := <-c.send:
			//如果没有消息
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			//有消息就写入，发送给web端
			fmt.Println("send: ", string(message))
			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Start listening/service.
func (n *node) StartListening() {
	lis, err := net.Listen("tcp", n.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	_n := grpc.NewServer() // n is for serving purpose

	chat.RegisterChatServer(_n, n)
	// Register reflection service on gRPC server.
	reflection.Register(_n)

	// start listening
	if err := _n.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Register self with the service discovery module.
// This implementation simply uses the key-value store. One major drawback is that when nodes crash. nothing is updated on the key-value store. Services are a better fit and should be used eventually.
func (n *node) registerService() {
	config := api.DefaultConfig()
	config.Address = n.SDAddress
	consul, err := api.NewClient(config)
	if err != nil {
		log.Panicln("Unable to contact Service Discovery.")
	}

	kv := consul.KV()
	p := &api.KVPair{Key: n.Name, Value: []byte(n.Addr)}
	_, err = kv.Put(p, nil)
	if err != nil {
		log.Panicln("Unable to register with Service Discovery.")
	}

	// store the kv for future use
	n.SDKV = *kv

	log.Println("Successfully registered with Consul.")
}

func wsHandler(res http.ResponseWriter, req *http.Request) {
	//将http协议升级成websocket协议
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	// url, _ := json.Marshal(req.URL)
	if err != nil {
		http.NotFound(res, req)
		return
	}
	fmt.Println(req.RequestURI)
	guestId := strings.Split(req.RequestURI, "id=")
	client := &Client{id: guestId[1], socket: conn, send: make(chan []byte)}
	//注册一个新的链接
	manager.register <- client

	//启动协程收web端传过来的消息
	go client.read()
	//启动协程把消息返回给web端
	go client.write()
}

// Setup a new grpc client for contacting the server at addr.
func (n *node) SetupClient(node string, addr string, timestamp int64, content string, event string, cid string, photo string, img string) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	n.Clients[node] = chat.NewChatClient(conn)
	r, err := n.Clients[node].SendMessage(context.Background(), &chat.Message{Name: n.Name, Timestamp: timestamp, Content: content, Event: event, Cid: cid, Photo: photo, Img: img})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting from the other node: %s, content: %s, timeStamp: %d, event: %s", node, r.Content, r.Timestamp, r.Event)
	n.NodeMap[node] = r.Timestamp
	maxTime := math.Max(float64(n.TimeStamp), float64(r.Timestamp)) + 1
	n.TimeStamp = int64(maxTime)
	n.NodeMap[n.Name] = n.TimeStamp
	fmt.Println(n.NodeMap)
}


// Busy Work module, greet every new member you find
func (n *node) StartGreet() {
	kvpairs, _, err := n.SDKV.List(n.NodeName, nil)
	if err != nil {
		log.Panicln(err)
		return
	}
	fmt.Println(n.Name + " starts")
	timetmp := n.TimeStamp
	for _, kventry := range kvpairs {
		if strings.Compare(kventry.Key, n.Name) != 0 {
			fmt.Println("send to: ", kventry.Key)
			n.SetupClient(kventry.Key, string(kventry.Value), timetmp, "", "hello", "","","")
		}
	}
}

// read the input operation, if the list is empty, then write to the front,
// else write to the front of the operation which has a larger timestamp,
// if this operation does not exist then write to the back.
func (n *node) InsertOperation(name string, timeStamp int64, content string, photo string, event string, img string, userName string) error {
	node := strings.Split(name, " ")
	number, err := strconv.Atoi(node[1])
	if err != nil {
		return err
	}
	if len(n.Todos) == 0 {
		n.Todos = make([]data, 1)
		n.Todos[0] = data{Name: name, TimeStamp: timeStamp, Content: content, Photo: photo, Event: event, Img64: img, Number: number, UserName: userName}
	} else {
		for index, todo := range n.Todos {
			if todo.TimeStamp >= timeStamp && todo.Number > number {
				n.Todos = append(n.Todos, data{})
				copy(n.Todos[index+1:], n.Todos[index:])
				n.Todos[index] = data{Name: name, TimeStamp: timeStamp, Content: content, Photo: photo, Event: event, Img64: img, Number: number, UserName: userName}
				break
			} else if index+1 == len(n.Todos) {
				n.Todos = append(n.Todos, data{Name: name, TimeStamp: timeStamp, Content: content, Photo: photo, Event: event, Img64: img, Number: number, UserName: userName})
			}
		}
	}
	return nil
}

// Busy Work module, greet every new member you find
func (n *node) GreetAll(cid string, data map[string]string) {

	kvpairs, _, err := n.SDKV.List(n.NodeName, nil)
	if err != nil {
		log.Panicln(err)
		return
	}
	timetmp := n.TimeStamp + 1
	// fmt.Println("cid: ",cid,", user: ", data["name"])
	if strings.Compare(data["event"],"release") != 0{
		err = n.InsertOperation(n.Name, timetmp, data["content"], data["photo"], data["event"], data["img_64"], data["name"])
		if err != nil {
			log.Panicln(err)
			return
		}
	}
	for _, kventry := range kvpairs {
		if strings.Compare(kventry.Key, n.Name) != 0 {
			fmt.Println("send to: ", kventry.Key)
			n.SetupClient(kventry.Key, string(kventry.Value), timetmp, data["content"], data["event"], cid, data["photo"], data["img_64"])
		}
	}
	fmt.Println(n.Todos)
	n.checkAndRelease(cid,data["content"], data["photo"], data["event"], data["img"], data["name"])
	
}

func (n *node) checkAndRelease(cid string, content string, photo string, event string, img string, userName string){

	kvpairs, _, err := n.SDKV.List(n.NodeName, nil)
	if err != nil {
		log.Panicln(err)
		return
	}
	if len(n.Todos) > 0 {
		var (
			isFirst bool = false
			isLarger bool = true
		)
		// if current node is the first element in the todo list
		if strings.Compare(n.Todos[0].Name, n.Name) == 0 {
			isFirst = true
		}
		for _, kventry := range kvpairs {
			if n.NodeMap[kventry.Key] <= n.Todos[0].TimeStamp {
				isLarger = false
			}
		}
		if isFirst == true && isLarger == true {
			data := make(map[string]string)
			data["event"] = "release"
			data["name"] = userName
			data["content"] = content
			data["photo"] = photo
			data["img_64"] = img
			jsonMessage, _ := json.Marshal(&Message{Name: n.Todos[0].UserName, Content: n.Todos[0].Content, Photo: n.Todos[0].Photo, Event: n.Todos[0].Event, Img64: n.Todos[0].Img64})
			fmt.Println("get: ", string(jsonMessage))
			// time.Sleep(1 * time.Second)
			// // TODO: process data
			manager.broadcast <- jsonMessage
			n.Todos = n.Todos[1:]
			time.Sleep(time.Second / 10)
			n.GreetAll(cid, data)
		}
	}
}


func main() {
	args := os.Args[1:]
	// example: go run server/server.go "Node 1" :5000 localhost:8500
	// example: go run server/server.go "Node 2" :5001 localhost:8500
	if len(args) < 3 {
		fmt.Println("Arguments required: <name> <listening address> <consul address>")
		os.Exit(1)
	}

	// args in order
	name := args[0]
	listenAddr := args[1]
	consulAddr := args[2]

	nameSplit := strings.Split(name, " ")
	nodeNumber, err := strconv.Atoi(nameSplit[1])
	nodeName := nameSplit[0]
	if err != nil {
		log.Fatal("input argument <name> should have string with an integer, for example: Gnode 1, or node 2, or xxx 3")
	}

	nd = node{Name: name, Addr: listenAddr, SDAddress: consulAddr, NodeName: nodeName, NodeNumber: nodeNumber} // noden is for opeartional purposes

	nd.Clients = make(map[string]chat.ChatClient)
	nd.NodeMap = make(map[string]int64)

	// initialize timeStamp
	nd.TimeStamp = 0
	nd.NodeMap[name] = int64(0)

	fmt.Println("Initializing...")
	go nd.StartListening()
	nd.registerService()

	//开一个goroutine执行开始程序
	time.Sleep(1 * time.Second)
	nd.StartGreet()
	go manager.start()
	fmt.Println("Starting application...")
	//注册默认路由为 /ws ，并使用wsHandler这个方法
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(nd.Addr+"0", nil)
}