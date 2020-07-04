package main

import (
	"fmt"
	"log"
	"time"

	"github.com/MakeNowJust/hotkey"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

	"os"
)

func main() {
	f, err := os.Open("sounds/e-epic/Disney Friend Like Me.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(streamer)
	select {}
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
