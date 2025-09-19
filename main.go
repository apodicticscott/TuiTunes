package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// simple command line args
	showVersion := flag.Bool("version", false, "show version")
	showHelp := flag.Bool("help", false, "show help")
	flag.Parse()

	if *showVersion {
		fmt.Println("TuiTunes v1.0.0")
		os.Exit(0)
	}

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// get music folder, default to current dir
	musicFolder := "."
	if len(flag.Args()) > 0 {
		musicFolder = flag.Args()[0]
	}

	// check if folder exists
	if _, err := os.Stat(musicFolder); os.IsNotExist(err) {
		log.Fatalf("can't find music folder: %s", musicFolder)
	}

	// start the player
	player := NewPlayer(musicFolder)
	if err := player.Initialize(); err != nil {
		log.Fatalf("failed to start player: %v", err)
	}

	// run the app
	app := tea.NewProgram(NewModel(player), tea.WithAltScreen())
	if _, err := app.Run(); err != nil {
		log.Fatalf("app crashed: %v", err)
	}
}

func printHelp() {
	fmt.Print(`TuiTunes - terminal music player

usage: tuitunes [music-folder]

examples:
  tuitunes              # use current folder
  tuitunes ~/Music      # use ~/Music folder

controls:
  space    play/pause
  n        next song
  p        previous song
  r        repeat on/off
  s        shuffle on/off
  /        search
  h        help
  q        quit

navigation:
  up/down  or j/k  move around
  g        go to top
  G        go to bottom
  enter    play song

works with: mp3, wav, flac, m4a, aac, ogg
`)
}
