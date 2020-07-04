#!/usr/bin/env bash

(
    cd "$1"

    for i in *.mp3; do
        j=$(echo -n $i | sed -e 's/.mp3/.ogg/g')
        echo "converting $i to $j"
        ffmpeg -y -i "$i"  -strict -2 -acodec vorbis -ac 2 -aq 50 "$j"
        rm "$i"
    done
)