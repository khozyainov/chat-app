package chat

import (
	"encoding/json"
	"log"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

type Token struct {
	Token string `json:"token"`
}

type Usr struct {
	Login    string
	Password string
	Name     string
}

type Msg struct {
	Token string `json:"token"`
	Body  string `json:"body"`
}

type Chg struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type Result struct {
	Status int         `json:"status"`
	Result interface{} `json:"result,omitempty"`
}

func (server *Server) handleApiSign(responseWriter http.ResponseWriter, request *http.Request) {

	if request.Method != "POST" {
		r := Result{405, "Method Not Allowed"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}

	log.Println("Handling create new account")
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	decoder := json.NewDecoder(request.Body)
	var u Usr
	err := decoder.Decode(&u)
	if err != nil {
		log.Println(err)
	}
	log.Println(u)

	login := u.Login
	password := u.Password
	name := u.Name

	if login == "" || password == "" || name == "" {
		r := Result{400, "Bad Request"}
		json.NewEncoder(responseWriter).Encode(r)
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
	err = db.QueryRow("select count(*) from Persons where login = ?", login).Scan(&loginExist)

	if err != nil {
		log.Println(err)
	}

	if loginExist != 0 {
		r := Result{400, "Bad Request"}
		json.NewEncoder(responseWriter).Encode(r)
		log.Println("User exists")
		return
	}

	res, err := db.Exec("insert into Persons (login, password, name, token) values(?, ?, ?, ?);", login, password, name, tokenString)

	if err != nil {
		log.Println(err)
	}

	id, err := res.LastInsertId()

	if err != nil {
		log.Println(err)
		r := Result{500, "Internal Server Error"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}
	log.Println("ID", id)
	t := Token{tokenString}
	r := Result{200, t}
	json.NewEncoder(responseWriter).Encode(r)
}

func (server *Server) handleChangeName(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		r := Result{405, "Method Not Allowed"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}

	decoder := json.NewDecoder(request.Body)
	var c Chg
	err := decoder.Decode(&c)
	if err != nil {
		log.Println(err)
	}

	name := c.Name
	token := c.Token

	if _, ok := server.usersTokens[token]; !ok {
		log.Println("Unauthorized")
		r := Result{401, "Unauthorized"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}

	change, err := db.Prepare("UPDATE Persons SET name=? WHERE token=?")
	if err != nil {
		panic(err.Error())
	}
	change.Exec(name, token)
	log.Println("Change name:", server.usersTokens[token], "to", name)
	server.usersTokens[token] = name
	r := Result{200, "OK"}
	json.NewEncoder(responseWriter).Encode(r)
}

func (server *Server) handleSendMessage(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		r := Result{405, "Method Not Allowed"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}
	log.Println("Handling send message")

	decoder := json.NewDecoder(request.Body)
	var m Msg
	err := decoder.Decode(&m)
	if err != nil {
		log.Println(err)
	}

	if _, ok := server.usersTokens[m.Token]; !ok {
		log.Println("Unauthorized")
		r := Result{401, "Unauthorized"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}

	message := Message{server.usersTokens[m.Token], m.Body}

	log.Println(message.UserName)

	if len(server.Messages) == cap(server.Messages) {
		server.Messages = append(server.Messages[:0], server.Messages[1:]...)
	}
	server.Messages = append(server.Messages, &message)
	server.sendAll(&message)
	r := Result{200, "OK"}
	json.NewEncoder(responseWriter).Encode(r)
}

func (server *Server) handleGetAllMessages(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		r := Result{405, "Method Not Allowed"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}

	decoder := json.NewDecoder(request.Body)
	var t Token
	err := decoder.Decode(&t)
	if err != nil {
		log.Println(err)
	}

	if _, ok := server.usersTokens[t.Token]; ok {
		r := Result{200, server}
		json.NewEncoder(responseWriter).Encode(r)
	} else {
		r := Result{401, "Unauthorized"}
		json.NewEncoder(responseWriter).Encode(r)
	}
}

func (server *Server) handleApiLogin(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		r := Result{405, "Method Not Allowed"}
		json.NewEncoder(responseWriter).Encode(r)
		return
	}

	decoder := json.NewDecoder(request.Body)
	var u Usr
	err := decoder.Decode(&u)
	if err != nil {
		log.Println(err)
	}
	log.Println(u)

	login := u.Login
	password := u.Password
	var t Token

	err = db.QueryRow("select name, token from Persons where login = ? and password = ?", login, password).Scan(&u.Name, &t.Token)

	name := u.Name

	if err != nil {
		log.Println(err)
	}

	if t.Token == "" {
		http.Error(responseWriter, http.StatusText(401), 401)
		log.Println("Wrong login or password")
		return
	}

	server.usersTokens[t.Token] = name

	log.Println("Success auth:", name, t.Token)

	r := Result{200, t}

	json.NewEncoder(responseWriter).Encode(r)
}
