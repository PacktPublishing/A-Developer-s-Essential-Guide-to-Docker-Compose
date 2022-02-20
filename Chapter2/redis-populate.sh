#!/bin/sh

cat $1| redis-cli -h $HOST -p $PORT
