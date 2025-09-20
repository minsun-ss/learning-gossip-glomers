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
- 3d was a very dumb solution (for me) - since there were no partitions, I went back to my 3b solution and it worked out of the box.
- 3e: cleaned up the code to look more "go" ish and changed the update to send a single gossip message instead of spamming more. Learned the hard way that Go's JSON marshalling marshals to float64 instead of int, and you have to go through float64 and recast to int to make it work. Think I probably could use some better error handling. Also discovered somewhat (belatedly) that the /store folder actually provides you the logs of your node's output to directly inspect instead of squinting at the command line. I feel moderately dumb for not realizing this.

## Challenge 4:

I thought I should've locked after every transaction and was using a separate key to track that, but it seemed to cause different threads to crash in various states of deadlock. Then I was like mm, every node knows what its own total should be and that's always good, so have it periodically update the value in the kv store with it under its own node id and then aggregate it all periodically. That seemed to work in a non crashy way. Didn't really know how to handle when there was network outages to the key value store itself, though, there didn't seem a good way to grab multiple keys in a single transaction (which would've been nice).

## Challenge 5a:

Implementing the requirements was not too tricky, although I got hung up on a bug for almost 2 hours due to me misreading the format of the poll - offset *first*, then msg, rather than msg, offset and one that Claude eventually pointed out to me was the bug after going down some unnecessary mutex rabbitholes.

As this challenge goes, this one had a relatively straightforward race conditions on your data structure, so as long as you were threadsafe/locked your data structure with a mutex of some sort, it was mostly ok. Learned the difference between different type of go mutexes with this (e.g., RLock vs Lock).
