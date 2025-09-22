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

## Challenge 5b:

This one I went back to having a separate lock per key to make this happen, since the keys are now being handled by multiple clients and a local key value map no longer cuts it for storage. The guarantee of linearizability of the key value store makes some of this easier, as you can use CompareAndSwap to swap locks without worrying about data races on the remote lock.

One of the bugs that took me a while to solve is that I needed to grab locks for both the send and the commit actions (since I was storing both by key), otherwise I'd throw errors on CompareAndSwap on update actions (which I kept in for decently strong guarantees on not having race conditions).

## Challenge 5c:

Baseline:

```
 :availability {:valid? true, :ok-fraction 0.99237156},
 :net {:all {:send-count 200998,
             :recv-count 200998,
             :msg-count 200998,
             :msgs-per-op 11.274287},
       :clients {:send-count 42486,
                 :recv-count 42486,
                 :msg-count 42486},
       :servers {:send-count 158512,
                 :recv-count 158512,
                 :msg-count 158512,
                 :msgs-per-op 8.891182},
```

After changing the commits to be a separate key:
```
:availability {:valid? true, :ok-fraction 0.99518573},
:net {:all {:send-count 175694,
            :recv-count 175694,
            :msg-count 175694,
            :msgs-per-op 9.611795},
      :clients {:send-count 44330,
                :recv-count 44330,
                :msg-count 44330},
      :servers {:send-count 131364,
                :recv-count 131364,
                :msg-count 131364,
                :msgs-per-op 7.1866074},
      :valid? true},
```

Given that I want to bring this down to around 6-7 ops, I think I have to find another way to reduce locking entirely.

- Optimistically praying lock contention isn't that frequent is a lie. So now we do need to have lock contention.
- So why not.... batch the jobs? You need to allocate an idx before batching, so no (if I keep the msgs, then maybe?)
- Separating messages to individual keys reduces lock contention on send but increases polling costs to look them all up (went from 9 to 13 ops)
- Routing messages to a single node to handle per key reduced ops to 7 but at the cost of availability going down the drain to .7 or below. Throughput drops a ton. This I imagine is similar to partitioning though.

I think batching would be ideal in which I can provide indexes after the fact and not with the reply.
Maybe I can by returning the msg much later?
