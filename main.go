package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const (
	GlobalKey = "counter::global"
	AddLock   = "lock::add"
)

const (
	Unlock = 0
	Lock   = 1
)

type AddRequest struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
}

type Server struct {
	node *maelstrom.Node
	kv   *maelstrom.KV
	mu   sync.RWMutex
}

func NewServer() *Server {
	node := maelstrom.NewNode()
	return &Server{
		node: node,
		kv:   maelstrom.NewSeqKV(node),
	}
}

func (s *Server) Serve() {
	s.node.Handle("add", s.addCommand)
	s.node.Handle("read", s.readCommand)
	s.node.Handle("echo", s.echoCommand)
	if err := s.node.Run(); err != nil {
		panic(err)
	}
}

func (s *Server) addCommand(msg maelstrom.Message) error {
	var body AddRequest
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()

	s.DistributedLock(ctx)
	current, err := s.kv.ReadInt(context.Background(), GlobalKey)
	if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
		fmt.Fprintf(os.Stderr, "error reading global counter: %s\n", err)
		return err
	}

	current += body.Delta
	if err := s.kv.Write(context.Background(), GlobalKey, current); err != nil {
		fmt.Fprintf(os.Stderr, "error writing global counter: %s\n", err)
		return err
	}
	s.DistributedUnlock(ctx)
	return s.node.Reply(msg, map[string]any{
		"type": "add_ok",
	})
}

func (s *Server) readCommand(msg maelstrom.Message) error {
	ctx := context.Background()
	s.DistributedLock(ctx)
	current, err := s.kv.ReadInt(context.Background(), GlobalKey)
	if err != nil && maelstrom.ErrorCode(err) != maelstrom.KeyDoesNotExist {
		fmt.Fprintf(os.Stderr, "error reading global counter: %s\n", err)
		return err
	}
	s.node.Reply(msg, map[string]any{
		"type":  "read_ok",
		"value": current,
	})
	s.DistributedUnlock(ctx)
	return nil
}

func (s *Server) echoCommand(msg maelstrom.Message) error {
	return s.node.Reply(msg, map[string]any{
		"type": "echo_ok",
	})
}

func (s *Server) DistributedLock(ctx context.Context) {
	backOff := time.Duration(50) * time.Millisecond
	for {
		if err := s.kv.CompareAndSwap(ctx, AddLock, Unlock, Lock, true); err == nil {
			break
		}
		time.Sleep(backOff)
	}
}

func (s *Server) DistributedUnlock(ctx context.Context) {
	backOff := time.Duration(50) * time.Millisecond
	for {
		if err := s.kv.CompareAndSwap(ctx, AddLock, Lock, Unlock, true); err == nil {
			break
		}
		time.Sleep(backOff)
	}
}

func main() {
	server := NewServer()

	server.Serve()
}
