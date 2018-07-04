package chat

import (
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var db, err = sql.Open("mysql", "root:Ww19082001@/chat")

var mySigningKey = []byte("secret")

type Account struct {
	login    string
	password string
}

type Server struct {
	connectedUsers     map[int]*User
	usersTokens        map[string]string
	Messages           []*Message `json: messages`
	addUser            chan *User
	removeUser         chan *User
	newIncomingMessage chan *Message
	errorChannel       chan error
	doneCh             chan bool
}

func NewServer() *Server {
	Messages := make([]*Message, 0, 5)
	connectedUsers := make(map[int]*User)
	UsersTokens := make(map[string]string)
	addUser := make(chan *User)
	removeUser := make(chan *User)
	newIncomingMessage := make(chan *Message)
	errorChannel := make(chan error)
	doneCh := make(chan bool)

	return &Server{
		connectedUsers,
		UsersTokens,
		Messages,
		addUser,
		removeUser,
		newIncomingMessage,
		errorChannel,
		doneCh,
	}
}

func (server *Server) AddUser(user *User) {
	log.Println("In AddUser")
	server.addUser <- user
}

func (server *Server) RemoveUser(user *User) {
	log.Println("Removing user")
	server.removeUser <- user
}

func (server *Server) ProcessNewIncomingMessage(message *Message) {
	server.newIncomingMessage <- message
}

func (server *Server) Done() {
	server.doneCh <- true
}

func (server *Server) sendPastMessages(user *User) {
	for _, msg := range server.Messages {
		user.Write(msg)
	}
}

func (server *Server) Err(err error) {
	server.errorChannel <- err
}

func (server *Server) sendAll(msg *Message) {
	log.Println("In Sending to all Connected users")
	for _, user := range server.connectedUsers {
		user.Write(msg)
	}
}

func (server *Server) Listen() {

	if err != nil {
		log.Println("Database error: ", err)
	}
	defer db.Close()

	log.Println("Server Listening .....")

	http.HandleFunc("/chat", server.handleChat)

	http.HandleFunc("/login", server.handleLogin)
	http.HandleFunc("/sign", server.handleSign)

	http.HandleFunc("/api/login", server.handleApiLogin)
	http.HandleFunc("/api/sign", server.handleApiSign)
	http.HandleFunc("/api/getAllMessages", server.handleGetAllMessages)
	http.HandleFunc("/api/changeName", server.handleChangeName)
	http.HandleFunc("/api/sendMessage", server.handleSendMessage)

	for {
		select {
		case user := <-server.addUser:
			log.Println("Added a new User")
			server.connectedUsers[user.id] = user
			log.Println("Now ", len(server.connectedUsers), " users are connected to chat room")
			server.sendPastMessages(user)
		case user := <-server.removeUser:
			log.Println("Removing user from chat room")
			delete(server.connectedUsers, user.id)
		case msg := <-server.newIncomingMessage:
			if len(server.Messages) == cap(server.Messages) {
				server.Messages = append(server.Messages[:0], server.Messages[1:]...)
			}
			log.Println("!!!!!!!" + server.usersTokens[msg.UserName])
			msg.UserName = server.usersTokens[msg.UserName]
			server.Messages = append(server.Messages, msg)
			server.sendAll(msg)
		case err := <-server.errorChannel:
			log.Println("Error : ", err)
		case <-server.doneCh:
			return
		}
	}
}

func (server *Server) handleSign(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == "GET" {
		http.ServeFile(responseWriter, request, "sign.html")
	} else if request.Method != "POST" {
		http.Error(responseWriter, http.StatusText(405), 405)
		return
	}

	log.Println("Handling create new account")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	login := request.FormValue("login")
	password := request.FormValue("password")
	name := request.FormValue("name")

	log.Println(name, password, login)

	if login == "" || password == "" || name == "" {
		http.Error(responseWriter, http.StatusText(400), 400)
		return
	}

	claims["login"] = login
	claims["password"] = password
	claims["name"] = name

	log.Println("Sign in with login: ", claims["login"], ", name: ", claims["name"])
	tokenString, _ := token.SignedString(mySigningKey)
	server.usersTokens[tokenString] = name

	log.Println(login, password, name, tokenString)

	var loginExist int
	err := db.QueryRow("select count(*) from Persons where login = ?", login).Scan(&loginExist)

	if err != nil {
		log.Println(err)
	}

	if loginExist != 0 {
		http.Error(responseWriter, http.StatusText(400), 400)
		log.Println("Login exists")
		return
	}

	res, err := db.Exec("insert into Persons (login, password, name, token) values(?, ?, ?, ?);", login, password, name, tokenString)

	if err != nil {
		log.Println(err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		log.Println(err)
		http.Error(responseWriter, http.StatusText(500), 500)
		return
	}
	log.Println(id)
	cookie := http.Cookie{Name: "token", Value: tokenString, MaxAge: 1440}
	http.SetCookie(responseWriter, &cookie)
	http.ServeFile(responseWriter, request, "chat.html")
	http.Redirect(responseWriter, request, "/chat", 200)
}

func (server *Server) handleLogin(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		http.Error(responseWriter, http.StatusText(405), 405)
		return
	}
	var acc Account

	acc.login = request.FormValue("login")
	acc.password = request.FormValue("password")

	var (
		token string
		name  string
	)
	err = db.QueryRow("select name, token from Persons where login = ? and password = ?", acc.login, acc.password).Scan(&name, &token)

	if err != nil {
		log.Println(err)
	}

	if token == "" {
		http.Error(responseWriter, http.StatusText(401), 401)
		log.Println("Wrong login or password")
		return
	}

	server.usersTokens[token] = name

	log.Println("Success auth:", name, token)

	cookie := http.Cookie{Name: "token", Value: token, MaxAge: 1440}
	http.SetCookie(responseWriter, &cookie)
	http.ServeFile(responseWriter, request, "chat.html")
	http.Redirect(responseWriter, request, "/chat", 200)
}

func (server *Server) handleChat(responseWriter http.ResponseWriter, request *http.Request) {
	log.Println("Handling chat request ")
	var messageObject Message
	conn, _ := upgrader.Upgrade(responseWriter, request, nil)
	err := conn.ReadJSON(&messageObject)
	log.Println("Message retireved when add user recieved: ", &messageObject)

	if err != nil {
		log.Println("Error while reading JSON from websocket ", err.Error())
	}
	user := NewUser(conn, server)

	log.Println("Going to add user")
	server.AddUser(user)

	log.Println("User added successfully")
	server.ProcessNewIncomingMessage(&messageObject)
	user.Listen()
}
