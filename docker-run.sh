#!/bin/sh

cmd="docker run --rm -it -v $(pwd):/checkers -w /checkers -p 1317:1317 -p 3000:3000 -p 4500:4500 -p 5000:5000 -p 26657:26657 --name checkers checkers_i $@"

echo $cmd
$cmd
