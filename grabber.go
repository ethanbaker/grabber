package grabber

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	ytdlp "github.com/lrstanley/go-ytdlp"
)

// Platform represents the supported platforms grabber can download
type Platform string

const (
	Tiktok         Platform = "tiktok"
	InstagramReels Platform = "instagram reels"
	Twitter        Platform = "twitter"
	YouTube        Platform = "youtube"
	Unknown        Platform = "unknown"
)

// MediaType represents the type of media downloaded
type MediaType string

const (
	Photo     MediaType = "photo"
	Video     MediaType = "video"
	Animation MediaType = "animation"
	None      MediaType = "none"
)

// FileResult represents the result of a download operation
type FileResult struct {
	Name      string
	Path      string
	Ext       string
	MediaType MediaType
}

// Options represents configuration options for the Grabber
type Options struct {
	BaseDir           string // BaseDir is the directory where downloaded files will be saved
	MaxFileSize       string // MaxFileSize is the maximum file size to download (e.g., "49M" for 49 megabytes)
	Format            string // Format specifies the format preference for yt-dlp
	RestrictFilenames bool   // RestrictFilenames sanitizes filenames if true
	RecodeVideo       string // RecodeVideo specifies the video format to recode to (e.g., "mp4")
	NoPlaylist        bool   // NoPlaylist downloads only single video/image if URL is part of a playlist
	WriteThumbnail    bool   // WriteThumbnail attempts to download thumbnail if true
	OutputTemplate    string // OutputTemplate specifies the output filename template
}

// Grabber handles media downloading functionality
type Grabber struct {
	options *Options
}

// New creates a new Grabber instance
func New() *Grabber {
	return &Grabber{
		options: defaultOptions(),
	}
}

// IdentifyPlatform checks the domain of the URL to determine the platform
func (g *Grabber) IdentifyPlatform(webUrl string) Platform {
	// Parse provided URL
	parsedUrl, err := url.Parse(webUrl)
	if err != nil {
		return Unknown
	}

	// Determine hostname and return corresponding platform
	host := strings.ToLower(parsedUrl.Hostname())
	path := strings.ToLower(parsedUrl.Path)
	if strings.Contains(host, "tiktok.com") {
		return Tiktok
	}
	if strings.Contains(host, "instagram.com") && strings.HasPrefix(path, "/reel/") {
		return InstagramReels
	}
	if strings.Contains(host, "twitter.com") || strings.Contains(host, "x.com") {
		return Twitter
	}
	if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be") {
		return YouTube
	}
	return Unknown
}

// Download downloads media from the provided URL using yt-dlp
func (g *Grabber) Download(mediaUrl string, name string) ([]FileResult, error) {
	// Build yt-dlp arguments
	outputPath := filepath.Join(g.options.BaseDir, name)
	args := g.buildArgs(outputPath, mediaUrl)

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return nil, fmt.Errorf("could not create download directory '%s': %w", outputPath, err)
	}

	// Execute yt-dlp
	client := ytdlp.New()
	result, err := client.Run(context.Background(), args...)
	if err != nil {
		return nil, fmt.Errorf("yt-dlp error: %w", err)
	}

	// Check if command was successful
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("yt-dlp failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	// Find the main downloaded media file(s)
	files, err := os.ReadDir(outputPath)
	if err != nil {
		return nil, fmt.Errorf("could not read download directory '%s': %w", outputPath, err)
	}

	primaryMediaFiles := []FileResult{}

	for _, file := range files {
		// Skip directories, incomplete files, and files that don't match the output template
		if file.IsDir() || strings.HasSuffix(file.Name(), ".part") {
			continue
		}

		// Construct full file path
		currentFilePath := filepath.Join(outputPath, file.Name())
		ext := strings.ToLower(filepath.Ext(currentFilePath))

		// Find file type
		mt := findFileType(ext)
		if mt == None {
			log.Printf("[INFO]: Skipping file '%s' with unsupported extension '%s'\n", file.Name(), ext)
			continue
		}

		// Create a File instance for the current file
		f := FileResult{
			Name:      file.Name(),
			Path:      currentFilePath,
			Ext:       ext,
			MediaType: mt,
		}

		// Add to primary media files
		primaryMediaFiles = append(primaryMediaFiles, f)
	}

	// If no primary media file was found, return an error
	if len(primaryMediaFiles) == 0 {
		return nil, fmt.Errorf("could not find primary downloaded media file from yt-dlp for '%s'", mediaUrl)
	}

	return primaryMediaFiles, nil
}

// buildArgs is a helper function that builds the arguments for yt-dlp based on the provided options
func (g *Grabber) buildArgs(outputPath string, mediaUrl string) []string {
	args := []string{
		"-o", filepath.Join(outputPath, g.options.OutputTemplate),
	}

	if g.options.RestrictFilenames {
		args = append(args, "--restrict-filenames")
	}

	if g.options.NoPlaylist {
		args = append(args, "--no-playlist")
	}

	if g.options.MaxFileSize != "" {
		args = append(args, "--max-filesize", g.options.MaxFileSize)
	}

	if g.options.WriteThumbnail {
		args = append(args, "--write-thumbnail")
	}

	if g.options.Format != "" {
		args = append(args, "--format", g.options.Format)
	}

	if g.options.RecodeVideo != "" {
		args = append(args, "--recode-video", g.options.RecodeVideo)
	}

	args = append(args, mediaUrl)

	return args
}

// findFileType is a helper function that determines the media type based on the file extension
func findFileType(ext string) MediaType {
	switch ext {
	case ".mp4", ".mkv", ".webm", ".mov":
		return Video
	case ".jpg", ".jpeg", ".png", ".webp":
		return Photo
	case ".gif":
		return Animation
	default:
		return None
	}
}
