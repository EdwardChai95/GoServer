#!/bin/bash

pid=`ps -ef | grep go | grep -v grep | awk '{print $2}'`
if [ -n "$pid" ]
then
	echo "kill -9 pid:" $pid
	kill -9 $pid
fi
