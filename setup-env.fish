#!/opt/homebrew/bin/fish

set -gx alice $(docker exec checkers checkersd keys show alice -a)
set -gx bob $(docker exec checkers checkersd keys show bob -a)

exec fish
