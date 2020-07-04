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

	playSfx(getRandomFile("sounds/e-epic"))
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

func setupHotkeys() {
	hkey := hotkey.New()

	quit := make(chan bool)

	hkey.Register(hotkey.Alt, 'M', func() {
		fmt.Println("Quit")
		quit <- true
	})

	fmt.Println("Start hotkey's loop")
	fmt.Println("Push Ctrl-Q to escape and quit")
	<-quit
}
