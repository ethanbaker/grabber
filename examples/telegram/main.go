package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/ethanbaker/grabber"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

/* -------- CONSTANTS -------- */

const DOWNLOAD_DIR = "media/"

/* -------- GLOBALS -------- */

var urlRegex = regexp.MustCompile(`https?://[^\s/$.?#].[^\s]*`)

var bot *tgbotapi.BotAPI
var mediaGrabber *grabber.Grabber

/* -------- FUNCTIONS -------- */

// Send media to the user from a given URL
func sendMediaMessage(rawUrl string, update tgbotapi.Update) {
	// Identify platform url belongs to
	platform := mediaGrabber.IdentifyPlatform(rawUrl)

	// If platform is unknown, send error message
	if platform == grabber.Unknown {
		log.Printf("[INFO]: url is from an unsupported platform\n")

		// If bot is in a group fail silently
		if update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup() {
			log.Printf("[INFO]: bot is in a group, failing silently\n")
			return
		}

		// If bot is in a private chat, send error message
		reply := "Sorry, I can't download from this URL"
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
			log.Printf("[INFO]: failed to delete processing message with id '%d' (%v)\n", processingMsg.MessageID, err)
			return
		}

		log.Printf("[INFO]: successfully deleted processing message with id '%d'\n", processingMsg.MessageID)
	}(update.Message.Chat.ID, processingMsg.MessageID)

	// Create a unique download path for this specific request to avoid collisions
	name := fmt.Sprintf("chat%d_msg%d_ts%d", update.Message.Chat.ID, update.Message.MessageID, update.Message.Date)

	// Download media using the grabber library
	results, err := mediaGrabber.Download(rawUrl, name)

	// Defer temporary file removal
	defer func(path string) {
		log.Printf("[INFO]: removing download folder '%s'\n", path)

		if err := os.RemoveAll(path); err != nil {
			log.Printf("[ERR]: failed to remove download folder '%s' (%v)\n", path, err)
		}
	}(name)

	// If an error is present from downloading, return error message
	if err != nil {
		log.Printf("[ERR]: error downloading media from '%s' (%v)\n", rawUrl, err)

		// Generate error message depending on error
		errorMsg := ""
		if strings.Contains(err.Error(), "Unsupported URL") || strings.Contains(err.Error(), "Unable to extract") {
			errorMsg = "The URL might be unsupported, private, or the content was removed"
		} else if strings.Contains(err.Error(), "Video unavailable") {
			errorMsg = "This video is unavailable"
		} else {
			errorMsg = "I'm unable to fetch the requested media" // Generic error
		}

		// Send error message reply
		reply := fmt.Sprintf("Sorry, I couldn't fetch the media from the provided URL. %s", errorMsg)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)

		return
	}

	// If results are empty but no error was returned, return generic error
	if len(results) == 0 {
		log.Printf("No files returned from download for url '%s' (but no error reported)\n", rawUrl)

		reply := "Sorry, I couldn't find the media link for the provided URL"
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)

		return
	}

	// We successfully downloaded, so we can return media
	log.Printf("[INFO]: successfully downloaded '%s'\n", rawUrl)

	// Truncate extra files depending on the platform
	switch platform {
	case grabber.YouTube: // For YouTube, we only want the main video file
		filter(&results, func(r grabber.FileResult) bool {
			return r.MediaType == grabber.Video
		})

	case grabber.Tiktok: // For TikTok, we only want the main video file
		filter(&results, func(r grabber.FileResult) bool {
			return r.MediaType == grabber.Video
		})

	case grabber.InstagramReels: // For Instagram Reels, we only want the main video file
		filter(&results, func(r grabber.FileResult) bool {
			return r.MediaType == grabber.Video
		})

		// Reencode videos for Telegram iOS
		for i := range results {
			if err := grabber.ReencodeForTelegramIOS(results[i].Path); err != nil {
				log.Printf("[ERR]: failed to re-encode video for iOS, failing silently (%v)\n", err)
			} else {
				log.Printf("[INFO]: successfully re-encoded video for iOS compatibility\n")
			}
		}
	}

	// Send each file (in case there are multiple files from the download)
	for _, result := range results {
		var sendable tgbotapi.Chattable
		fileToSend := tgbotapi.FilePath(result.Path)

		// Switch media type to create a telegram message
		switch result.MediaType {
		case grabber.Photo:
			log.Printf("[INFO]: sending image\n")

			photoMsg := tgbotapi.NewPhoto(update.Message.Chat.ID, fileToSend)
			photoMsg.ReplyToMessageID = update.Message.MessageID
			sendable = photoMsg

		case grabber.Video:
			log.Printf("[INFO]: sending video\n")

			videoMsg := tgbotapi.NewVideo(update.Message.Chat.ID, fileToSend)
			videoMsg.ReplyToMessageID = update.Message.MessageID
			videoMsg.SupportsStreaming = true
			sendable = videoMsg

		case grabber.Animation:
			log.Printf("[INFO]: sending animation\n")

			animMsg := tgbotapi.NewAnimation(update.Message.Chat.ID, fileToSend)
			animMsg.ReplyToMessageID = update.Message.MessageID
			sendable = animMsg

		default: // Send as document if type is unknown or not directly supported as photo/video
			log.Printf("[INFO]: media type '%s' for %s is unknown or not photo/video, sending as document\n", result.MediaType, result.Path)

			docMsg := tgbotapi.NewDocument(update.Message.Chat.ID, fileToSend)
			docMsg.ReplyToMessageID = update.Message.MessageID
			sendable = docMsg
		}

		// Send the content to the user
		if _, err := bot.Send(sendable); err != nil {
			log.Printf("[ERR]: failed to send media '%s' to chat '%d' (%v)\n", result.Path, update.Message.Chat.ID, err)

			reply := fmt.Sprintf("I downloaded the media from %s but failed to send it to you. Telegram API error: %s\n", rawUrl, err.Error())
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
			msg.ReplyToMessageID = update.Message.MessageID
			bot.Send(msg)
		} else {
			log.Printf("[INFO]: successfully sent media from '%s' to chat '%d'\n", rawUrl, update.Message.Chat.ID)
		}
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

	// Initialize the grabber with default options
	mediaGrabber = grabber.New().WithBaseDir(DOWNLOAD_DIR).WithRecodeVideo("mp4")

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

		log.Printf("[INFO]: message from '%s' - '%s'\n", update.Message.From.UserName, update.Message.Text)

		// Capture urls and download media
		for _, rawUrl := range foundUrls {
			sendMediaMessage(rawUrl, update)
		}
	}
}

// filter filters the results based on a condition
func filter(results *[]grabber.FileResult, condition func(grabber.FileResult) bool) {
	filtered := make([]grabber.FileResult, 0, len(*results))

	for _, r := range *results {
		if condition(r) {
			filtered = append(filtered, r)
		}
	}

	*results = filtered
}
