// utility functions
package grabber

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ReencodeForIOS takes input and output file paths and re-encodes the video to be iOS/Telegram-compatible.
func ReencodeForTelegramIOS(path string) error {
	// Ensure ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %v", err)
	}

	// Create a temporary output file
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	name := base[:len(base)-len(ext)]
	tempPath := filepath.Join(dir, name+"_temp"+ext)

	// Defer removal of the temporary file
	defer func() {
		if err := os.Remove(tempPath); err != nil {
			fmt.Printf("[ERR]: failed to remove temporary file '%s' (%v)\n", tempPath, err)
		} else {
			fmt.Printf("[INFO]: successfully removed temporary file '%s'\n", tempPath)
		}
	}()

	// Construct ffmpeg command
	cmd := exec.Command(
		"ffmpeg",
		"-i", path,
		"-c:v", "libx264",
		"-profile:v", "baseline",
		"-level", "3.0",
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-movflags", "+faststart",
		tempPath,
	)

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg re-encode failed: %v", err)
	}

	// Replace original file with re-encoded version
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("failed to replace original file: %v", err)
	}

	return nil
}
