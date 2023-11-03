#!/bin/bash

# start all the servers in tmux
for i in {1..9}
do
    ssh ckchu2@fa23-cs425-870$i.cs.illinois.edu "tmux kill-session -t sdfs"
    echo "vm$i done"
done

i=10
ssh ckchu2@fa23-cs425-87$i.cs.illinois.edu "tmux kill-session -t sdfs"
echo "vm$i done"
