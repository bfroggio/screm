#!/usr/bin/env bash

youtube-dl -f "bestaudio[ext=m4a]" -o "$1".m4a "$2"
ffmpeg -i "$1".m4a -acodec libvorbis "$1".ogg
rm "$1".m4a
