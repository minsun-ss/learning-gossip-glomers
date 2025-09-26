package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var kv map[int]int
var mu sync.RWMutex

func main() {
	n := maelstrom.NewNode()
	kv := map[int]int{}

	n.Handle("txn", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		response := map[string]any{"type": "txn_ok"}

		// I think reads have to come in before the writes
		for _, val := range body["txn"].([]interface{}) {
			txn := val.([]any)
			switch txn[0] {
			case "r":
				keyFlt := txn[1].(float64)
				key := int(keyFlt)
				if value, exists := kv[key]; exists {
					txn[2] = float64(value)
				} else {
					txn[2] = nil
				}
			}
		}

		for _, val := range body["txn"].([]interface{}) {
			txn := val.([]any)
			switch txn[0] {
			case "w":
				keyFlt := txn[1].(float64)
				key := int(keyFlt)
				valueFlt := txn[2].(float64)
				value := int(valueFlt)
				mu.Lock()
				kv[key] = value
				mu.Unlock()
			}
		}

		response["txn"] = body["txn"]

		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
