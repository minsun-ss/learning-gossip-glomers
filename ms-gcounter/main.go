package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(n)
	// var current int
	var deltaTotal int

	n.Handle("init", func(msg maelstrom.Message) error {
		kv.Write(context.Background(), "gcounter", 0)
		kv.Write(context.Background(), n.ID(), 0)
		return n.Reply(msg, map[string]any{"type": "init_ok"})
	})

	n.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		deltaFlt := body["delta"].(float64)
		delta := int(deltaFlt)
		deltaTotal += delta

		ctx := context.Background()
		kv.Write(ctx, n.ID(), deltaTotal)

		return n.Reply(msg, map[string]any{"type": "add_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		result, err := kv.ReadInt(context.Background(), "gcounter")
		if err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{"type": "read_ok", "value": result})
	})

	go func() {
		// let init do its magic right about here
		time.Sleep(2000 * time.Millisecond)

		// update the gcounter periodically based on what the nodes are publishing
		for {
			nodeTotal := 0
			ctx := context.Background()
			for _, node := range n.NodeIDs() {
				// hmmm should i continue the node to be 0 or just... what? could keep a prior
				// (stale) counter somewhere
				nValue, err := kv.ReadInt(ctx, node)
				if err != nil {
					continue
				}
				nodeTotal += nValue
			}
			kv.Write(ctx, "gcounter", nodeTotal)

			time.Sleep(200 * time.Millisecond)
		}
	}()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
