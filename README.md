# TuiTunes

A simple terminal music player I made because I wanted something lightweight to play music in the terminal.

## Features

- Play music files (mp3, wav, flac, m4a, aac, ogg)
- Basic controls (play, pause, next, previous)
- Search through your music
- Repeat and shuffle (shuffle doesn't work yet)
- Works with any folder structure

## Installation

1. Make sure you have Go installed
2. Clone this repo
3. Run `go build`
4. Use the `tuitunes` binary

## Usage

```bash
# use current folder
./tuitunes

# use specific folder
./tuitunes ~/Music

# show help
./tuitunes -help
```

## Controls

- `space` - play/pause
- `n` - next song
- `p` - previous song
- `r` - repeat on/off
- `s` - shuffle on/off (not implemented yet)
- `/` - search
- `h` - help
- `q` - quit

## Navigation

- `up/down` or `j/k` - move around
- `g` - go to top
- `G` - go to bottom
- `enter` - play selected song

## How it works

It scans whatever folder you give it for music files and builds a playlist. The UI is built with Bubble Tea which makes it easy to create terminal interfaces.

The audio is handled by the beep library which supports most common formats.

## Notes

- Volume control is handled by your system
- It tries to guess artist/album from folder names
- Search looks through song titles, artists, and albums
- Repeat mode loops the playlist
- Shuffle is planned but not done yet

## License

MIT