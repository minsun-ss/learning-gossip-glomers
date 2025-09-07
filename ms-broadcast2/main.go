package main

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var NodeID string
var seen []int

func main() {
	n := maelstrom.NewNode()
	topology := map[string][]string{}

	// need to figure out the nodes
	n.Handle("init", func(msg maelstrom.Message) error {
		// result := map[string]any{}
		NodeID = n.ID()
		log.Printf("What is my id: %s", NodeID)

		return nil
	})

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		result := map[string]any{}
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// store the messages into its own item
		message, ok := body["message"].(float64)
		if !ok {
			return fmt.Errorf("expected int, got %T", body["message"])
		}
		messageInt := int(message)

		// check if already in seen, and if true, don't broadcast, if false do broadcast and append to seen
		// this very much assumes topology doesn't magically change on me though
		containsMessage := slices.Contains(seen, messageInt)

		if !containsMessage {
			seen = append(seen, messageInt)
			for _, node := range topology[NodeID] {
				n.Send(node, body)
			}
		}

		result["type"] = "broadcast_ok"
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
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		rawTopology := body["topology"].(map[string]interface{})

		for key, value := range rawTopology {
			interfaces := value.([]interface{})
			strings := make([]string, len(interfaces))
			for i, v := range interfaces {
				strings[i] = v.(string)
			}
			topology[key] = strings
		}

		result["type"] = "topology_ok"
		return n.Reply(msg, result)
	})

	err := n.Run()

	if err != nil {
		log.Fatal(err)
	}
}
