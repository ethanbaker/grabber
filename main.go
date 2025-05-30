package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

/* -------- CONSTANTS -------- */

const DOWNLOAD_DIR = "./media/"

/* -------- GLOBALS -------- */

var urlRegex = regexp.MustCompile(`https?://[^\s/$.?#].[^\s]*`)

var bot *tgbotapi.BotAPI

/* -------- TYPES -------- */

// Supported platforms
type Platform string

const (
	Tiktok    Platform = "tiktok"
	Instagram Platform = "instagram"
	Twitter   Platform = "twitter"
	Unknown   Platform = "unknown"
)

// identifyPlatform checks the domain of the URL to determine the platform
func identifyPlatform(webUrl string) Platform {
	// Parse provided URL
	parsedUrl, err := url.Parse(webUrl)
	if err != nil {
		log.Printf("[ERR]: error parsing url '%s' (%v)", webUrl, err)
		return Unknown
	}

	// Determine hostname and return corresponding platform
	host := strings.ToLower(parsedUrl.Hostname())
	if strings.Contains(host, "tiktok.com") {
		return Tiktok
	}
	if strings.Contains(host, "instagram.com") {
		return Instagram
	}
	if strings.Contains(host, "twitter.com") || strings.Contains(host, "x.com") {
		return Twitter
	}
	return Unknown
}

// ytDlpDownload function calls yt-dlp to download media
func ytDlpDownload(mediaUrl string, downloadPath string) (filePath string, mediaType string, err error) {
	// Ensure the download directory exists
	err = os.MkdirAll(downloadPath, 0755)
	if err != nil {
		return "", "", fmt.Errorf("failed to create download directory '%s' (%v)", downloadPath, err)
	}

	// yt-dlp command:
	// -o specifies the output template. We use a generic name within the specific downloadPath
	// Using "%(id)s.%(ext)s" or "%(title)s.%(ext)s" is common
	// For simplicity and uniqueness within the `downloadPath` (which is already unique per message)
	outputTemplate := filepath.Join(downloadPath, "media.%(ext)s")

	cmd := exec.Command("yt-dlp",
		"-o", outputTemplate,
		"--restrict-filenames",  // Sanitize filenames further if needed, though our template is simple
		"--no-playlist",         // Download only the single video/image if URL is part of a playlist
		"--max-filesize", "49M", // Telegram bot API limit for files sent by bot is 50MB. Stay slightly under
		"--write-thumbnail",                                                    // Attempt to download thumbnail (useful for video)
		"--format", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best", // Prefer mp4, good for Telegram
		mediaUrl,
	)

	log.Printf("[INFO]: executing command '%s'\n", cmd.String())

	output, err := cmd.CombinedOutput() // Get both stdout and stderr
	if err != nil {
		log.Printf("[ERR]: yt-dlp execution failed for '%s'\n\tOutput: %s", mediaUrl, string(output))
		return "", "", fmt.Errorf("yt-dlp error (%w)\n\tOutput: %s", err, string(output))
	}

	log.Printf("[INFO]: yt-dlp output for '%s'\n\tOutput: %s", mediaUrl, string(output))

	// Find the main downloaded media file(s). yt-dlp might download media and thumbnail
	files, err := os.ReadDir(downloadPath)
	if err != nil {
		return "", "", fmt.Errorf("could not read download directory '%s' (%w)", downloadPath, err)
	}

	var downloadedMediaFile string
	var downloadedThumbnailFile string // Optional

	for _, file := range files {
		if !file.IsDir() && !strings.HasSuffix(file.Name(), ".part") {
			currentFilePath := filepath.Join(downloadPath, file.Name())
			ext := strings.ToLower(filepath.Ext(currentFilePath))

			// Prioritize video/image over other files (like .webp thumbnails if we want jpg/png)
			// This logic can be more sophisticated based on yt-dlp output parsing
			if ext == ".mp4" || ext == ".mkv" || ext == ".webm" || ext == ".mov" { // Video extensions
				downloadedMediaFile = currentFilePath
				break // Found primary video
			} else if ext == ".jpg" || ext == ".jpeg" || ext == ".png" { // Image extensions
				// If it's named 'media.jpg', it's likely the main image or a thumbnail converted by yt-dlp's format option
				// If it's from --write-thumbnail, it might have a different name pattern
				// For simplicity, assume 'media.<ext>' is the primary target
				if strings.HasPrefix(file.Name(), "media.") {
					downloadedMediaFile = currentFilePath
				} else {
					// Could be a thumbnail downloaded by --write-thumbnail
					// We might want to send this separately or use it if the main media is video
					downloadedThumbnailFile = currentFilePath
				}
			}
		}
	}

	// If main media is video, and we have a separate thumbnail, we might use it
	// For now, we just return the primary media file
	_ = downloadedThumbnailFile // Mark as used if not directly returned

	if downloadedMediaFile == "" {
		return "", "", fmt.Errorf("could not find primary downloaded media file from yt-dlp for '%s'\n\tOutput: %s", mediaUrl, string(output))
	}

	// Determine media type by extension of the primary file
	ext := strings.ToLower(filepath.Ext(downloadedMediaFile))
	determinedMediaType := "unknown"
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp": // .webp can be a photo
		determinedMediaType = "photo"
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		determinedMediaType = "video"
	case ".gif": // .gif can be sent as Animation or Video by Telegram
		determinedMediaType = "video" // Sending as video is often more compatible
	}

	if determinedMediaType == "unknown" {
		log.Printf("[INFO]: downloaded file '%s' has unknown media type extension '%s'. Will attempt to send as document", downloadedMediaFile, ext)
	}

	log.Printf("[INFO]: yt-dlp successfully downloaded '%s' (type: %s)", downloadedMediaFile, determinedMediaType)
	return downloadedMediaFile, determinedMediaType, nil
}

// Send media to the user from a given URL
func sendMediaMessage(rawUrl string, update tgbotapi.Update) {
	// Identify platform url belongs to
	platform := identifyPlatform(rawUrl)

	// If platform is unknown, send error message
	if platform == Unknown {
		log.Printf("[INFO]: url is from an unsupported platform\n")

		// If bot is in a group fail silently
		if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			log.Printf("[INFO]: bot is in a group, failing silently\n")
			return
		}

		// If bot is in a private chat, send error message
		reply := "Sorry, I can't download from this URL. I support TikTok, Instagram, and Twitter/X"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)

		return
	}

	// Send processing message when downloading content
	reply := fmt.Sprintf("Found URL from %s. Attempting to fetch media, please wait...", platform)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
	msg.ReplyToMessageID = update.Message.MessageID
	processingMsg, _ := bot.Send(msg)

	// Defer deleting processing message
	defer func(chatID int64, messageID int) {
		delCfg := tgbotapi.NewDeleteMessage(chatID, messageID)

		if _, err := bot.Request(delCfg); err != nil {
			log.Printf("[INFO]: failed to delete processing message with id '%d' (%v)", processingMsg.MessageID, err)
			return
		}

		log.Printf("[INFO]: successfully deleted processing message with id '%d'\n", processingMsg.MessageID)
	}(update.Message.Chat.ID, processingMsg.MessageID)

	// Create a unique download path for this specific request to avoid collisions
	filename := fmt.Sprintf("chat%d_msg%d_ts%d", update.Message.Chat.ID, update.Message.MessageID, update.Message.Date)
	downloadPath := filepath.Join(DOWNLOAD_DIR, filename)

	if downloadPath == "" {
		log.Printf("[INFO]: failed to create download path\n")

		// If download path is empty, send error message
		reply := "Sorry, I can't download from this URL"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		bot.Send(msg)

		return
	}

	// Download media
	filePath, mediaType, err := ytDlpDownload(rawUrl, downloadPath)

	// Defer temporary file removal
	defer func(path string) {
		log.Printf("[INFO]: removing download folder '%s'\n", path)

		if err := os.RemoveAll(path); err != nil {
			log.Printf("[ERR]: failed to remove download folder '%s' (%v)\n", path, err)
		}
	}(downloadPath)

	// If an error is present from downloading, return error message
	if err != nil {
		log.Printf("[ERR]: error downloading media from '%s' (%v)", rawUrl, err)

		// Generate error message depending on ytDlp error
		errorMsg := ""
		if strings.Contains(err.Error(), "Unsupported URL") || strings.Contains(err.Error(), "Unable to extract") {
			errorMsg = "The URL might be unsupported, private, or the content was removed"
		} else if strings.Contains(err.Error(), "Video unavailable") {
			errorMsg = "This video is unavailable"
		} else {
			errorMsg = "Unable to fetch media" // Generic error
		}

		// Send error message reply
		reply := fmt.Sprintf("Sorry, I couldn't fetch the media from the provided URL.\nError: %s", errorMsg)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)

		return
	}

	// If filepath is empty but no ytDlp error was returned, return generic error
	if filePath == "" {
		log.Printf("No file path returned from download for url '%s' (but no error reported)", rawUrl)

		reply := "Sorry, I couldn't find the media link for the provided URL"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)

		return
	}

	// We successfully downloaded, so we can return media
	log.Printf("[INFO]: successfully downloaded '%s'\n", rawUrl)

	var sendable tgbotapi.Chattable
	fileToSend := tgbotapi.FilePath(filePath)

	// Switch media type to create a telegram message
	switch mediaType {
	case "photo":
		log.Printf("[INFO]: sending image")

		photoMsg := tgbotapi.NewPhoto(update.Message.Chat.ID, fileToSend)
		photoMsg.ReplyToMessageID = update.Message.MessageID
		sendable = photoMsg

	case "video":
		log.Printf("[INFO]: sending video")

		videoMsg := tgbotapi.NewVideo(update.Message.Chat.ID, fileToSend)
		videoMsg.ReplyToMessageID = update.Message.MessageID
		videoMsg.SupportsStreaming = true
		sendable = videoMsg

	default: // Send as document if type is unknown or not directly supported as photo/video
		log.Printf("[INFO]: media type '%s' for %s is unknown or not photo/video, sending as document\n", mediaType, filePath)

		docMsg := tgbotapi.NewDocument(update.Message.Chat.ID, fileToSend)
		docMsg.ReplyToMessageID = update.Message.MessageID
		sendable = docMsg
	}

	// Send the content to the user
	if _, err := bot.Send(sendable); err != nil {
		log.Printf("[ERR]: failed to send media '%s' to chat '%d' (%v)\n", filePath, update.Message.Chat.ID, err)

		reply := fmt.Sprintf("I downloaded the media from %s but failed to send it to you. Telegram API error: %s", rawUrl, err.Error())
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	} else {
		log.Printf("[INFO]: successfully sent media from '%s' to chat '%d'\n", rawUrl, update.Message.Chat.ID)
	}
}

/* -------- MAIN -------- */

func main() {
	// Load environment
	err := godotenv.Load()
	if err != nil {
		log.Fatal("[ERR]: error loading .env file, make sure it exists in the current directory")
	}

	// Get telegram bot token
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("[ERR]: TELEGRAM_BOT_TOKEN environment variable not set")
	}

	// Create telegram bot
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("[ERR]: error creating bot api (%v)\n", err)
	}

	bot.Debug = false // Set to false in production for cleaner logs
	log.Printf("[INFO]: authorized on account %s\n", bot.Self.UserName)

	// Initialize bot updates channel
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	// Create download directory
	if err := os.MkdirAll(DOWNLOAD_DIR, 0755); err != nil {
		log.Fatalf("[ERR]: failed to create base download directory (%v)\n", err)
	}

	// Handle updates from the telegram bot
	for update := range updates {
		// Ignore empty messages
		if update.Message == nil || update.Message.Text == "" {
			continue
		}

		// Only analyze messages that contain URLs
		foundUrls := urlRegex.FindAllString(update.Message.Text, -1)
		if len(foundUrls) == 0 {
			continue
		}

		log.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

		// Capture urls and download media
		for _, rawUrl := range foundUrls {
			sendMediaMessage(rawUrl, update)
		}
	}
}
