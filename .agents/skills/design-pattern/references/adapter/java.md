# Adapter

Convert the interface of a class into another interface clients expect.

## Intent

- Convert the interface of a class into another interface clients expect
- Allow incompatible interfaces to work together
- Provide a unified interface to a set of interfaces in a subsystem

## Structure

```
Client
    │
    └── Target (interface)
           │
           └── Adapter ──> Adaptee
```

## Implementation

```java
// Target interface
public interface MediaPlayer {
    void play(String filename);
}

// Adaptee
public class AdvancedMediaPlayer {
    public void playMp4(String filename) {
        System.out.println("Playing MP4: " + filename);
    }
    
    public void playVlc(String filename) {
        System.out.println("Playing VLC: " + filename);
    }
}

// Adapter
public class MediaAdapter implements MediaPlayer {
    private AdvancedMediaPlayer advancedPlayer;
    
    public MediaAdapter() {
        this.advancedPlayer = new AdvancedMediaPlayer();
    }
    
    @Override
    public void play(String filename) {
        if (filename.endsWith(".mp4")) {
            advancedPlayer.playMp4(filename);
        } else if (filename.endsWith(".vlc")) {
            advancedPlayer.playVlc(filename);
        }
    }
}

// Client code
public class AudioPlayer implements MediaPlayer {
    private MediaAdapter adapter;
    
    @Override
    public void play(String filename) {
        if (filename.endsWith(".mp3")) {
            System.out.println("Playing MP3: " + filename);
        } else if (filename.endsWith(".mp4") || filename.endsWith(".vlc")) {
            adapter = new MediaAdapter();
            adapter.play(filename);
        } else {
            System.out.println("Invalid media format");
        }
    }
}
```

## When to Use

- When you want to use an existing class but its interface is incompatible
- When you want to create a reusable class that cooperates with unrelated classes
- When you need to integrate third-party libraries with different interfaces
