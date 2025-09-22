# Challenge #5b: Multi-Node Kafka-Style Log

In this challenge, you’ll need to take your Single-Node Kafka system and distribute it out to multiple nodes.

Your nodes can use the linearizable key/value store provided by Maelstrom to implement your distributed, replicated log. This challenge is about correctness and not efficiency. You only need to keep up with a reasonable request rate. It’s important to consider which components require linearizability versus sequential consistency.
