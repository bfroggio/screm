#!/usr/bin/env bash

extension="${1#*.}"
ffmpeg -i "$1" -filter:a "volume=0.5" "louder.$extension"
mv "louder.$extension" "$1"
