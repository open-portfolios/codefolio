# Adapter

Convert the interface of a class into another interface clients expect.

## Intent

- Convert the interface of a class into another interface clients expect
- Allow incompatible interfaces to work together
- Provide a unified interface to a set of interfaces in a subsystem

## Implementation

```python
from abc import ABC, abstractmethod


# Target interface
class MediaPlayer(ABC):
    @abstractmethod
    def play(self, filename: str) -> None:
        pass


# Adaptee
class AdvancedMediaPlayer(object):
    def play_mp4(self, filename: str) -> None:
        print(f"Playing MP4: {filename}")
    
    def play_vlc(self, filename: str) -> None:
        print(f"Playing VLC: {filename}")


# Adapter
class MediaAdapter(MediaPlayer):
    player: AdvancedMediaPlayer
    
    def __init__(self) -> None:
        self.player = AdvancedMediaPlayer()
    
    def play(self, filename: str) -> None:
        if filename.endswith(".mp4"):
            self.player.play_mp4(filename)
        elif filename.endswith(".vlc"):
            self.player.play_vlc(filename)


# Client code
class AudioPlayer(MediaPlayer):
    adapter: MediaAdapter | None
    
    def __init__(self) -> None:
        self.adapter = None
    
    def play(self, filename: str) -> None:
        if filename.endswith(".mp3"):
            print(f"Playing MP3: {filename}")
        else:
            self.adapter = MediaAdapter()
            self.adapter.play(filename)


if __name__ == "__main__":
    player = AudioPlayer()
    player.play("song.mp3")
    player.play("movie.mp4")
    player.play("video.vlc")
```

## When to Use

- When you want to use an existing class but its interface is incompatible
- When you need to integrate third-party libraries with different interfaces
- When you want to create a reusable class that cooperates with unrelated classes
