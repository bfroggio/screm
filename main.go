package main

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
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
const twitchHelpCommand string = "sounds"

var hkey = hotkey.New()
var quit = make(chan bool)
var mutex = &sync.Mutex{}
var lastSampleRate beep.SampleRate
var done = make(chan bool)
var pause = make(chan bool)
var ctrl = &beep.Ctrl{}

var welcomedUsers = make(map[string]int)
var recentlyPlayedSounds = make(map[string]string)
var playCounts = make(map[string]map[string]int)

// Used for sorting maps
type keyValue struct {
	Key   string
	Value int
}

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
			twitchWelcome := generateTwitchWelcome(message.User)
			if len(twitchWelcome) > 0 {
				client.Say(viper.GetString("twitch_username"), twitchWelcome)
			}
		}
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		if !executeTwitchMessage(message, allSoundDirectories) {
			if strings.ToLower(message.Message) == twitchHelpCommand {
				twitchHelp := generateTwitchHelp(allSoundDirectories)
				client.Say(viper.GetString("twitch_username"), twitchHelp)
			} else {
				notify(message)
			}
		}
	})

	client.Join(viper.GetString("twitch_username"))

	err = client.Connect()
	if err != nil {
		return err
	}

	return nil
}

func generateTwitchWelcome(user string) string {
	_, userAlreadyWelcomed := welcomedUsers[user]
	if !userAlreadyWelcomed {
		welcomedUsers[user] = 1
		return "Welcome, " + user + "! Type \"" + twitchHelpCommand + "\" to play sound effects!"
	}

	return ""
}

func generateTwitchHelp(allSoundDirectories []string) string {
	// TODO: Limit to only approved users (by message.User.Name)
	helpMessage := "You can play a sound effect in the stream by typing keywords: "

	for _, soundCategory := range allSoundDirectories {
		helpMessage = helpMessage + soundCategory[2:] + " (" + string(soundCategory[0]) + "), "
	}

	return strings.TrimSuffix(helpMessage, ", ")
}

func notify(message twitch.PrivateMessage) {
	// Don't notify me of my own messages
	if message.User.ID != viper.GetString("twitch_username") {
		playSfx(soundsDir+"/chat-notification.ogg", false)
	}
}

func executeTwitchMessage(message twitch.PrivateMessage, allSoundDirectories []string) bool {
	messageContent := strings.ToLower(message.Message)
	log.Println("Got message:", messageContent)

	// TODO: Limit to only approved users (by message.User.Name)
	for _, soundCategory := range allSoundDirectories {
		categoryShortcut := strings.ToLower(string(soundCategory[0]))
		categoryName := strings.ToLower(string(soundCategory[2:]))

		if messageContent == categoryShortcut || messageContent == categoryName {
			log.Println("Playing a \"" + soundCategory + "\" sound at " + message.User.Name + "'s request")
			randomSfx(soundCategory)()
			return true
		}
	}

	return false
}

func configureShortcuts() error {
	hkey.Register(hotkey.Shift+hotkey.Alt, 'Q', func() {
		fmt.Println("Quit")
		quit <- true
	})

	hkey.Register(hotkey.Alt, hotkey.SPACE, func() {
		pause <- true
	})

	err := registerShortcuts()
	if err != nil {
		return err
	}

	fmt.Println("Listening for keyboard shortcuts. Press Shift+Alt+Q to quit.")

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

	mutex.Lock()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	lastSampleRate = format.SampleRate
	mutex.Unlock()

	speaker.Play(beep.Seq(streamer, beep.Callback(func() { done <- true })))
	<-done // Block until the sound file is done playing

	return nil
}

func randomSfx(directory string) func() {
	return func() {
		log.Println("Playing a random sound effect from", directory)

		randomFile, err := getRandomFile(soundsDir + "/" + directory)
		if err != nil {
			log.Println("Error reading file")
		}

		err = playSfx(randomFile, true)
		if err != nil {
			log.Println("Error playing file:", err.Error())
		}
	}
}

func playSfx(path string, replace bool) error {
	go func(replace bool) error {
		streamer, format, err := decodeFile(path)
		if err != nil {
			// TODO: Bubble this error up somehow
			log.Println("Error decoding file:", err.Error())
			return err
		}
		defer streamer.Close()

		mutex.Lock()
		resampled := beep.Resample(4, lastSampleRate, format.SampleRate, streamer)
		lastSampleRate = format.SampleRate
		mutex.Unlock()

		log.Println("Playing " + path)

		if replace {
			speaker.Lock()
			ctrl.Paused = true
			speaker.Unlock()
			ctrl = &beep.Ctrl{Streamer: beep.Seq(resampled, beep.Callback(func() { done <- true })), Paused: false}
			speaker.Play(ctrl)
		} else {
			newCtrl := &beep.Ctrl{Streamer: beep.Seq(resampled, beep.Callback(func() { done <- true })), Paused: false}
			speaker.Play(newCtrl)
		}

		for {
			select {
			case <-done:
				return nil
			case <-pause:
				speaker.Lock()
				ctrl.Streamer = nil
				speaker.Unlock()
				return nil
			}
		}
	}(replace)

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

	// Initialize the map of play counts if needed
	if len(playCounts[directory]) == 0 || len(playCounts[directory]) != len(allFiles) {
		playCounts[directory] = make(map[string]int)

		for _, file := range allFiles {
			playCounts[directory][directory+"/"+file.Name()] = 0
		}
	}

	// Don't play the same sound twice in a row
	randomFile := recentlyPlayedSounds[directory]

	// Only pick a random file if there's more than one file in the directory
	if len(allFiles) > 1 {
		// Sort the list of play counts by number of plays
		sortedPlayCounts := sortMapToSlice(playCounts[directory])

		selectionSize := len(sortedPlayCounts) / 3
		if selectionSize == 0 {
			selectionSize = 1
		}
		log.Println("Selection size:", selectionSize, len(sortedPlayCounts))
		for randomFile == recentlyPlayedSounds[directory] {
			randomIndex := rand.Intn(selectionSize)
			randomFile = sortedPlayCounts[randomIndex].Key
			// Prevent infinite loop conditions caused by directories with a small number of files
			if selectionSize < len(allFiles) {
				selectionSize = selectionSize + 1
			}
		}
	} else {
		randomFile = directory + "/" + allFiles[0].Name()
	}

	recentlyPlayedSounds[directory] = randomFile
	playCounts[directory][randomFile] = playCounts[directory][randomFile] + 1

	return randomFile, nil
}

// This is gross but I'm too lazy to fix it
func sortMapToSlice(m map[string]int) []keyValue {
	var ss []keyValue
	for k, v := range m {
		ss = append(ss, keyValue{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value
	})

	log.Println("Sorted", ss)

	return ss
}
