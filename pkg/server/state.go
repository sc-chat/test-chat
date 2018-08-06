package server

import (
	"sync"

	"github.com/sc-chat/test-chat/pkg/chat"
)

// ClientProcessor provide methods to work with clients
type ClientProcessor interface {
	Add(name, token string) bool
	Remove(token string) (string, bool)
	GetNameByToken(token string) (string, bool)
	AddStream(token string) chan chat.ResponseStream
	CloseStream(token string)
	Broadcast(s chat.ResponseStream)
}

// ClientsState implements ClientProcessor interface
type ClientsState struct {
	Tokens  map[string]string
	Names   map[string]map[string]bool
	Streams map[string]chan chat.ResponseStream

	tokenMtx  sync.RWMutex
	nameMtx   sync.RWMutex
	streamMtx sync.RWMutex
}

// Add method add new client with name and token to maps
func (c *ClientsState) Add(name, token string) bool {
	c.addToken(name, token)

	return !c.appendClientToken(name, token)
}

// GetNameByToken method returns client name
func (c *ClientsState) GetNameByToken(token string) (string, bool) {
	c.tokenMtx.RLock()
	name, ok := c.Tokens[token]
	c.tokenMtx.RUnlock()

	return name, ok
}

// Remove method removes client by token
// returns client name and bool flag which is true if last client token was deleted
func (c *ClientsState) Remove(token string) (string, bool) {
	name, ok := c.removeToken(token)

	// client not found
	if !ok {
		return "", false
	}

	return name, c.removeClientToken(name, token)
}

// Broadcast method sends event to all connected clients
func (c *ClientsState) Broadcast(s chat.ResponseStream) {
	c.streamMtx.RLock()

	for _, stream := range c.Streams {
		select {
		case stream <- s:
			// nothing to do
		default:
			// client stream is full, dropping message
		}
	}

	c.streamMtx.RUnlock()
}

// AddStream method adds new stream to stream map
func (c *ClientsState) AddStream(token string) chan chat.ResponseStream {
	stream := make(chan chat.ResponseStream, 100)
	c.streamMtx.Lock()
	c.Streams[token] = stream
	c.streamMtx.Unlock()

	return stream
}

// CloseStream method close stream and remove it from stream map
func (c *ClientsState) CloseStream(token string) {
	c.streamMtx.Lock()
	stream, ok := c.Streams[token]
	if ok {
		delete(c.Streams, token)
		close(stream)
	}
	c.streamMtx.Unlock()
}

func (c *ClientsState) addToken(name, token string) {
	c.tokenMtx.Lock()
	c.Tokens[token] = name
	c.tokenMtx.Unlock()
}

func (c *ClientsState) removeToken(token string) (string, bool) {
	c.tokenMtx.Lock()
	name, ok := c.Tokens[token]
	if ok {
		delete(c.Tokens, token)
	}
	c.tokenMtx.Unlock()

	return name, ok
}

func (c *ClientsState) appendClientToken(name, token string) bool {
	c.nameMtx.Lock()
	_, ok := c.Names[name]
	if !ok {
		c.Names[name] = make(map[string]bool)
	}
	c.Names[name][token] = true
	c.nameMtx.Unlock()

	return ok
}

func (c *ClientsState) removeClientToken(name, token string) bool {
	c.nameMtx.Lock()
	defer c.nameMtx.Unlock()
	_, ok := c.Names[name]
	if !ok {
		return false
	}

	_, ok = c.Names[name][token]
	if !ok {
		return false
	}

	delete(c.Names[name], token)

	if len(c.Names[name]) > 0 {
		return false
	}

	delete(c.Names, name)
	return true
}

// NewClientState returns ClientsState pointer
func NewClientState() *ClientsState {
	return &ClientsState{
		Tokens:  make(map[string]string),
		Names:   make(map[string]map[string]bool),
		Streams: make(map[string]chan chat.ResponseStream),
	}
}
