# Grabber Library

A Go library for downloading media from various platforms (TikTok, Instagram, Twitter/X, YouTube/YouTube Shorts) using yt-dlp.

## Features

- Generic interface for downloading media from supported platforms
- Configurable options for download behavior
- Platform detection
- Support for photos and videos
- Built on top of `github.com/lrstanley/go-ytdlp`

## Installation

```bash
go get github.com/ethanbaker/grabber
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/ethanbaker/grabber"
)

func main() {
    // Create a grabber with default options
    g := grabber.New()
    
    // Download media from YouTube Shorts
    name := fmt.Sprintf("download-%s", time.Now().Format("20060102_150405"))
    results, err := g.Download("https://www.youtube.com/shorts/dQw4w9WgXcQ", name)
    if err != nil {
        log.Fatal(err)
    }
    
    // Print results
    for _, file := range results {
        fmt.Printf("Downloaded %s (%s)\n", file.Name, file.MediaType)
        fmt.Printf("Path: %s\n", file.Path)
    }
}
```

### Custom Configuration

```go
package main

import (
    "github.com/ethanbaker/grabber"
)

func main() {
    // Create custom options
    options := &grabber.Options{
        BaseDir:           "./my-downloads/",
        MaxFileSize:       "100M",
        Format:            "best",
        RestrictFilenames: true,
        NoPlaylist:        true,
        WriteThumbnail:    false,
        OutputTemplate:    "%(title)s.%(ext)s",
    }
    
    // Create grabber with custom options
    g := grabber.New().WithOptions(options)
    
    // Use the grabber...
}
```

## Supported Platforms

- TikTok
- Instagram  
- Twitter/X
- YouTube/YouTube Shorts

## Examples

See the `examples/` directory for complete implementations:

- `examples/simple/` - Simple command-line tool for downloading media
- `examples/telegram/` - A Telegram bot that downloads and sends media

### Running the Simple Example

```bash
cd examples/simple
go run main.go "https://www.tiktok.com/@example/video/123"
```

### Running the Telegram Example

1. Create a `.env` file in `examples/telegram/` with your bot token:
   ```
   TELEGRAM_BOT_TOKEN=your_bot_token_here
   ```

2. Run the bot:
   ```bash
   cd examples/telegram
   go run main.go
   ```

## API Reference

### Types

#### `Grabber`
The main interface for downloading media.

#### `Options`
Configuration struct for customizing download behavior:

- `BaseDir`: Directory where files will be saved
- `MaxFileSize`: Maximum file size (e.g., "49M")
- `Format`: yt-dlp format string
- `RestrictFilenames`: Sanitize filenames
- `NoPlaylist`: Download single item only
- `WriteThumbnail`: Download thumbnails
- `OutputTemplate`: Filename template

#### `FileResult`
Result of a download operation:

- `Name`: Name of the downloaded file
- `Path`: Full path to the downloaded file
- `Ext`: File extension
- `MediaType`: Type of media ("photo", "video", "gif")

#### `Platform`
Enum for supported platforms:

- `Tiktok`
- `Instagram` 
- `Twitter`
- `YouTube`
- `Unknown`

### Methods

#### `New() *Grabber`
Creates a new Grabber instance with default options.

#### `(g *Grabber) WithOptions(options *Options) *Grabber`
Creates a new Grabber instance with custom options.

#### `(g *Grabber) WithBaseDir(baseDir string) *Grabber`
Sets the base directory for downloads.

#### `(g *Grabber) Download(url, name string) ([]FileResult, error)`
Downloads media from the given URL to the specified directory name.

#### `(g *Grabber) IdentifyPlatform(url string) Platform`
Identifies the platform from a URL.
