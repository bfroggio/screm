#!/usr/bin/env bash

(
    cd "$1"
    for filename in *
    do
        extension="${filename#*.}"
        short_filename="${filename%.*}"

        ffmpeg -i "$filename" -ar 44100 "out.wav"
        rm "$filename"
        mv "out.wav" "$short_filename.wav"
    done
)