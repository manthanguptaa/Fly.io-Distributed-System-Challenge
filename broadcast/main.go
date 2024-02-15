package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Broadcast struct {
	Type    string `json:"type"`
	Message any    `json:"message"`
}

type BroadcastResponse struct {
	Type string `json:"type"`
}

type Read struct {
	Type string `json:"type"`
}

type ReadResponse struct {
	Type     string `json:"type"`
	Messages []any  `json:"messages"`
}

type Topology struct {
	Type     string              `json:"type"`
	MsgID    int                 `json:"msg_id"`
	Topology map[string][]string `json:"topology"`
}

type TopologyResponse struct {
	Type string `json:"type"`
}

func main() {
	n := maelstrom.NewNode()

	var messageList = make([]any, 0)

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var broadcast Broadcast
		if err := json.Unmarshal(msg.Body, &broadcast); err != nil {
			return err
		}
		messageList = append(messageList, broadcast.Message)
		response := BroadcastResponse{Type: "broadcast_ok"}
		nodeIds := n.NodeIDs()
		for _, node := range nodeIds {
			if node == n.ID() {
				continue
			}
			n.Send(node, broadcast)
		}
		return n.Reply(msg, response)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var read Read
		if err := json.Unmarshal(msg.Body, &read); err != nil {
			return err
		}
		response := ReadResponse{Type: "read_ok", Messages: messageList}
		return n.Reply(msg, response)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var topology Topology
		if err := json.Unmarshal(msg.Body, &topology); err != nil {
			return err
		}
		response := TopologyResponse{Type: "topology_ok"}
		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
