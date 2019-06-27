// 封装API，隐藏细节
// 封装Connection结构，隐藏websocket底层连接
// 封装Connection的API，提供ReadMessage/WriteMessage/Close等线程安全的接口
package impl

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Connection struct {
	wsConn    *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan []byte
	isClosed  bool
	mutex     sync.Mutex
}

func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan []byte, 1),
	}

	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()

	return
}

// 读取数据
func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("ReadMessage connection is closed")
	}

	return
}

// 发送数据
func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("WriteMessage connection is closed")
	}

	return
}

// 关闭连接
func (conn *Connection) Close() {
	conn.wsConn.Close()

	conn.mutex.Lock()
	if !conn.isClosed {
		// 只能关闭一次
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()
}

// 不断的从websocket连接上读取数据
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}

	}
ERR:
	log.Println("读数据协程关闭")
	conn.Close()
}

// 不断的往websocket连接上写数据
func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)
	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}

		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
ERR:
	log.Println("写数据协程关闭")
	conn.Close()
}
