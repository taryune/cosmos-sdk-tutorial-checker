#!/bin/sh

export alice=$(docker exec checkers checkersd keys show alice -a)
export bob=$(docker exec checkers checkersd keys show bob -a)
