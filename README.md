# Screm Bot 3000

[![Man Hours](https://img.shields.io/endpoint?url=https%3A%2F%2Fmh.jessemillar.com%2Fhours%3Frepo%3Dhttps%3A%2F%2Fgithub.com%2Fjessemillar%2Fscrem.git)](https://jessemillar.com/r/man-hours)

## Overview

I like playing comedic sound effects while streaming videogames but couldn't find a sound board program that let me configure my sound effects and keyboard shortcuts with Git version control while also randomizing the sound effects that were played (it's more fun if I'm surprised by the sound effect as well as my viewers). Thus, Screm Bot 3000 was born!

## Usage

1. Create a `config.toml` file as outlined in the ["Config File" section](#config-file) below
1. Put any sound effects you want to use in `./sounds` inside a directory with a name something like `./sounds/e-epic`
	- The first letter of the directory name (e.g. `e` in the example above) is the keyboard button used to play a random sound file from that directory (`Alt + e` in this case)
1. Launch `screm.exe` by double clicking on it
1. Trigger a sound effect with your configured keyboard shortcuts!

## Config File

`config.toml` should be in the same directory as the `screm.exe` binary. There are a few properties that can go in your `config.toml` file. Properties listed below are optional unless otherwise noted. See `sample-config.toml` for an example with fake configuration values.

- `twitch_username` (required): The username for your Twitch account/channel. Screm Bot 3000 can't read your channel's chat messages without this
- `twitch_secret`: The OAUTH token needed for Screm Bot 3000 to post messages to your Twitch chat
- `twitch_authorized_users`: A list of Twitch usernames for users authorized to trigger sound effects on your stream from Twitch chat

## FAQ

### Does this work on any operating systems besides Windows?

No. Not at the moment at least. The library I'm using for capturing global keyboard shortcuts is Windows-specific and my use case only involves Windows.

### Why does the program instantly crash when I open `screm.exe`?

You likely don't have a valid `config.toml` created. Try running the program with `go run main.go` to get a more helpful error.

### The Windows Sound Mixer can't change the output device of `screm.exe`.

I know. It's a bug somewhere in the audio library used by Screm Bot 3000. Follow the steps below to fix it every time you launch `screm.exe`. If it doesn't work the first time, try again. The workaround's fiddly.

1. Open "Sound Mixer"
1. Change your default sound output device (the dropdown near the top of the page) from its current device (e.g. "USB Sound Card") to whatever device you want Screm Bot 3000 to output to (e.g. "VoiceMeeter Input")
1. Start `screm.exe`
1. Play a sound effect with a keyboard shortcut
1. While the sound effect is playing:
	1. Switch `screm.exe`'s output to your desired output (e.g. "VoiceMeeter Input")
	1. Change the default sound output to the previous value (e.g. "USB Sound Card")
1. Verify that sounds from various programs are going through your desired outputs

### Some of my sound files sound like a whale or chipmunk...?

Yeah, so, there's this thing in audio files called a "sample rate". It's basically how fast the sound data gets played to the speaker. The [audio library I'm using](https://github.com/faiface/beep) is *supposed* to support different sample rates across your collection of sound files, but I've tested, retested, refactored, and cried over why that's not working so instead I hardcoded the library to expect 44100 as the sample rate and put a `fix-sample-rates.sh` script in `./sounds` that forces all audio files in a provided directory (e.g. `./fix-sample-rates.sh e-epic`) to 44100 `.wav`. That's the only way I've found to prevent crashes. Writing Screm Bot 3000 in Go was a mistake.