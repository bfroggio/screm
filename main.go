package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/MakeNowJust/hotkey"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"

	"os"
)

func main() {
	rand.Seed(time.Now().Unix())

	hkey := hotkey.New()

	quit := make(chan bool)

	hkey.Register(hotkey.Alt, 'Q', func() {
		fmt.Println("Quit")
		quit <- true
	})

	// TODO: Dynamically generate these based on directory structure
	hkey.Register(hotkey.Alt, 'D', randomSfx("d-defeat"))
	hkey.Register(hotkey.Alt, 'E', randomSfx("e-epic"))
	hkey.Register(hotkey.Alt, 'F', randomSfx("f-fail"))
	hkey.Register(hotkey.Alt, 'G', randomSfx("g-greetings"))
	hkey.Register(hotkey.Alt, 'J', randomSfx("j-jingles"))
	hkey.Register(hotkey.Alt, 'M', randomSfx("m-music"))
	hkey.Register(hotkey.Alt, 'R', randomSfx("r-random"))
	hkey.Register(hotkey.Alt, 'S', randomSfx("s-success"))
	hkey.Register(hotkey.Alt, 'T', randomSfx("t-fortnite"))
	hkey.Register(hotkey.Alt, 'U', randomSfx("u-upset"))

	fmt.Println("Push Alt-Q to escape and quit")
	<-quit
}

func randomSfx(directory string) func() {
	// TODO: Bubble errors
	return func() { playSfx(getRandomFile("sounds/" + directory)) }
}

func getRandomFile(directory string) string {
	f, err := os.Open(directory)
	if err != nil {
		log.Fatal(err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	randomIndex := rand.Intn(len(files))
	return directory + "/" + files[randomIndex].Name()
}

func playSfx(path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := vorbis.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	sr := format.SampleRate * 2
	speaker.Init(sr, sr.N(time.Second/10))

	resampled := beep.Resample(4, format.SampleRate, sr, streamer)

	done := make(chan bool)
	speaker.Play(beep.Seq(resampled, beep.Callback(func() {
		done <- true
	})))

	<-done
}
