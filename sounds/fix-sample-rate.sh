#!/usr/bin/env bash

(
    cd "$1"
    for filename in *
    do
        extension="${filename#*.}"
        short_filename="${filename%.*}"

        if extension="mp3"
        then
            ffmpeg -i "$filename" -ar 44100 -vn -c:a libvorbis "out.ogg"
            mv "out.ogg" "$short_filename.ogg"
            rm "$filename"
        else
            ffmpeg -i "$filename" -ar 44100 "sample-rate-fix.$extension"
            mv "sample-rate-fix.$extension" "$filename"
        fi
    done
)