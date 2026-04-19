package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Client struct{
	id string
	channel chan[]byte
}

type SSEHub struct{
	mu sync.RWMutex
	clients map[string]*Client
	register chan *Client
	unregister chan *Client
	broadcast chan []byte

}


func NewSSEHub()*SSEHub{
	return &SSEHub{
		clients: make(map[string]*Client),
		register: make(chan *Client ),
		unregister: make(chan *Client),
		broadcast: make(chan []byte,256),
	}
}

func (h *SSEHub) Run(){
	for{
		select {
		case client:= <- h.register:
			h.mu.Lock()
			h.clients[client.id]= client
			h.mu.Unlock()
			log.Printf("[SSE] client connected: %s total=%d\n",client.id,len(h.clients))
		case client := <- h.unregister:
			h.mu.Lock()
			if _,ok:= h.clients[client.id];ok{
				delete(h.clients,client.id)
				close(client.channel)
			}
			h.mu.Unlock()
			log.Printf("[SSE] client disconnected: %s total=%d\n",client.id,len(h.clients))
		case message := <-h.broadcast:
			h.mu.RLock()
			for _,client:= range h.clients{
				select{
				case client.channel <- message:
				default:
					log.Printf("[SSE] client too slow\n")
				}
			}
			h.mu.RUnlock()
			
		}
	}
}

func (h *SSEHub) BroadcastRaw(msg []byte){
	select{
	case h.broadcast <-msg:
	default:
		log.Println("broadcast channel full,dropping")
	}
}

//Event types

type SSEEvent struct{
	Type string `json:"type"`
	Data any  `json:"data"`
}

func (h *SSEHub) Publish(eventType string,data any){
	event:=SSEEvent{Type: eventType,Data: data}
	b,err:= json.Marshal(event)
	if err !=nil{
		log.Printf("SSE marshal erros: %v\n",err)
		return
	}
	select {
	case h.broadcast <-b:
	default:
		log.Printf("SSE broadcast channel full, dropping event")
	}
}

//Stream
func (h *SSEHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "stream not supported", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    flusher.Flush()

    client := &Client{
        id:      r.RemoteAddr,
        channel: make(chan []byte, 64),
    }
    h.register <- client
    defer func() {
        h.unregister <- client
    }()
    fmt.Fprintf(w, "data: {\"type\":\"connected\"}\n\n")
    flusher.Flush()

    for {
        select {
        case msg, ok := <-client.channel:
            if !ok {
                return
            }
            fmt.Fprintf(w, "data: %s\n\n", msg)
            flusher.Flush()
        case <-r.Context().Done():
            return
        }
    }
}