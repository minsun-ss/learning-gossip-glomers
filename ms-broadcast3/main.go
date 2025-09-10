package main

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var NodeID string
var seen []int
var mu sync.Mutex

type QueuedMessage struct {
	Source      string
	Destination string
	Msg         map[string]any
}

func ReadHandler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		result := map[string]any{}

		result["type"] = "read_ok"
		result["messages"] = seen

		return n.Reply(msg, result)
	}
}

func BroadcastHandler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
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
		// also append to queue to validate that it was confirmed sent
		containsMessage := slices.Contains(seen, messageInt)

		if !containsMessage {
			mu.Lock()
			seen = append(seen, messageInt)
			mu.Unlock()
		}

		result["type"] = "broadcast_ok"
		return n.Reply(msg, result)

	}
}

func TopologyHandler(n *maelstrom.Node, topology map[string][]string) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
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
	}
}

func InitHandler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		NodeID = n.ID()
		return nil
	}
}

func GossipHandler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return err
		}

		data, ok := body["data"].([]interface{})
		if ok {
			for _, val := range data {
				valFloat, ok := val.(float64)
				if ok {
					valInt := int(valFloat)
					if !slices.Contains(seen, valInt) {
						mu.Lock()
						seen = append(seen, valInt)
						mu.Unlock()
					}
				}
			}
		}

		return nil
	}
}

func main() {
	n := maelstrom.NewNode()
	topology := map[string][]string{}

	n.Handle("init", InitHandler(n))
	n.Handle("broadcast", BroadcastHandler(n))
	n.Handle("read", ReadHandler(n))
	n.Handle("topology", TopologyHandler(n, topology))
	n.Handle("broadcast_ok", func(msg maelstrom.Message) error {
		return nil
	})
	n.Handle("gossip", GossipHandler(n))

	go func() {
		for {
			for node := range topology {
				result := map[string]any{}
				result["type"] = "gossip"
				result["data"] = seen

				if node != NodeID && len(seen) > 0 {
					n.Send(node, result)
				}
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()

	err := n.Run()

	if err != nil {
		log.Fatal(err)
	}
}
