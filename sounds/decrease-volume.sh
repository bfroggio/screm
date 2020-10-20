#!/usr/bin/env bash

extension="${1#*.}"
ffmpeg -i "$1" -filter:a "volume=0.5" "quieter.$extension"
mv "quieter.$extension" "$1"
