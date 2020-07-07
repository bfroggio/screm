#!/usr/bin/env bash

extension="${1#*.}"
echo "$extension"
# ffmpeg -i "$1" -filter:a "volume=1.5" "louder.ogg"
ffmpeg -i "$1" -filter:a loudnorm "louder.$extension"
mv "louder.$extension" "$1"
