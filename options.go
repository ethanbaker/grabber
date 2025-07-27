package grabber

// defaultOptions returns a default configuration
func defaultOptions() *Options {
	return &Options{
		BaseDir:           "media/",
		MaxFileSize:       "49M",
		Format:            "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
		RestrictFilenames: true,
		NoPlaylist:        true,
		WriteThumbnail:    true,
		RecodeVideo:       "",
		OutputTemplate:    "media.%(ext)s",
	}
}

// WithOptions creates a new Grabber instance with the provided options
func (g *Grabber) WithOptions(options *Options) *Grabber {
	g.options = options
	return g
}

// Only set base directory for the grabber
func (g *Grabber) WithBaseDir(baseDir string) *Grabber {
	g.options.BaseDir = baseDir
	return g
}

// WithMaxFileSize sets the maximum file size to download
func (g *Grabber) WithMaxFileSize(size string) *Grabber {
	g.options.MaxFileSize = size
	return g
}

// WithFormat sets the format preference for yt-dlp
func (g *Grabber) WithFormat(format string) *Grabber {
	g.options.Format = format
	return g
}

// WithRestrictFilenames enables or disables filename sanitization
func (g *Grabber) WithRestrictFilenames(restrict bool) *Grabber {
	g.options.RestrictFilenames = restrict
	return g
}

// WithNoPlaylist sets whether to download only a single video/image if the URL is part of a playlist
func (g *Grabber) WithNoPlaylist(noPlaylist bool) *Grabber {
	g.options.NoPlaylist = noPlaylist
	return g
}

// WithWriteThumbnail enables or disables thumbnail downloading
func (g *Grabber) WithWriteThumbnail(write bool) *Grabber {
	g.options.WriteThumbnail = write
	return g
}

// WithRecodeVideo sets the video recoding format
func (g *Grabber) WithRecodeVideo(format string) *Grabber {
	g.options.RecodeVideo = format
	return g
}

// WithOutputTemplate sets the output filename template
func (g *Grabber) WithOutputTemplate(template string) *Grabber {
	g.options.OutputTemplate = template
	return g
}
