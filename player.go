package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

// song info
type Track struct {
	Path     string
	Title    string
	Artist   string
	Album    string
	Duration time.Duration
	Format   string
}

// what the player is doing
type PlayerState int

const (
	Stopped PlayerState = iota
	Playing
	Paused
)

// the music player
type Player struct {
	MusicDir       string
	Tracks         []Track
	CurrentTrack   int
	State          PlayerState
	Repeat         bool
	Shuffle        bool
	SearchQuery    string
	FilteredTracks []Track

	// audio stuff
	Streamer    beep.StreamSeeker
	Ctrl        *beep.Ctrl
	Format      beep.Format
	LoadedTrack int // which song is actually loaded
}

// make a new player
func NewPlayer(musicDir string) *Player {
	return &Player{
		MusicDir: musicDir,
		Repeat:   false,
		Shuffle:  false,
	}
}

// start up the player
func (p *Player) Initialize() error {
	// setup audio
	sampleRate := beep.SampleRate(44100)
	speaker.Init(sampleRate, sampleRate.N(time.Second/10))

	// find all the music
	if err := p.ScanMusicFiles(); err != nil {
		return fmt.Errorf("couldn't find music: %w", err)
	}

	// start with all tracks visible
	p.FilteredTracks = p.Tracks

	return nil
}

// find all music files
func (p *Player) ScanMusicFiles() error {
	// what file types we can play
	formats := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".flac": true,
		".m4a":  true,
		".aac":  true,
		".ogg":  true,
	}

	p.Tracks = []Track{}

	err := filepath.WalkDir(p.MusicDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if formats[ext] {
			track := Track{
				Path:   path,
				Title:  filepath.Base(path),
				Format: ext,
			}

			// get song info from filename
			p.extractMetadata(&track)
			p.Tracks = append(p.Tracks, track)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// sort by filename
	sort.Slice(p.Tracks, func(i, j int) bool {
		return p.Tracks[i].Path < p.Tracks[j].Path
	})

	return nil
}

// get song info from filename and folder
func (p *Player) extractMetadata(track *Track) {
	// just use filename for now
	baseName := filepath.Base(track.Path)
	ext := filepath.Ext(baseName)
	track.Title = strings.TrimSuffix(baseName, ext)

	// try to get artist/album from folder structure
	relPath, _ := filepath.Rel(p.MusicDir, track.Path)
	parts := strings.Split(filepath.Dir(relPath), string(filepath.Separator))

	if len(parts) >= 2 {
		track.Artist = parts[0]
		track.Album = parts[1]
	} else if len(parts) == 1 && parts[0] != "." {
		track.Artist = parts[0]
	}
}

// play the current song
func (p *Player) Play() error {
	if len(p.FilteredTracks) == 0 {
		return fmt.Errorf("no songs found")
	}

	if p.CurrentTrack >= len(p.FilteredTracks) {
		p.CurrentTrack = 0
	}

	// if we're paused on the same song, just resume
	if p.State == Paused && p.Streamer != nil && p.LoadedTrack == p.CurrentTrack {
		p.Ctrl.Paused = false
		p.State = Playing
		return nil
	}

	// if already playing the same song, do nothing
	if p.State == Playing && p.Streamer != nil && p.LoadedTrack == p.CurrentTrack {
		return nil
	}

	track := p.FilteredTracks[p.CurrentTrack]

	// stop whatever's playing
	if p.Ctrl != nil {
		p.Ctrl.Paused = true
	}

	// load the new song
	streamer, format, err := p.loadTrack(track.Path)
	if err != nil {
		return fmt.Errorf("can't load song: %w", err)
	}

	p.Streamer = streamer
	p.Format = format
	p.Ctrl = &beep.Ctrl{Streamer: beep.ResampleRatio(4, 1.0, streamer), Paused: false}
	p.LoadedTrack = p.CurrentTrack

	speaker.Play(p.Ctrl)
	p.State = Playing

	return nil
}

// pause the song
func (p *Player) Pause() {
	if p.Ctrl == nil {
		return
	}

	if p.State == Playing {
		p.Ctrl.Paused = true
		p.State = Paused
	}
}

// resume from where we paused
func (p *Player) Resume() {
	if p.Ctrl == nil {
		return
	}

	if p.State == Paused {
		p.Ctrl.Paused = false
		p.State = Playing
	}
}

// stop everything
func (p *Player) Stop() {
	if p.Ctrl != nil {
		p.Ctrl.Paused = true
	}
	p.State = Stopped
	p.LoadedTrack = -1
}

// go to next song
func (p *Player) NextTrack() error {
	if len(p.FilteredTracks) == 0 {
		return fmt.Errorf("no songs")
	}

	p.CurrentTrack++
	if p.CurrentTrack >= len(p.FilteredTracks) {
		if p.Repeat {
			p.CurrentTrack = 0
		} else {
			p.CurrentTrack = len(p.FilteredTracks) - 1
			return fmt.Errorf("at the end")
		}
	}

	return p.Play()
}

// go to previous song
func (p *Player) PreviousTrack() error {
	if len(p.FilteredTracks) == 0 {
		return fmt.Errorf("no songs")
	}

	p.CurrentTrack--
	if p.CurrentTrack < 0 {
		if p.Repeat {
			p.CurrentTrack = len(p.FilteredTracks) - 1
		} else {
			p.CurrentTrack = 0
			return fmt.Errorf("at the beginning")
		}
	}

	return p.Play()
}

// toggle repeat on/off
func (p *Player) ToggleRepeat() {
	p.Repeat = !p.Repeat
}

// toggle shuffle on/off
func (p *Player) ToggleShuffle() {
	p.Shuffle = !p.Shuffle
	// TODO: actually shuffle the list
}

// search for songs
func (p *Player) Search(query string) {
	p.SearchQuery = query
	if query == "" {
		p.FilteredTracks = p.Tracks
		return
	}

	p.FilteredTracks = []Track{}
	query = strings.ToLower(query)

	for _, track := range p.Tracks {
		if strings.Contains(strings.ToLower(track.Title), query) ||
			strings.Contains(strings.ToLower(track.Artist), query) ||
			strings.Contains(strings.ToLower(track.Album), query) {
			p.FilteredTracks = append(p.FilteredTracks, track)
		}
	}
}

// get the current song
func (p *Player) GetCurrentTrack() *Track {
	if len(p.FilteredTracks) == 0 || p.CurrentTrack >= len(p.FilteredTracks) {
		return nil
	}
	return &p.FilteredTracks[p.CurrentTrack]
}

// how far into the song we are
func (p *Player) GetPosition() time.Duration {
	if p.Streamer == nil {
		return 0
	}
	return p.Format.SampleRate.D(p.Streamer.Position())
}

// how long the song is
func (p *Player) GetLength() time.Duration {
	if p.Streamer == nil {
		return 0
	}
	return p.Format.SampleRate.D(p.Streamer.Len())
}

// load a music file
func (p *Player) loadTrack(path string) (beep.StreamSeeker, beep.Format, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, beep.Format{}, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3":
		streamer, format, err := mp3.Decode(file)
		return streamer, format, err
	case ".wav":
		streamer, format, err := wav.Decode(file)
		return streamer, format, err
	default:
		// try mp3 for everything else
		streamer, format, err := mp3.Decode(file)
		return streamer, format, err
	}
}
