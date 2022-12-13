#!/bin/sh

cmd="docker exec -it checkers $@"
echo $cmd
$cmd
