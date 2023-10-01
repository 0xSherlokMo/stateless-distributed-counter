# stateless-distributed-counter

Implement stateless, global, and distributed safe counter that runs against maelstrom. writes into `Stdout` and reads from `Stdin`. any errors are written directly into `Stderr`. 

`read` RPC is always consistant. the design of this counter allow CP (consistency and partioining). it's designed to be highly available but reads are always correct applying Sequential consistency.

##### How to run?

you can either:

- run each node alone on your own and write to their `Stdin` (write your own client).
- run maelstrom distributed client that has `g-workload` to test the implementation with network-partioning.

  ```bash
  ./maelstrom test -w g-counter \ 
  --bin ~/go/bin/maelstrom-counter \
  --node-count 3 \
  --rate 100 \ 
  --time-limit 20 \ 
  --nemesis partition
  ```

##### TO-DO

- write a retriable errors in case of kv was down (I was straight to the point, so didn't care about it)
