package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var (
	kafkalogs map[string]KafkaLog
	mu        sync.RWMutex
)

type KafkaLog struct {
	LogStream []int
	Committed int
}

func main() {
	mu.Lock()
	kafkalogs = map[string]KafkaLog{}
	mu.Unlock()

	n := maelstrom.NewNode()

	n.Handle("send", func(msg maelstrom.Message) error {
		mu.Lock()
		defer mu.Unlock()

		response := map[string]any{}
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		kkey := body["key"].(string)
		logkeyFloat := body["msg"].(float64)
		logkey := int(logkeyFloat)

		klog, exists := kafkalogs[kkey]
		if !exists {
			kafkalogs[kkey] = KafkaLog{
				LogStream: []int{logkey},
				Committed: 0,
			}
			response["offset"] = 0
		} else {
			klog.LogStream = append(klog.LogStream, logkey)
			kafkalogs[kkey] = klog
			response["offset"] = len(klog.LogStream) - 1
		}

		response["type"] = "send_ok"
		log.Printf("WRITE: key=%s, value=%d, offset=%d", kkey, logkey, response["offset"])

		return n.Reply(msg, response)
	})

	n.Handle("poll", func(msg maelstrom.Message) error {
		mu.Lock()
		defer mu.Unlock()
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		response := map[string]any{}
		response["type"] = "poll_ok"
		response["msgs"] = map[string][][]int{}
		resp := response["msgs"].(map[string][][]int)

		for k, v := range body["offsets"].(map[string]interface{}) {
			kafkaCommitPositionFloat := v.(float64)
			kafkaCommitPosition := int(kafkaCommitPositionFloat)
			kafkanode, exists := kafkalogs[k]
			returnmsgs := [][]int{}

			if exists {
				for pos, val := range kafkanode.LogStream {
					if pos >= kafkaCommitPosition {
						res := []int{pos, val}
						returnmsgs = append(returnmsgs, res)
					}
				}
			}
			resp[k] = returnmsgs
		}

		return n.Reply(msg, response)
	})

	// this message informs the node that messages have been successfully processed
	// up to and including the given offset
	n.Handle("commit_offsets", func(msg maelstrom.Message) error {
		mu.Lock()
		defer mu.Unlock()
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		response := map[string]any{}

		for k, val := range body["offsets"].(map[string]interface{}) {
			positionFlt := val.(float64)
			position := int(positionFlt)
			klog, exists := kafkalogs[k]
			if exists {
				klog.Committed = position
				kafkalogs[k] = klog
			}
		}

		response["type"] = "commit_offsets_ok"

		return n.Reply(msg, response)
	})

	n.Handle("list_committed_offsets", func(msg maelstrom.Message) error {
		mu.Lock()
		defer mu.Unlock()
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		response := map[string]any{}
		response["type"] = "list_committed_offsets_ok"
		response["offsets"] = map[string]int{}

		resp := response["offsets"].(map[string]int)

		for _, v := range body["keys"].([]interface{}) {
			kafkaNodeKey := v.(string)
			kafkanode, exists := kafkalogs[kafkaNodeKey]
			// keys that do not exist on the node can be omitted
			if exists {
				resp[kafkaNodeKey] = kafkanode.Committed
			}
		}

		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}
