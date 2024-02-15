package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Server struct {
	n             *maelstrom.Node
	nodeId        string
	messagesMutex sync.RWMutex
	messages      map[int]struct{}
}

func main() {
	n := maelstrom.NewNode()
	s := &Server{n: n, nodeId: n.ID(), messages: make(map[int]struct{})}

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var broadcast map[string]any
		if err := json.Unmarshal(msg.Body, &broadcast); err != nil {
			return err
		}
		message := int(broadcast["message"].(float64))
		s.messagesMutex.Lock()
		if _, exists := s.messages[message]; exists {
			s.messagesMutex.Unlock()
			return nil
		}
		s.messages[message] = struct{}{}
		s.messagesMutex.Unlock()

		if err := gossip(s, msg.Src, broadcast); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{
			"type": "broadcast_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var read map[string]any
		if err := json.Unmarshal(msg.Body, &read); err != nil {
			return err
		}
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": getAllIds(s),
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func getAllIds(s *Server) []int {
	s.messagesMutex.RLock()
	defer s.messagesMutex.RUnlock()
	ids := make([]int, 0, len(s.messages))
	for key := range s.messages {
		ids = append(ids, key)
	}
	return ids
}

func gossip(s *Server, src string, body map[string]any) error {
	for _, dst := range s.n.NodeIDs() {
		if dst == src || dst == s.nodeId {
			continue
		}
		dst := dst
		go func() {
			if err := s.n.Send(dst, body); err != nil {
				panic(err)
			}
		}()
	}
	return nil
}
