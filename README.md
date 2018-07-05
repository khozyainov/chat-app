# chat-app
Simple chat

To start  `go run main.go`

Then open localhost:8888


- Код ответа сервера в случае обработанного исключения будет 200, а тело ответа будет представлено json-документом по схеме: {"status":401,"result":"Unauthorized"}
- Код ответа в случае необработанного исключения может быть другим: 500, 501, 502, 403, 404 и т.д.
- В случае успешного выполнения запроса, ответ будет представлен json-документом по схеме: {"status":200,"result":*}, где * будет отличаться от запроса к запросу.


[Регистрация](#sign)
[Авторизация](#login)
[Смена имени](#changeName)
[Отправить сообщение](#sendMessage)
[Получить все сообщения](#getAllMessages)


### Регистрация
<a name="sign">[POST] /api/sign</a>
--------
Регистрация и получение авторизационного токена

Request:
- login: Логин
- password: Пароль
- name: Имя

Ответ:
```javascript
{"status":200,"result":{"token":"eyJhbGciOiJW0xZqnYokU09Mc54I"}}
```
>

### Авторизация
<a name="login">[POST] /api/login</a>
--------
Авторизация и получение авторизационного токена

Request:
- login: Логин
- password: Пароль

Ответ:
```javascript
{"status":200,"result":{"token":"eyJhbGciOiJW0xZqnYokU09Mc54I"}}
```
>

### Смена ника
<a name="changeName">[POST] /api/changeName</a>
--------
Смена ника

Request:
- token: авторизационный токен
- name: Новое имя

Ответ:
```javascript
{"status":200,"result":"OK"}}
```
>

### Отправить сообщение
<a name="sendMessage">[POST] /api/sendMessage</a>
--------
Отправка сообщения

Request:
- token: авторизационный токен
- body: сообщение

Ответ:
```javascript
{"status":200,"result":"OK"}}
```
>

### Получить все сообщения
<a name="getAllMessages">[POST] /api/getAllMessages</a>
--------
Получение всех сообщений

Request:
- token: авторизационный токен

Ответ:
```javascript
{"status":200,"result":{"Messages":[{"userName":"Tony","body":"Hello world"}]}}
```
>
