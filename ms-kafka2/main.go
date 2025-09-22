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
	kafkalogs map[string]KafkaLog
	mu        sync.RWMutex
)

type KafkaLog struct {
	LogStream []int `json:"LogStream"`
	Committed int   `json:"Committed"`
}

func UnmarshalKV(kvstring any, kvBody *KafkaLog) error {
	if decodedBytes, err := base64.StdEncoding.DecodeString(kvstring.(string)); err != nil {
		return err
	} else {
		if err := json.Unmarshal(decodedBytes, &kvBody); err != nil {
			return fmt.Errorf("Json Unmarshalling Error: %v", err)
		}
		return nil
	}
}

// let's create a lock key
func main() {
	mu.Lock()
	kafkalogs = map[string]KafkaLog{}
	mu.Unlock()

	n := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(n)

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
			lg := KafkaLog{
				LogStream: []int{value},
				Committed: 0,
			}
			jsonBytes, err := json.Marshal(lg)
			if err != nil {
				return err
			}
			kv.Write(ctx, key, jsonBytes)
			response["offset"] = 0
		} else {
			var kvBody KafkaLog
			var kvBodyTest KafkaLog
			UnmarshalKV(klog, &kvBody)
			UnmarshalKV(klog, &kvBodyTest)
			log.Printf("Unmarshalling in send for key %s %v", key, kvBody)
			kvBody.LogStream = append(kvBody.LogStream, value)
			jsonBytes, _ := json.Marshal(kvBody)
			err := kv.CompareAndSwap(ctx, key, klog, jsonBytes, false)
			if err != nil {
				log.Print("Update seemed to fail here for whatever reason, check the send")
			}
			response["offset"] = len(kvBody.LogStream) - 1
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
			var kvBody KafkaLog
			UnmarshalKV(knode, &kvBody)
			log.Printf("Unmarshalling in poll: %s %v", key, kvBody)
			for pos, val := range kvBody.LogStream {
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

		for key, val := range body["offsets"].(map[string]interface{}) {

			lockKey := key + "lock"
			for {
				err := kv.CompareAndSwap(ctx, lockKey, 0, 1, true)
				if err != nil {
					time.Sleep(200 * time.Millisecond)
				} else {
					break
				}
			}

			positionFlt := val.(float64)
			position := int(positionFlt)
			knode, err := kv.Read(ctx, key)
			if err != nil {
				// do nothing
			} else {
				var kvBody KafkaLog
				UnmarshalKV(knode, &kvBody)
				log.Printf("Unmarshalling in commit offsets: %s %v", key, kvBody)
				kvBody.Committed = position
				jsonBytes, _ := json.Marshal(kvBody)
				kv.Write(ctx, key, jsonBytes)
			}
			kv.CompareAndSwap(ctx, lockKey, 1, 0, false)
		}

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

		for _, v := range body["keys"].([]interface{}) {
			key := v.(string)
			knode, error := kv.Read(ctx, key)
			// keys that do not exist on the node can be omitted
			if error != nil {
				// do nothing
			} else {
				var kvBody KafkaLog
				UnmarshalKV(knode, &kvBody)
				log.Printf("Unmarshalling in list committed offsets: %s %v", key, kvBody)

				resp[key] = kvBody.Committed
			}
		}

		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
