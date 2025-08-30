1:
	@cd ms-echo && go build -o maelstrom-echo
	@./maelstrom/maelstrom-echo test -w echo --bin ./ms-echo/ggtest --node-count 1 --time-limit 10

2:
	@cd ms-unique && go build -o maelstrom-unique-ids
	@./maelstrom test -w unique-ids --bin ~/ms-unique/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
