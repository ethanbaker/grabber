<!--
  Created by: Ethan Baker (contact@ethanbaker.dev)
  
  Adapted from:
    https://github.com/othneildrew/Best-README-Template/
-->

<div id="top"></div>


<!-- PROJECT SHIELDS/BUTTONS -->
[![GoDoc](https://godoc.org/github.com/ethanbaker/grabber?status.svg)](https://godoc.org/github.com/ethanbaker/grabber)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanbaker/grabber)](https://goreportcard.com/report/github.com/ethanbaker/grabber)
![v1.0.0](https://img.shields.io/badge/status-v1.0.0-green)
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]


<!-- PROJECT LOGO -->
<br><br><br>
<div align="center">
  <h3 align="center">Grabber</h3>

  <p align="center">
    A Go library for downloading media from various platforms using yt-dlp
  </p>
</div>


<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>


<!-- ABOUT -->
## About

Grabber is a Go library that provides a simple and unified interface for downloading media from various social media platforms including TikTok, Instagram, Twitter/X, and YouTube. Built on top of the powerful `yt-dlp` tool, it offers a clean Go API with configurable options for download behavior, platform detection, and support for both photos and videos.

Key features include:
- Generic interface for downloading media from supported platforms
- Configurable options for download behavior
- Automatic platform detection
- Support for photos, videos, and GIFs
- Built on top of `github.com/lrstanley/go-ytdlp`

<p align="right">(<a href="#top">back to top</a>)</p>


### Built With

* [Go](https://golang.org/)
* [yt-dlp](https://github.com/yt-dlp/yt-dlp)
* [go-ytdlp](https://github.com/lrstanley/go-ytdlp)

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- GETTING STARTED -->
## Getting Started

To get started with Grabber, you'll need to have Go installed on your system and ensure that the required dependencies are available.


### Prerequisites

- Go 1.24.3 or later
- `yt-dlp` (automatically handled by the library)
- Optional: `ffmpeg` (required for video processing)

### Installation

Install the library using Go modules:

```bash
go get github.com/ethanbaker/grabber
```

Other prerequisites, such as `ffmpeg`, can be installed for specific post processing/reencoding tasks. For example, `ffmpeg` is used to reencode `mp4` videos being sent to iOS from telegram. To install `ffmpeg` for linux, run:

```
sudo apt install ffmpeg
```

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- USAGE EXAMPLES -->
## Usage

### Basic Usage

Here's a simple example of how to use Grabber to download media:

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/ethanbaker/grabber"
)

func main() {
    // Create a grabber with default options
    g := grabber.New()

    // Check if URL was provided as argument
    if len(os.Args) < 2 {
        log.Fatal("Usage: go run main.go <url>")
    }

    url := os.Args[1]

    // Identify the platform
    platform := g.IdentifyPlatform(url)
    if platform == grabber.Unknown {
        log.Fatal("Unsupported platform for URL:", url)
    }

    fmt.Printf("Detected platform: %s\n", platform)
    fmt.Printf("Downloading from: %s\n", url)

    // Download the media
    name := fmt.Sprintf("download-%s", time.Now().Format("20060102_150405"))
    result, err := g.Download(url, name)
    if err != nil {
        log.Fatalf("Download failed: %v", err)
    }

    // Print the results
    for _, file := range result {
        fmt.Printf("Successfully downloaded file %s!\n", file.Name)
        fmt.Printf("\tPath: %s\n", file.Path)
        fmt.Printf("\tType: %s\n", file.MediaType)
    }

    // Perform post download operations based on platform
    // ...
}
```

### Custom Configuration

You can customize the download behavior by providing custom options:

```go
package main

import (
    "github.com/ethanbaker/grabber"
)

func main() {
    // Create grabber with a few tweaked options
    g := grabber.New().WithBaseDir("downloads/").
            WithMaxFileSize("50M").
            WithNoPlaylist(true)

    // ----------------------------------------

    // Create a grabber with fully custom options
    options := &grabber.Options{
        BaseDir:           "./my-downloads/",
        MaxFileSize:       "100M",
        Format:            "best",
        RestrictFilenames: true,
        NoPlaylist:        true,
        WriteThumbnail:    false,
        OutputTemplate:    "%(title)s.%(ext)s",
    }
    
    g := grabber.New().WithOptions(options)
}
```

The default options for a grabber are:

```go
&grabber.Options{
    BaseDir:           "media/",
    MaxFileSize:       "49M",
    Format:            "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
    RestrictFilenames: true,
    NoPlaylist:        true,
    WriteThumbnail:    true,
    RecodeVideo:       "",
    OutputTemplate:    "media.%(ext)s",
}
```

### Telegram Bot Integration

The library includes an example Telegram bot implementation that demonstrates how to integrate Grabber into a messaging application. The bot can receive URLs from users and automatically download and send the media back. This is particularly useful for creating bots that can archive or redistribute content from various social media platforms.

To run the Telegram bot example:
1. Create a bot token through [@BotFather](https://t.me/botfather)
2. Set up your environment with the bot token
3. Run the bot and send it URLs from supported platforms

### Supported Platforms

- TikTok
- Instagram Reels
- Twitter/X
- YouTube

_For more examples, please refer to the [documentation][documentation-url]._

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ROADMAP -->
## Roadmap

- [x] Core downloading functionality
- [x] Platform detection
- [x] Configurable options
- [x] Telegram bot example
- [ ] Additional platform support
- [ ] Discord implementation

See the [open issues][issues-url] for a full list of proposed features (and known issues).

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTRIBUTING -->
## Contributing

For issues and suggestions, please include as much useful information as possible.
Review the [documentation][documentation-url] and make sure the issue is actually
present or the suggestion is not included. Please share issues/suggestions on the
[issue tracker][issues-url].

For patches and feature additions, please submit them as [pull requests][pulls-url]. 
Please adhere to the [conventional commits][conventional-commits-url] standard for
commit messaging. In addition, please try to name your git branch according to your
new patch. [These standards][conventional-branches-url] are a great guide you can follow.

You can follow these steps below to create a pull request:

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/amazing-feature`)
3. Commit your Changes (`git commit -m "feat: add amazing feature"`)
4. Push to the Branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- LICENSE -->
## License

This project is licensed under the MIT License - see the [LICENSE][license-url] file for details.

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- CONTACT -->
## Contact

Ethan Baker - contact@ethanbaker.dev - [LinkedIn][linkedin-url]

Project Link: [https://github.com/ethanbaker/grabber][project-url]

<p align="right">(<a href="#top">back to top</a>)</p>


<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

* [yt-dlp](https://github.com/yt-dlp/yt-dlp) - The powerful media downloader that powers this library
* [go-ytdlp](https://github.com/lrstanley/go-ytdlp) - Go bindings for yt-dlp

<p align="right">(<a href="#top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/ethanbaker/grabber.svg
[forks-shield]: https://img.shields.io/github/forks/ethanbaker/grabber.svg
[stars-shield]: https://img.shields.io/github/stars/ethanbaker/grabber.svg
[issues-shield]: https://img.shields.io/github/issues/ethanbaker/grabber.svg
[license-shield]: https://img.shields.io/github/license/ethanbaker/grabber.svg
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?logo=linkedin&colorB=555

[contributors-url]: <https://github.com/ethanbaker/grabber/graphs/contributors>
[forks-url]: <https://github.com/ethanbaker/grabber/network/members>
[stars-url]: <https://github.com/ethanbaker/grabber/stargazers>
[issues-url]: <https://github.com/ethanbaker/grabber/issues>
[pulls-url]: <https://github.com/ethanbaker/grabber/pulls>
[license-url]: <https://github.com/ethanbaker/grabber/blob/master/LICENSE>
[linkedin-url]: <https://linkedin.com/in/ethandbaker>
[project-url]: <https://github.com/ethanbaker/grabber>

[documentation-url]: <https://godoc.org/github.com/ethanbaker/grabber>

[conventional-commits-url]: <https://www.conventionalcommits.org/en/v1.0.0/#summary>
[conventional-branches-url]: <https://docs.microsoft.com/en-us/azure/devops/repos/git/git-branching-guidance?view=azure-devops>
