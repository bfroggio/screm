package main

import (
	"fmt"

	"github.com/MakeNowJust/hotkey"
)

func main() {
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
