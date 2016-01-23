#!/usr/bin/env bash

for d in */ ; do
	if [ "$d" != "bower_components/" ]; then
        cd $d
        echo "Building $d"
        gopherjs build -m
        cd ..
    fi
done