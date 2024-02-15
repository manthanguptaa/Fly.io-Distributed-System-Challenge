package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var count int = 0

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		count = count + 1
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		unique_id := strconv.FormatInt(time.Now().Unix(), 10) + "-" + string(n.ID()) + "-" + strconv.Itoa(count)
		body["type"] = "generate_ok"
		body["id"] = unique_id
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
