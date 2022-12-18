> Note: I no longer stream to Twitch. You're welcome to try using this project or to modify the code yourself, but I am unable to offer support if you get stuck.

# Screm Bot 3000

## Overview

I like playing comedic sound effects while streaming videogames but couldn't find a sound board program that let me configure my sound effects and keyboard shortcuts with Git version control while also randomizing the sound effects that were played (it's more fun if I'm surprised by the sound effect as well as my viewers). I also wanted my Twitch chat to be able to trigger sound effects via a chat command (`!sfx`). Thus, Screm Bot 3000 was born!

## Usage

1. Create a `config.toml` file as outlined in the ["Config File" section](#config-file) below
1. Put any sound effects you want to use in `./sounds` inside a directory with a name something like `./sounds/e-epic`
	- The first letter of the directory name (e.g. `e` in the example above) is the keyboard button used to play a random sound file from that directory (`Alt + e` in this case)
1. Launch `screm.exe` by double clicking on it
1. Trigger a sound effect with your configured keyboard shortcuts!

## Config File

`config.toml` should be in the same directory as the `screm.exe` binary. There are a few properties that can go in your `config.toml` file. Properties listed below are optional unless otherwise noted. See `sample-config.toml` for an example with fake configuration values.

- `twitch_username` (required): The username for your Twitch account/channel. Screm Bot 3000 can't read your channel's chat messages without this.
- `twitch_bot_username` (required): The username for the Screm chat bot. Can be the same as `twitch_username`.
- `twitch_secret`: The OAUTH token for the `twitch_bot_username` account. Needed to post messages to your Twitch chat.
- `twitch_authorized_users`: A list of Twitch usernames for users authorized to trigger sound effects on your stream from Twitch chat.
- `welcome_message_enabled`: Whether or not Screm Bot 3000 should automatically welcome non-bot viewers to your channel (defaulted to `false`).
- `disable_keyboard_shortcuts`: If you don't want to automatically configure keyboard shortcuts (because you use another soundboard program like an Elgato Stream Deck), you can set `disable_keyboard_shortcuts` to `true` (defaults to `false`) and Screm Bot 3000 will only listen for Twitch chat messages.

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
