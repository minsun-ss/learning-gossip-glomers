# Gossip Glomers 

Notes: maelstrom is not happy with log/slog output to stdout

Part of Recurse Fall 2025 DDIA readings!

# Thoughts

## Challenge 1: 

Very straightforward, more a setup than anything else.

## Challenge 2: 

I think the use of golang makes this easier than it could be? If I needed to provide an int ID, though, I think this might be more challenging.

## Challenge 3: 

- 3a was very straightforward on single node setup...
- ...and so was 3b, if you are ok with thinking topology never changes once provided to you (which... probably seems wrong on a distributed setup)
- I did not realize 3c existed until John told me about it. T_T .... 3 hours later, I implemented what I feel was the dumbest solution on the planet to make this work
- 3d was dumb... cos I went to my 3b implementation ahahaha
- 3e: cleaned up the code to look more "go" ish and changed the update to send a single gossip message instead of spamming more
