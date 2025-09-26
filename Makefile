clean:
	@echo "Wiping out the store..."
	@rm -rf store

1:
	@echo "Testing gossip glomers echo..."
	@rm -f maelstrom-echo
	@cd ms-echo && go build -o ../maelstrom-echo
	@./maelstrom/maelstrom test -w echo --bin maelstrom-echo --node-count 1 --time-limit 10
	@rm -f maelstrom-echo

2:
	@echo "Testing unique ids..."
	@rm -f maelstrom-unique-ids
	@cd ms-unique && go build -o ../maelstrom-unique-ids
	@./maelstrom/maelstrom test -w unique-ids --bin ./maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
	@rm -f maelstrom-unique-ids

3a:
	@echo "Testing broadcast part A..."
	@rm -f maelstrom-broadcast
	@cd ms-broadcast && go build -o ../maelstrom-broadcast
	@./maelstrom/maelstrom test -w broadcast --bin ./maelstrom-broadcast --node-count 1 --time-limit 20 --rate 10
	@rm -f maelstrom-broadcast

3b:
	@echo "Testing broadcast part B..."
	@rm -f maelstrom-broadcast2
	@cd ms-broadcast2 && go build -o ../maelstrom-broadcast2
	@./maelstrom/maelstrom test -w broadcast --bin ./maelstrom-broadcast2 --node-count 5 --time-limit 20 --rate 10
	@rm -f maelstrom-broadcast2

3c:
	@echo "Testing broadcast part C..."
	@rm -f maelstrom-broadcast3
	@cd ms-broadcast3 && go build -o ../maelstrom-broadcast3
	@./maelstrom/maelstrom test -w broadcast --bin ./maelstrom-broadcast3 --node-count 5 --time-limit 20 --rate 10 --nemesis partition
	@rm -f maelstrom-broadcast3

3d:
	@echo "Testing broadcast part D..."
	@rm -f maelstrom-broadcast2
	@cd ms-broadcast2 && go build -o ../maelstrom-broadcast2
	@./maelstrom/maelstrom test -w broadcast --bin ./maelstrom-broadcast2 --node-count 25 --time-limit 20 --rate 100 --latency 100
	@rm -f maelstrom-broadcast2

3e:
	@echo "Testing broadcast part E..."
	@rm -f maelstrom-broadcast3
	@cd ms-broadcast3 && go build -o ../maelstrom-broadcast3
	@./maelstrom/maelstrom test -w broadcast --bin ./maelstrom-broadcast3 --node-count 25 --time-limit 20 --rate 100 --latency 100
	@rm -f maelstrom-broadcast3

4:
	@echo "Testing grow only counter..."
	@rm -f maelstrom-gcounter
	@cd ms-gcounter && go build -o ../maelstrom-gcounter
	@./maelstrom/maelstrom test -w g-counter --bin ./maelstrom-gcounter --node-count 3 --rate 100 --time-limit 20 --nemesis partition
	@rm -f maelstrom-gcounter

5a:
	@echo "Testing kafka node part A..."
	@rm -f maelstrom-kafka
	@cd ms-kafka && go build -o ../maelstrom-kafka
	@./maelstrom/maelstrom test -w kafka --bin ./maelstrom-kafka --node-count 1 --concurrency 2n --time-limit 20 --rate 1000
	@rm -f maelstrom-kafka

5b:
	@echo "Testing kafka node part B..."
	@rm -f maelstrom-kafka2
	@cd ms-kafka2 && go build -o ../maelstrom-kafka2
	@./maelstrom/maelstrom test -w kafka --bin ./maelstrom-kafka2 --node-count 2 --concurrency 2n --time-limit 20 --rate 1000
	@rm -f maelstrom-kafka2

5c:
	@echo "Testing kafka node part C..."
	@rm -f maelstrom-kafka3
	@cd ms-kafka3 && go build -o ../maelstrom-kafka3
	@./maelstrom/maelstrom test -w kafka --bin ./maelstrom-kafka3 --node-count 2 --concurrency 2n --time-limit 20 --rate 1000
	@rm -f maelstrom-kafka3

6a:
	@echo "Testing kv store part A..."
	@rm -f maelstrom-kv
	@cd ms-kv && go build -o ../maelstrom-kv
	@./maelstrom/maelstrom test -w txn-rw-register --bin ./maelstrom-kv --node-count 1 --time-limit 20 --rate 1000 --concurrency 2n --consistency-models read-uncommitted --availability total
	@rm -f maelstrom-kv

6b1:
	@echo "Testing kv store part B..."
	@rm -f maelstrom-kv
	@cd ms-kv && go build -o ../maelstrom-kv
	@./maelstrom/maelstrom test -w txn-rw-register --bin ./maelstrom-kv --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted
	@rm -f maelstrom-kv

6b2:
	@echo "Testing kv store part B..."
	@rm -f maelstrom-kv
	@cd ms-kv && go build -o ../maelstrom-kv
	@./maelstrom/maelstrom test -w txn-rw-register --bin ./maelstrom-kv --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted --availability total --nemesis partition
	@rm -f maelstrom-kv


6c:
	@echo "Testing kv store part C..."
	@rm -f maelstrom-kv
	@cd ms-kv && go build -o ../maelstrom-kv
	@./maelstrom/maelstrom test -w txn-rw-register --bin ./maelstrom-kv --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-committed --availability total –-nemesis partition
	@rm -f maelstrom-kv

changelog:
	git cliff --unreleased --tag v$(shell cat VERSION) --prepend changelog.md
