package app

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type ChatServer struct {
	mutex   *sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func (cs *ChatServer) Join(conn *websocket.Conn) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cs.clients[conn] = struct{}{}
}

func (cs *ChatServer) Leave(conn *websocket.Conn) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	delete(cs.clients, conn)
}

func (cs *ChatServer) Boardcast(msg []byte) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	for client := range cs.clients {
		/*
			limitation ของ gorilla ตอน write จะไม่สามารถเขียนแบบ concurrent ได้
		*/
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func WebSocketHandler() http.HandlerFunc {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	chat := &ChatServer{
		mutex:   &sync.Mutex{},
		clients: make(map[*websocket.Conn]struct{}),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		/*
			upgrade คือ ส่วนที่ช่วยแปลงตัว http ที่เป็นการสือสารแบบ directional ให้เป็นแบบ bi-directional
		*/
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		chat.Join(conn)
		defer chat.Leave(conn)

		for {
			/*
				error ที่สามารถเกิดขึ้นได้เช่น browser ทำการปิด tab (close connection) ขณะที่ server กำลังอ่าน message
			*/
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			log.Println(msg)
			chat.Boardcast(msg)
		}
	}
}
