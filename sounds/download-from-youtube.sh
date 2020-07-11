#!/usr/bin/env bash

youtube-dl -f "bestaudio[ext=m4a]" -o "$1".m4a "$2"
ffmpeg -i "$1.m4a" -ar 44100 "$1.wav"
rm "$1".m4a
