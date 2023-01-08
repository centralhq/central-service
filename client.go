// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"bytes"
	"log"
	"encoding/json"
	"time"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	handler *OperationManager
	
	uuid string
}

func NewClient(hub *Hub, conn *websocket.Conn, handler *OperationManager, uuid string) *Client {
	return &Client{
		hub: hub,
		conn: conn,
		send: make(chan []byte),
		handler: handler,
		uuid: uuid,
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// TODO: insert setter logic here, but need to get the counter from the postgres
		var operation *CentralObject
 
		err = json.Unmarshal(message, &operation)

		if err != nil {
			log.Println(err)
			return
		}
			
		updatedOperation, err := c.updateObject(operation)

		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		bytes := c.toBytes(updatedOperation)
		
		c.hub.broadcast <- bytes
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) toBytes(op *CentralObject) []byte {

	result := AckOperation{
		Status:     		"success",
		OperationType: 		op.OperationType,
		CentralUid:       	op.CentralUid,
		ObjectId:			op.ObjectId,
		MutationCounter:	op.MutationCounter,
		ConflictId: 		op.ConflictId,
		Payload:    		op.Payload,
	}
	bytes, err := json.Marshal(result)

	if err != nil {
		log.Println(err)
	}

	return bytes
}

func (c *Client) updateObject(op *CentralObject) (*CentralObject, error) {
	op.CentralUid = c.uuid
	newCounter, err := c.handler.executeSync(op.ObjectId)
	op.MutationCounter = newCounter

	if err != nil {
		log.Println(err)
		return nil, err
	} 

	return op, nil
}