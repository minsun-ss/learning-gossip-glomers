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

var n maelstrom.Node
var NodeID string
var seen []int
var queuedMessages []QueuedMessage
var mu sync.Mutex

type QueuedMessage struct {
	Source      string
	Destination string
	Msg         map[string]any
}

// so if I understand correctly, RPC will return an acknowledgement

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
		// also append to queue to validate that it was confirmed sent
		containsMessage := slices.Contains(seen, messageInt)

		if !containsMessage {
			seen = append(seen, messageInt)
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

	n.Handle("broadcast_ok", func(msg maelstrom.Message) error {
		return nil
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

	go func() {
		for {
			for node := range topology {
				result := map[string]any{}
				result["type"] = "broadcast"

				for i := range seen {
					if node != NodeID {
						result["message"] = i
						n.Send(node, result)
					}
				}
			}
		}
		time.Sleep(2)
	}()

	err := n.Run()

	if err != nil {
		log.Fatal(err)
	}
}

func Send(n *maelstrom.Node, seen []int, dest string, msg map[string]any) {
	go func() {
		for _, val := range seen {
			msg["message"] = val
			n.Send(dest, msg)
		}
	}()
}

// take a confirmation message and knock it off the queue
func CheckStatus(msg maelstrom.Message) error {
	var body map[string]any
	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		return err
	}

	// remove item from the queue if the broadcast is ok
	if body["type"] == "broadcast_ok" {
		result := string(msg.Body)
		log.Printf("error message src: %s dest %s, raw %s", msg.Src, msg.Dest, result)

		for idx, m := range queuedMessages {
			data, _ := json.Marshal(m.Msg)
			if (m.Source == msg.Dest) && (m.Destination == msg.Src) && (body["in_reply_to"] == m.Msg["msg_id"]) {
				log.Printf("FOUND queued source %s queued destination %s queued %s", m.Source, m.Destination, data)
				mu.Lock()
				if idx < len(queuedMessages) {
					queuedMessages = append(queuedMessages[:idx], queuedMessages[idx+1:]...)
				}
				mu.Unlock()
				break
			}

		}
	}
	return nil
}
