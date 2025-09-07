package main

import (
	"encoding/json"
	"fmt"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var seen []int

func main() {
	n := maelstrom.NewNode()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		result := map[string]any{}
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		message, ok := body["message"].(float64)
		if !ok {
			return fmt.Errorf("expected int, got %T", body["message"])
		}
		messageInt := int(message)

		result["type"] = "broadcast_ok"
		seen = append(seen, messageInt)

		return n.Reply(msg, result)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		result := map[string]any{}

		result["type"] = "read_ok"
		result["messages"] = seen

		return n.Reply(msg, result)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		result := map[string]any{}

		result["type"] = "topology_ok"
		return n.Reply(msg, result)
	})

	err := n.Run()

	if err != nil {
		log.Fatal(err)
	}
}
