package module

import (
	"fmt"
    "github.com/gorilla/websocket"
    "github.com/novalagung/gubrak/v2"
    "io/ioutil"
    "log"
    "net/http"
    "strings"
)

func NewChatRoom() *typestruct.ChatRoom {
	return &typestruct.ChatRoom{
		Clients:    make([]*typestruct.Client, 0),
		Register:   make(chan *typestruct.Client),
		Unregister: make(chan *typestruct.Client),
		Broadcast:  make(chan typestruct.Message),
	}
}
func BroadcastMessage(message typestruct.Message) {
	NewChatRoom().Broadcast <- message
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        content, err := ioutil.ReadFile("index.html")
        if err != nil {
            http.Error(w, "Could not open requested file", http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "%s", content)
    })

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        currentGorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
    if err != nil {
        http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
    }

    username := r.URL.Query().Get("username")
    currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username}
    connections = append(connections, &currentConn)

	app.ws = new WebSocket("ws://127.0.0.1:3080/ws?username=" + name)

    go handleIO(&currentConn, connections)
    })

    fmt.Println("Server starting at :3000")
    http.ListenAndServe(":3000", nil)
}

func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
    defer func() {
        if r := recover(); r != nil {
            log.Println("ERROR", fmt.Sprintf("%v", r))
        }
    }()

    broadcastMessage(currentConn, MESSAGE_NEW_USER, "")

    for {
        payload := SocketPayload{}
        err := currentConn.ReadJSON(&payload)
        if err != nil {
            if strings.Contains(err.Error(), "websocket: close") {
                broadcastMessage(currentConn, MESSAGE_LEAVE, "")
                ejectConnection(currentConn)
                return
            }

            log.Println("ERROR", err.Error())
            continue
        }

        broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
    }
}

func ejectConnection(currentConn *WebSocketConnection) {
	filtered := gubrak.From(connections).Reject(func(each *WebSocketConnection) bool {
		return each == currentConn
	}).Result()
	connections = filtered.([]*WebSocketConnection)
}

func broadcastMessage(currentConn *WebSocketConnection, kind, message string) {
	for _, eachConn := range connections {
		if eachConn == currentConn {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
		})
	}
}