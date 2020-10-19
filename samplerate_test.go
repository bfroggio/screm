package main

import (
	"fmt"
	"log"
	"testing"
	"time"
)

var testDir = "test_sounds/sailing"
var rates = []string{
	"8000",
	"16000",
	"32000",
	"44100",
	"96000",
}

func TestBad(t *testing.T) {
	err := configureSpeaker()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	log.Println("Most of these should sound very bad.")

	for _, i := range rates {
		log.Printf("Playing sample rate: %v.", i)

		wait := make(chan struct{}, 1)
		playSfx(fmt.Sprintf("%s_%s.wav", testDir, i), false, wait)
		<-wait
	}
}

func TestGood(t *testing.T) {
	err := configureSpeaker()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	log.Println("These should sound the same (mostly).")

	for _, i := range rates {
		log.Printf("Playing sample rate: %v.", i)

		wait := make(chan struct{}, 1)
		playSfx(fmt.Sprintf("%s_%s.wav", testDir, i), true, wait)
		<-wait
	}
}
