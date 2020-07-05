package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"github.com/gempir/go-twitch-irc/v2"

	"os"
)

const soundsDir string = "sounds"

// var hkey = hotkey.New()
var ctrl = &beep.Ctrl{}

func main() {
	rand.Seed(time.Now().Unix())

	err := configureTwitch()
	if err != nil {
		log.Fatal("Could not connect to Twitch:", err.Error())
	}

	err = configureShortcuts()
	if err != nil {
		log.Fatal("Could not configure shortcuts:", err.Error())
	}
}

func configureTwitch() error {
	allSoundDirectories, err := getSoundDirectories()
	if err != nil {
		return err
	}

	client := twitch.NewAnonymousClient()

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// TODO: Limit to only approved users (by message.User.Name)
		for _, soundCategory := range allSoundDirectories {
			// Remove the first character and the dash from the directory name
			if strings.Contains(strings.ToLower(message.Message), soundCategory[2:]) {
				fmt.Println("Playing a \"" + soundCategory + "\" sound at " + message.User.Name + "'s request")
				playSfx(soundCategory)
			}
		}
	})

	client.Join("xqcow")

	err = client.Connect()
	if err != nil {
		return err
	}

	return nil
}

func configureShortcuts() error {
	/*
		quit := make(chan bool)

		fmt.Println("Push Shift+Alt+Q to quit")
		hkey.Register(hotkey.Shift+hotkey.Alt, 'Q', func() {
			fmt.Println("Quit")
			quit <- true
		})

		hkey.Register(hotkey.Alt, hotkey.SPACE, func() {
			ctrl = &beep.Ctrl{}
		})

		err := registerShortcuts()
		if err != nil {
			return err
		}

		<-quit // Keep the program alive until we kill it with a keyboard shortcut
	*/

	return nil
}

func registerShortcuts() error {
	allFiles, err := getFiles(soundsDir)
	if err != nil {
		return err
	}

	for _, dir := range allFiles {
		if dir.IsDir() {
			// TODO: Make sure the uint32 cast works
			// hkey.Register(hotkey.Alt, uint32(unicode.ToUpper(rune(dir.Name()[0]))), randomSfx(dir.Name()))
		}
	}

	return nil
}

func randomSfx(directory string) func() {
	return func() {
		randomFile, err := getRandomFile(soundsDir + "/" + directory)
		if err != nil {
			log.Println("Error reading file")
		}

		err = playSfx(randomFile)
		if err != nil {
			log.Println("Error playing file:", err.Error())
		}
	}
}

func playSfx(path string) error {
	streamer, format, err := decodeFile(path)
	if err != nil {
		return err
	}
	defer streamer.Close()

	sr := format.SampleRate * 2
	speaker.Init(sr, sr.N(time.Second/10))

	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan bool)
	ctrl = &beep.Ctrl{Streamer: beep.Seq(resampled, beep.Callback(func() { done <- true })), Paused: false}
	speaker.Play(ctrl)

	<-done // Block until the sound file is done playing

	return nil
}

func decodeFile(path string) (beep.StreamSeekCloser, beep.Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, beep.Format{}, err
	}

	if strings.Contains(path, ".flac") {
		return flac.Decode(f)
	} else if strings.Contains(path, ".wav") {
		return wav.Decode(f)
	} else if strings.Contains(path, ".mp3") {
		return mp3.Decode(f)
	}

	return vorbis.Decode(f)
}

func getFiles(directory string) ([]os.FileInfo, error) {
	openedDirectory, err := os.Open(directory)
	if err != nil {
		return []os.FileInfo{}, err
	}

	allFiles, err := openedDirectory.Readdir(-1)
	openedDirectory.Close()
	if err != nil {
		return []os.FileInfo{}, err
	}

	return allFiles, nil
}

func getSoundDirectories() ([]string, error) {
	categories := []string{}

	allFiles, err := getFiles(soundsDir)
	if err != nil {
		return []string{}, err
	}

	for _, file := range allFiles {
		if file.IsDir() {
			categories = append(categories, file.Name())
		}
	}

	return categories, nil
}

func getRandomFile(directory string) (string, error) {
	allFiles, err := getFiles(directory)
	if err != nil {
		return "", err
	}

	randomIndex := rand.Intn(len(allFiles))
	return directory + "/" + allFiles[randomIndex].Name(), nil
}
