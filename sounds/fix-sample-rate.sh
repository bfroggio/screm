#!/usr/bin/env bash

(
    cd "$1"
    for filename in *
    do
        extension="${filename#*.}"
        short_filename="${filename%.*}"

        ffmpeg -i "$filename" -ar 44100 -vn -c:a libvorbis "out.ogg"
        rm "$filename"
        mv "out.ogg" "$short_filename.ogg"
    done
)