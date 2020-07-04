#!/usr/bin/env bash

(
    cd "$1"

    for i in *.mp3; do
        j=$(echo -n $i | sed -e 's/.mp3/.ogg/g')
        echo "converting $i to $j"
        ffmpeg -i "$i" -vn -c:a libvorbis "$j" && rm "$i"
    done
)