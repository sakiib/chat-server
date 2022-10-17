package server

import (
	"bufio"
	"fmt"
	"github.com/sakiib/chat-server/utils"
	"github.com/spf13/cast"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

type Server struct {
	clients   []*Client
	chatRooms map[string]*ChatRoom
	incoming  chan *Message
	join      chan *Client
}

func NewServer() *Server {
	s := &Server{
		clients:   make([]*Client, 0),
		chatRooms: make(map[string]*ChatRoom),
		incoming:  make(chan *Message),
		join:      make(chan *Client),
	}
	s.Listen()
	return s
}

func (s *Server) Listen() {
	go func() {
		for {
			select {
			case message := <-s.incoming:
				s.Parse(message)
			case client := <-s.join:
				s.Join(client)
			}
		}
	}()
}

func (s *Server) Join(client *Client) {
	s.clients = append(s.clients, client)
	client.outgoing <- "successfully joined the chat server\n"
	go func() {
		for message := range client.incoming {
			s.incoming <- message
		}
	}()
}

func (s *Server) Parse(message *Message) {
	switch {
	case strings.HasPrefix(message.text, utils.CMD_WhoAmI):
		s.FindIdentity(message)
	case strings.HasPrefix(message.text, utils.CMD_UserList):
		s.ListUsers(message)
	case strings.HasPrefix(message.text, utils.CMD_SendMessage):
		s.SendMessage(message)
	case strings.HasPrefix(message.text, utils.CMD_SendToUsers):
		s.SendToUsers(message)
	case strings.HasPrefix(message.text, utils.CMD_Help):
		s.Help(message.client)
	default:
		message.client.outgoing <- "unknown command\n"
	}
}

func (s *Server) SendToUsers(message *Message) {
	if message.client.chatRoom == nil {
		return
	}
	users := utils.ParseUsersList(message.text)
	log.Println("users: ", users)
	message.client.chatRoom.ToUsers(message, users)
}

func (s *Server) SendMessage(message *Message) {
	if message.client.chatRoom == nil {
		return
	}
	message.client.chatRoom.Broadcast(message.String())
}

func (s *Server) FindIdentity(message *Message) {
	if message.client.chatRoom == nil {
		return
	}

	message.client.chatRoom.ReturnIdentity(message)
}

func (s *Server) ListUsers(message *Message) {
	if message.client.chatRoom == nil {
		return
	}

	message.client.chatRoom.List(message)
}

func (s *Server) CreateChatRoom(client *Client, name string) {
	if s.chatRooms[name] != nil {
		return
	}
	chatRoom := NewChatRoom(name)
	s.chatRooms[name] = chatRoom
	client.outgoing <- fmt.Sprintf("%s chat room created\n", chatRoom.name)
	log.Printf("client %v created chat room\n", client.id)
}

func (s *Server) JoinChatRoom(client *Client, name string) {
	s.chatRooms[name].Join(client)
	log.Printf("client %v joined chat room\n", client.id)
}

func (s *Server) Help(client *Client) {
	client.outgoing <- "\nAvailable Commands:\n"
	client.outgoing <- "whoAmI - sends the client's self identity\n"
	client.outgoing <- "userList - lists the connected clients(ID)\n"
	client.outgoing <- "sendMessage - sends messages to all the clients\n"
	client.outgoing <- "sendToUsers - sends messages to the client with provided IDs\n"
	client.outgoing <- "help - lists all the available commands\n"
	client.outgoing <- "\n"
}

type ChatRoom struct {
	name     string
	clients  []*Client
	messages []string
}

func NewChatRoom(name string) *ChatRoom {
	return &ChatRoom{
		name:     name,
		clients:  make([]*Client, 0),
		messages: make([]string, 0),
	}
}

func (chatRoom *ChatRoom) Join(client *Client) {
	client.chatRoom = chatRoom
	for _, message := range chatRoom.messages {
		client.outgoing <- message
	}
	chatRoom.clients = append(chatRoom.clients, client)
	chatRoom.Broadcast(fmt.Sprintf("%v joined the chat\n", client.id))
}

func (chatRoom *ChatRoom) List(message *Message) {
	var users []string
	for _, client := range chatRoom.clients {
		users = append(users, cast.ToString(client.id))
	}

	for _, client := range chatRoom.clients {
		if client.id == message.client.id {
			client.outgoing <- strings.Join(users, ", ")
		}
	}
}

func (chatRoom *ChatRoom) ReturnIdentity(message *Message) {
	for _, client := range chatRoom.clients {
		if client.id == message.client.id {
			client.outgoing <- fmt.Sprintf("whoAmI? My ID %v\n", message.client.id)
		}
	}
}

func (chatRoom *ChatRoom) Broadcast(message string) {
	chatRoom.messages = append(chatRoom.messages, message)
	for _, client := range chatRoom.clients {
		client.outgoing <- message
	}
}

func (chatRoom *ChatRoom) ToUsers(message *Message, users []string) {
	for _, user := range users {
		for _, client := range chatRoom.clients {
			if cast.ToString(client.id) == user {
				chatRoom.messages = append(chatRoom.messages, message.String())
				client.outgoing <- message.String()
			}
		}
	}
}

type Client struct {
	id       int64
	chatRoom *ChatRoom
	incoming chan *Message
	outgoing chan string
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func NewClient(conn net.Conn) *Client {
	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	client := &Client{
		id:       rand.Int63n(100),
		chatRoom: nil,
		incoming: make(chan *Message),
		outgoing: make(chan string),
		conn:     conn,
		reader:   reader,
		writer:   writer,
	}

	client.Listen()
	return client
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func (client *Client) Read() {
	for {
		str, err := client.reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		message := NewMessage(time.Now(), client, strings.TrimSuffix(str, "\n"))
		client.incoming <- message
	}
	close(client.incoming)
}

func (client *Client) Write() {
	for str := range client.outgoing {
		_, err := client.writer.WriteString(str)
		if err != nil {
			log.Println(err)
			break
		}
		err = client.writer.Flush()
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func (client *Client) Quit() {
	client.conn.Close()
}

type Message struct {
	time   time.Time
	client *Client
	text   string
}

func NewMessage(time time.Time, client *Client, text string) *Message {
	return &Message{
		time:   time,
		client: client,
		text:   text,
	}
}

func (message *Message) String() string {
	if strings.Index(message.text, " ") != -1 {
		message.text = message.text[strings.Index(message.text, " "):]
	}

	if strings.Index(message.text, "[") != -1 {
		message.text = message.text[:strings.Index(message.text, "[")]
	}

	return fmt.Sprintf("%d: %s\n", message.client.id, message.text)
}

func StartChatServer() {
	Server := NewServer()

	listener, err := net.Listen(utils.TYPE, utils.PORT)
	if err != nil {
		log.Println("Error: ", err)
		os.Exit(1)
	}

	defer listener.Close()
	log.Println("chat server running on PORT " + utils.PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error: ", err)
			continue
		}
		cl := NewClient(conn)
		Server.CreateChatRoom(cl, utils.DEFAULT_ROOM)
		Server.JoinChatRoom(cl, utils.DEFAULT_ROOM)
		Server.Join(cl)
	}
}
