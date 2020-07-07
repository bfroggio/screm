package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/MakeNowJust/hotkey"
	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
	"github.com/gempir/go-twitch-irc/v2"
	"github.com/spf13/viper"

	"os"
)

const soundsDir string = "sounds"

var hkey = hotkey.New()
var quit = make(chan bool)
var lastSampleRate beep.SampleRate

func main() {
	rand.Seed(time.Now().Unix())

	err := readConfigFile()
	if err != nil {
		log.Fatal("Could not read config file:", err.Error())
	}

	go func() {
		err := configureTwitch()
		if err != nil {
			log.Fatal("Could not connect to Twitch:", err.Error())
		}
	}()

	go func() {
		err := configureShortcuts()
		if err != nil {
			log.Fatal("Could not configure shortcuts:", err.Error())
		}
	}()

	err = configureSpeaker()
	if err != nil {
		log.Fatal("Could not configure speaker:", err.Error())
	}

	<-quit // Keep the program alive until we kill it with a keyboard shortcut
}

func readConfigFile() error {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	return nil
}

func configureTwitch() error {
	allSoundDirectories, err := getSoundDirectories()
	if err != nil {
		return err
	}

	client := &twitch.Client{}
	if len(viper.GetString("twitch_secret")) > 0 {
		client = twitch.NewClient(viper.GetString("twitch_username"), viper.GetString("twitch_secret"))
	} else {
		client = twitch.NewAnonymousClient()
	}

	client.OnUserJoinMessage(func(message twitch.UserJoinMessage) {
		if len(viper.GetString("twitch_secret")) > 0 {
			twitchHelp := generateTwitchHelp(message.User, allSoundDirectories)
			client.Say(viper.GetString("twitch_username"), twitchHelp)
		}
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if !executeTwitchMessage(message, allSoundDirectories) {
			notify()
		}
	})

	client.Join(viper.GetString("twitch_username"))

	err = client.Connect()
	if err != nil {
		return err
	}

	return nil
}

func generateTwitchHelp(user string, allSoundDirectories []string) string {
	// TODO: Limit to only approved users (by message.User.Name)
	helpMessage := "Welcome, " + user + "! You can play a sound effect in the stream by typing keywords: "

	for _, soundCategory := range allSoundDirectories {
		helpMessage = helpMessage + soundCategory[2:] + " (" + string(soundCategory[0]) + "), "
	}

	return strings.TrimSuffix(helpMessage, ", ")
}

func notify() {
	// TODO: Make this play over top other sound effects
	playSfx(soundsDir + "/chat-notification.ogg")
}

func executeTwitchMessage(message twitch.PrivateMessage, allSoundDirectories []string) bool {
	log.Println("Got message:", message.Message)

	// TODO: Limit to only approved users (by message.User.Name)
	for _, soundCategory := range allSoundDirectories {
		// Remove the first character and the dash from the directory name
		if message.Message == string(soundCategory[0]) || strings.Contains(strings.ToLower(message.Message), soundCategory[2:]) || strings.Contains(soundCategory[2:], strings.ToLower(message.Message)) {
			log.Println("Playing a \"" + soundCategory + "\" sound at " + message.User.Name + "'s request")
			randomSfx(soundCategory)()
			return true
		}
	}

	return false
}

func configureShortcuts() error {
	fmt.Println("Push Shift+Alt+Q to quit")
	hkey.Register(hotkey.Shift+hotkey.Alt, 'Q', func() {
		fmt.Println("Quit")
		quit <- true
	})

	hkey.Register(hotkey.Alt, hotkey.SPACE, func() {
		// TODO: This is dirty, see TODO below
		playSfx("")
	})

	err := registerShortcuts()
	if err != nil {
		return err
	}

	return nil
}

func registerShortcuts() error {
	allFiles, err := getFiles(soundsDir)
	if err != nil {
		return err
	}

	for _, dir := range allFiles {
		if dir.IsDir() {
			hkey.Register(hotkey.Alt, uint32(unicode.ToUpper(rune(dir.Name()[0]))), randomSfx(dir.Name()))
		}
	}

	return nil
}

func configureSpeaker() error {
	path := soundsDir + "/startup.mp3"

	streamer, format, err := decodeFile(path)
	if err != nil {
		return err
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	lastSampleRate = format.SampleRate

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() { done <- true })))
	<-done // Block until the sound file is done playing

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
	// TODO: Figure out a better way to stop playing sound effects
	if len(path) == 0 {
		path = soundsDir + "/silence.ogg"
	}

	streamer, format, err := decodeFile(path)
	if err != nil {
		return err
	}
	defer streamer.Close()

	resampled := beep.Resample(4, lastSampleRate, format.SampleRate, streamer)
	lastSampleRate = format.SampleRate

	log.Println("Playing " + path)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() { done <- true })))
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
