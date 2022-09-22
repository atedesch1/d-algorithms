#!/bin/bash

# Terminal Emulator
term="alacritty"

num_procs=$1
starting_port=10000

ports=""

for i in $(seq 1 $num_procs); do
  ports="${ports} :$(($starting_port+$i))"
done

cmd="$term -e bash -c './bin/shared_resource' & sleep .5; $term -e bash -c './bin/process 1 $ports' &"

for i in $(seq 2 $num_procs); do
  cmd="${cmd} sleep .5; $term -e bash -c './bin/process $i $ports' &"
done

eval $cmd