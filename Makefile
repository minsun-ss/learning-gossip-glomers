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


3:
	@echo "Testing broadcast..."
	@rm -f maelstrom-broadcast
	@cd ms-broadcast && go build -o ../maelstrom-broadcast
	@./maelstrom/maelstrom test -w broadcast --bin ./maelstrom-broadcast --node-count 1 --time-limit 20 --rate 10
	@rm -f maelstrom-broadcast

changelog:
	git cliff --unreleased --tag v$(shell cat VERSION) --prepend changelog.md
