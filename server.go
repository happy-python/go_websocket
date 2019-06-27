package main

import (
	"github.com/gorilla/websocket"
	"go_websocket/impl"
	"log"
	"net/http"
	"time"
)

func handle(w http.ResponseWriter, r *http.Request) {
	var (
		upgrader websocket.Upgrader
		wsConn   *websocket.Conn
		err      error
		data     []byte
		conn     *impl.Connection
	)

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}

	if conn, err = impl.InitConnection(wsConn); err != nil {
		return
	}

	// 每隔1秒向客户端发送一条心跳
	go func() {
		for {
			time.Sleep(1 * time.Second)
			if err := conn.WriteMessage([]byte("heartbeat from server")); err != nil {
				log.Println(err.Error())
				return
			}
		}
	}()

	// 读到什么，返回什么
	for {
		if data, err = conn.ReadMessage(); err != nil {
			log.Println(err.Error())
			goto ERR
		}
		if err = conn.WriteMessage(data); err != nil {
			log.Println(err.Error())
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func main() {
	http.HandleFunc("/", handle)
	if err := http.ListenAndServe(":8888", nil); err != nil {
		log.Fatal(err.Error())
	}
}
