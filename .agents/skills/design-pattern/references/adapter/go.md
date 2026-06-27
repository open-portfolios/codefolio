# Adapter

Convert the interface of a class into another interface clients expect.

## Intent

- Convert the interface of a class into another interface clients expect
- Allow incompatible interfaces to work together
- Provide a unified interface to a set of interfaces in a subsystem

## Implementation

```go
package main

import "fmt"

// Target interface
type MediaPlayer interface {
	Play(filename string)
}

// Adaptee
type AdvancedMediaPlayer struct{}

func (p *AdvancedMediaPlayer) PlayMp4(filename string) {
	fmt.Println("Playing MP4:", filename)
}

func (p *AdvancedMediaPlayer) PlayVlc(filename string) {
	fmt.Println("Playing VLC:", filename)
}

// Adapter
type MediaAdapter struct {
	player *AdvancedMediaPlayer
}

func NewMediaAdapter() *MediaAdapter {
	return &MediaAdapter{player: &AdvancedMediaPlayer{}}
}

func (a *MediaAdapter) Play(filename string) {
	if len(filename) > 4 {
		if filename[len(filename)-4:] == ".mp4" {
			a.player.PlayMp4(filename)
		} else if filename[len(filename)-4:] == ".vlc" {
			a.player.PlayVlc(filename)
		}
	}
}

// Client code
type AudioPlayer struct {
	adapter *MediaAdapter
}

func NewAudioPlayer() *AudioPlayer {
	return &AudioPlayer{}
}

func (p *AudioPlayer) Play(filename string) {
	if len(filename) > 4 && filename[len(filename)-4:] == ".mp3" {
		fmt.Println("Playing MP3:", filename)
	} else {
		p.adapter = NewMediaAdapter()
		p.adapter.Play(filename)
	}
}

func main() {
	player := NewAudioPlayer()
	player.Play("song.mp3")
	player.Play("movie.mp4")
	player.Play("video.vlc")
}
```

## When to Use

- When you want to use an existing type but its interface is incompatible
- When you need to integrate third-party libraries with different interfaces
- When you want to create a reusable type that cooperates with unrelated types
