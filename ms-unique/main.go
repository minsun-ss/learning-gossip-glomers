package main

import (
	"log"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		result := map[string]any{}
		id := uuid.New()

		result["type"] = "generate_ok"
		result["id"] = id.String()

		return n.Reply(msg, result)
	})

	err := n.Run()

	if err != nil {
		log.Fatal(err)
	}
}
