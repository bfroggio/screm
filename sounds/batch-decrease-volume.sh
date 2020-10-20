#!/usr/bin/env bash

(
    cd "$1"
    for filename in *
    do
        extension="${filename#*.}"
        ffmpeg -i "$filename" -filter:a "volume=0.75" "quieter.$extension"
        mv "quieter.$extension" "$filename"
    done
)