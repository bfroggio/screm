#!/usr/bin/env bash

extension="${1#*.}"
# ffmpeg -i "$1" -filter:a "volume=1.5" "louder.ogg"
ffmpeg -i input.wav -filter:a volumedetect -f null /dev/null
ffmpeg -i "$1" -filter:a loudnorm "louder.$extension"
mv "louder.$extension" "$1"
