package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var (
	mu sync.RWMutex
)

type KafkaMessage struct {
	Data []int `json:"Data"`
}

type KafkaCommits struct {
	Committed map[string]int `json:"Committed"`
}

func UnmarshalMessage[T any](kvstring any, kvBody T) error {
	if decodedBytes, err := base64.StdEncoding.DecodeString(kvstring.(string)); err != nil {
		return err
	} else {
		if err := json.Unmarshal(decodedBytes, &kvBody); err != nil {
			return fmt.Errorf("Json Unmarshalling Error: %v", err)
		}
		return nil
	}
}

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(n)
	commitlock := "commitlock"

	n.Handle("init", func(msg maelstrom.Message) error {

		ctx := context.Background()
		// grab the commit lock
		kv.CompareAndSwap(ctx, commitlock, 0, 1, true)
		jsonBytes, _ := json.Marshal(KafkaCommits{Committed: map[string]int{}})
		kv.Write(ctx, "commits", jsonBytes)
		kv.CompareAndSwap(ctx, commitlock, 1, 0, true)
		return nil
	})

	// ok another try: let's push each key to a separate key
	n.Handle("send", func(msg maelstrom.Message) error {
		ctx := context.Background()

		response := map[string]any{}
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		key := body["key"].(string)
		lockKey := key + "lock"
		valueFloat := body["msg"].(float64)
		value := int(valueFloat)

		// try to acquire the lock for the key
		for {
			err := kv.CompareAndSwap(ctx, lockKey, 0, 1, true)
			if err != nil {
				time.Sleep(200 * time.Millisecond)
			} else {
				break
			}
		}

		klog, error := kv.Read(ctx, key)
		if error != nil {
			lg := KafkaMessage{
				Data: []int{value},
			}
			jsonBytes, err := json.Marshal(lg)
			if err != nil {
				return err
			}
			kv.Write(ctx, key, jsonBytes)
			response["offset"] = 0
		} else {
			var kvBody KafkaMessage
			UnmarshalMessage(klog, &kvBody)
			kvBody.Data = append(kvBody.Data, value)
			jsonBytes, _ := json.Marshal(kvBody)
			err := kv.Write(ctx, key, jsonBytes)
			if err != nil {
				log.Print("Update seemed to fail here for whatever reason, check the send")
			}
			response["offset"] = len(kvBody.Data) - 1
		}

		response["type"] = "send_ok"
		err := kv.CompareAndSwap(ctx, lockKey, 1, 0, false)
		if err != nil {
			return err
		}

		return n.Reply(msg, response)
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		ctx := context.Background()
		response := map[string]any{}
		response["type"] = "poll_ok"
		response["msgs"] = map[string][][]int{}
		resp := response["msgs"].(map[string][][]int)

		for key, v := range body["offsets"].(map[string]interface{}) {
			startingOffsetFloat := v.(float64)
			startingOffset := int(startingOffsetFloat)
			returnmsgs := [][]int{}
			knode, err := kv.Read(ctx, key)
			if err != nil {
				return err
			}
			var kvBody KafkaMessage
			UnmarshalMessage(knode, &kvBody)
			for pos, val := range kvBody.Data {
				if pos >= startingOffset {
					res := []int{pos, val}
					returnmsgs = append(returnmsgs, res)
				}
			}
			resp[key] = returnmsgs
		}

		return n.Reply(msg, response)
	})

	// this message informs the node that messages have been successfully processed
	// up to and including the given offset
	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		ctx := context.Background()
		response := map[string]any{}

		// acquire lock for commits
		kv.CompareAndSwap(ctx, commitlock, 0, 1, false)

		var kvBody KafkaCommits
		knode, _ := kv.Read(ctx, "commits")
		UnmarshalMessage(knode, &kvBody)

		for key, val := range body["offsets"].(map[string]interface{}) {
			positionFlt := val.(float64)
			position := int(positionFlt)
			kvBody.Committed[key] = position
		}

		jsonBytes, _ := json.Marshal(kvBody)
		kv.Write(ctx, "commits", jsonBytes)
		kv.CompareAndSwap(ctx, commitlock, 1, 0, false) // release lock for commits

		response["type"] = "commit_offsets_ok"
		return n.Reply(msg, response)
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		ctx := context.Background()

		response := map[string]any{}
		response["type"] = "list_committed_offsets_ok"
		response["offsets"] = map[string]int{}

		resp := response["offsets"].(map[string]int)

		var kvBody KafkaCommits
		knode, _ := kv.Read(ctx, "commits")
		UnmarshalMessage(knode, &kvBody)
		for _, v := range body["keys"].([]interface{}) {
			key := v.(string)
			// keys that do not exist on the node can be omitted
			resp[key] = kvBody.Committed[key]
		}

		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
