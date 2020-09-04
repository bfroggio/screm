package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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
const twitchCommand string = "!sfx"
const botCheckerAPI string = "https://api.twitchinsights.net/v1/bots/all"

type botCheckerResponse struct {
	Bots  [][]interface{} `json:"bots"`
	Total int             `json:"_total"`
}

var hkey = hotkey.New()
var quit = make(chan bool)
var done = make(chan bool)
var pause = make(chan bool)
var ctrl = &beep.Ctrl{}
var mutex = &sync.Mutex{}

var welcomedUsers = make(map[string]int)
var recentlyPlayedSounds = make(map[string]string)
var playCounts = make(map[string]map[string]int)
var allBots = botCheckerResponse{}

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

	if !viper.GetBool("disable_keyboard_shortcuts") {
		go func() {
			err := configureShortcuts()
			if err != nil {
				log.Fatal("Could not configure shortcuts:", err.Error())
			}
		}()
	}

	go func() {
		err := configureBotChecker()
		if err != nil {
			// Don't end the program if we can't connect to the bot checker API since it's not essential
			log.Print("Could not connect to bot checker API:", err.Error())
		}
	}()

	time.Sleep(1 * time.Second)

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
		client = twitch.NewClient(viper.GetString("twitch_bot_username"), viper.GetString("twitch_secret"))
	} else {
		client = twitch.NewAnonymousClient()
	}

	client.OnUserJoinMessage(func(message twitch.UserJoinMessage) {
		if viper.GetBool("welcome_message_enabled") {
			if len(viper.GetString("twitch_secret")) > 0 {
				if isAuthorized(message.User) {
					twitchWelcome := generateTwitchWelcome(message.User)
					if len(twitchWelcome) > 0 {
						client.Say(viper.GetString("twitch_username"), twitchWelcome)
					}
				}
			}
		}
	})

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		response := executeTwitchMessage(message, allSoundDirectories)
		if len(response) > 0 && len(viper.GetString("twitch_secret")) > 0 {
			responseLines := strings.Split(response, "\\n")
			for _, line := range responseLines {
				log.Println("Saying:", line)
				client.Say(viper.GetString("twitch_username"), line)
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
	if !userAlreadyWelcomed && !isBot(user) {
		welcomedUsers[user] = 1
		return "Welcome, " + user + "! Type \"" + twitchCommand + "\" to play a sound effect on stream!"
	}

	return ""
}

func generateTwitchUnauthorizedMessage(user string) string {
	return "Sorry, " + user + ", you're not authorized to play sound effects. Ask " + viper.GetString("twitch_username") + " nicely to add you to the authorized list!"
}

func generateTwitchHelp(allSoundDirectories []string) string {
	helpMessage := "You can play a sound effect on stream with commands like:\\n"

	for _, soundCategory := range getXRandomItems(allSoundDirectories, 2) {
		helpMessage = helpMessage + twitchCommand + " " + soundCategory[2:] + "\\n"
	}

	return strings.TrimSuffix(helpMessage, "\\n")
}

func getXRandomItems(list []string, itemCount int) []string {
	allItems := []string{}

	for len(allItems) < itemCount {
		randomItem := list[rand.Intn(len(list))]
		if !contains(allItems, randomItem) {
			allItems = append(allItems, randomItem)
		}
	}

	return allItems
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func executeTwitchMessage(message twitch.PrivateMessage, allSoundDirectories []string) string {
	// TODO: Have some sort of backoff for how quickly Twitch can trigger sound effects
	log.Println("Got message:", message.Message)

	// Ignore messages from the bot
	if message.User.Name == viper.GetString("twitch_bot_username") {
		return ""
	}

	// Show a help message if no argument is passed to the command
	if strings.ToLower(message.Message) == twitchCommand {
		if isAuthorized(message.User.Name) {
			return generateTwitchHelp(allSoundDirectories)
		}

		return generateTwitchUnauthorizedMessage(message.User.DisplayName)
	} else if strings.HasPrefix(message.Message, twitchCommand) {
		messageContent := strings.TrimPrefix(strings.ToLower(message.Message), twitchCommand+" ")

		for _, soundCategory := range allSoundDirectories {
			categoryShortcut := strings.ToLower(string(soundCategory[0]))
			categoryName := strings.ToLower(string(soundCategory[2:]))

			if messageContent == categoryShortcut || messageContent == categoryName {
				if isAuthorized(message.User.Name) {
					log.Println("Playing a \"" + soundCategory + "\" sound at " + message.User.DisplayName + "'s request")
					randomSfx(soundCategory)()
					return "Playing a \"" + soundCategory[2:] + "\" sound for " + message.User.DisplayName + "!"
				}

				return generateTwitchUnauthorizedMessage(message.User.DisplayName)
			}
		}
	}

	return ""
}

func isAuthorized(user string) bool {
	allAuthorizedUsers := viper.GetStringSlice("twitch_authorized_users")

	// Let all users play sound effects if we haven't specified a list of authorized users
	if len(allAuthorizedUsers) == 0 {
		return true
	}

	if strings.ToLower(user) == strings.ToLower(viper.GetString("twitch_username")) {
		return true
	}

	for _, authorizedUser := range allAuthorizedUsers {
		if strings.ToLower(user) == strings.ToLower(authorizedUser) {
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
		if ctrl.Streamer != nil {
			pause <- true
		}
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
	path := soundsDir + "/startup.wav"
	_, format, err := decodeFile(path)
	if err != nil {
		return err
	}
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	playSfx(path)
	return nil
}

func randomSfx(directory string) func() {
	return func() {
		log.Println("Playing a random sound effect from", directory)

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
	// Use a Goroutine so keyboard shortcuts work during sound playback
	go func() error {
		mutex.Lock()
		streamer, _, err := decodeFile(path)
		if err != nil {
			// TODO: Bubble this error up somehow
			log.Println("Error decoding sound file:", err.Error())
			return err
		}
		defer streamer.Close()

		log.Println("Playing " + path)

		ctrl.Paused = true
		ctrl = &beep.Ctrl{Streamer: beep.Seq(streamer, beep.Callback(func() { done <- true })), Paused: false}
		speaker.Play(ctrl)
		mutex.Unlock()

		for {
			select {
			case <-done:
				return nil
			case <-pause:
				speaker.Lock()
				ctrl.Paused = true
				ctrl.Streamer = nil
				speaker.Unlock()
				return nil
			}
		}
	}()

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

	return ss
}

func configureBotChecker() error {
	httpClient := http.Client{}

	req, err := http.NewRequest(http.MethodGet, botCheckerAPI, nil)
	if err != nil {
		return err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &allBots)
	if err != nil {
		return err
	}

	return nil
}

func isBot(username string) bool {
	for _, bot := range allBots.Bots {
		if bot[0] == username {
			return true
		}
	}

	return false
}
