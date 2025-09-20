# edge-tts-go

`edge-tts-go` is a Go library and command-line tool that allows you to use Microsoft Edge's online text-to-speech service without needing Windows or the Edge browser. This is a Go port of the Python [edge-tts](https://github.com/rany2/edge-tts) package.

## Recent Updates

### v1.0.0 - 2025-09-20

**Major Upgrade**: Synchronized with Python edge-tts v7.2.3

#### 🔄 API & Infrastructure Updates
- **Updated API Endpoints**: Migrated from legacy `speech.platform.bing.com` to new `api.msedgeservices.com/tts/cognitiveservices` endpoints
- **Chromium Version**: Updated from 130.0.2849.68 to 140.0.3485.14 to match latest browser compatibility
- **WebSocket Protocol**: Enhanced WebSocket handling with compression support (equivalent to Python's `compress=15`)
- **DRM Token Generation**: Improved Sec-MS-GEC token generation with clock skew correction

#### 🎯 Feature Enhancements
- **Default Boundary**: Changed default boundary from `WordBoundary` to `SentenceBoundary` to match Python v7.2.3 behavior
- **SubMaker Improvements**: 
  - Enhanced subtitle generation logic to properly handle SentenceBoundary vs WordBoundary
  - Added automatic sorting and re-indexing of subtitle cues
  - Improved timestamp accuracy and formatting
- **VoicesManager**: Updated `Find()` method to use structured `VoicesManagerFind` criteria
- **Sync Streaming**: Added `StreamSync()` method for synchronous streaming operations

#### 🛠️ Technical Improvements
- **Error Handling**: Comprehensive error types matching Python edge-tts exceptions
- **WebSocket Headers**: Cleaned up and optimized WebSocket connection headers
- **Code Quality**: Removed unused imports and improved code consistency
- **Backward Compatibility**: Maintained API compatibility while adding new features

#### 📚 Examples & Documentation
- **New Example**: Added `examples/with_subtitles/main.go` demonstrating subtitle generation
- **Updated Documentation**: Refreshed code examples and usage instructions

#### ✅ Testing & Validation
- All core functionality tested and verified working
- Audio generation, streaming, and subtitle creation confirmed functional
- No breaking changes to existing API usage patterns

## Features

- Convert text to speech using Microsoft Edge's online TTS service
- List available voices
- Stream audio data
- Generate subtitles
- Command-line interface
- Asynchronous API

- Convert text to speech using Microsoft Edge's online TTS service
- List available voices
- Stream audio data
- Generate subtitles
- Command-line interface
- Asynchronous API

## Installation

### Library

```bash
go get github.com/difyz9/edge-tts-go
```

### Command-line tool

```bash
go install github.com/difyz9/edge-tts-go/cmd/edge-tts@latest
```

## Usage

### Command-line

```bash
# Basic usage
edge-tts --text "Hello, World!" --write-media output.mp3

# List available voices
edge-tts --list-voices

# Use a specific voice
edge-tts --text "Hello, World!" --voice en-US-GuyNeural --write-media output.mp3

# Generate subtitles
edge-tts --text "Hello, World!" --write-media output.mp3 --write-subtitles output.srt

# Read text from a file
edge-tts --file input.txt --write-media output.mp3

# Adjust speech parameters
edge-tts --text "Hello, World!" --rate +10% --volume +10% --pitch +10Hz --write-media output.mp3
```

### Library

#### Basic Example

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/difyz9/edge-tts-go/pkg/communicate"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Text to convert to speech
	text := "Hello, World! This is a simple example of how to use the edge-tts-go library."

	// Voice to use
	voice := "en-US-GuyNeural"

	// Create a new Communicate instance
	comm, err := communicate.NewCommunicate(
		text,
		voice,
		"+0%",  // rate
		"+0%",  // volume
		"+0Hz", // pitch
		"",     // proxy
		10,     // connectTimeout
		60,     // receiveTimeout
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Communicate instance: %v\n", err)
		os.Exit(1)
	}

	// Save the audio to a file
	err = comm.Save(ctx, "output.mp3", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving audio: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Audio saved to output.mp3")
}
```

#### Streaming with Subtitles

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/difyz9/edge-tts-go/pkg/communicate"
	"github.com/difyz9/edge-tts-go/pkg/submaker"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Text to convert to speech
	text := "Hello, World! This is an example of how to use the edge-tts-go library with streaming and subtitles."

	// Voice to use
	voice := "en-US-GuyNeural"

	// Create a new Communicate instance
	comm, err := communicate.NewCommunicate(
		text,
		voice,
		"+0%",  // rate
		"+0%",  // volume
		"+0Hz", // pitch
		"",     // proxy
		10,     // connectTimeout
		60,     // receiveTimeout
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Communicate instance: %v\n", err)
		os.Exit(1)
	}

	// Create a SubMaker instance
	sm := submaker.NewSubMaker()

	// Open the output files
	audioFile, err := os.Create("output.mp3")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating audio file: %v\n", err)
		os.Exit(1)
	}
	defer audioFile.Close()

	subFile, err := os.Create("output.srt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating subtitle file: %v\n", err)
		os.Exit(1)
	}
	defer subFile.Close()

	// Stream the audio and metadata
	chunkChan, errChan := comm.Stream(ctx)

	// Process the chunks
	for chunk := range chunkChan {
		if chunk.Type == "audio" {
			_, err := audioFile.Write(chunk.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing audio data: %v\n", err)
				os.Exit(1)
			}
		} else if chunk.Type == "WordBoundary" {
			err := sm.Feed(chunk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error feeding WordBoundary: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// Check for errors
	if err := <-errChan; err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming: %v\n", err)
		os.Exit(1)
	}

	// Merge cues to reduce the number of cues
	err = sm.MergeCues(10) // 10 words per cue
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error merging cues: %v\n", err)
		os.Exit(1)
	}

	// Write the subtitles to the file
	_, err = fmt.Fprint(subFile, sm.GetSRT())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing subtitles: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Audio saved to output.mp3")
	fmt.Println("Subtitles saved to output.srt")
}
```

#### Listing Available Voices

```go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/difyz9/edge-tts-go/pkg/voices"
)

func main() {
	// Create a context
	ctx := context.Background()

	// Get the list of voices
	voiceList, err := voices.ListVoices(ctx, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing voices: %v\n", err)
		os.Exit(1)
	}

	// Print the voices in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tGender\tLocale\tContentCategories\tVoicePersonalities")

	for _, voice := range voiceList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			voice.ShortName,
			voice.Gender,
			voice.Locale,
			strings.Join(voice.VoiceTag.ContentCategories, ", "),
			strings.Join(voice.VoiceTag.VoicePersonalities, ", "),
		)
	}

	w.Flush()
}
```

## License

This project is licensed under the GPL-3.0 License - see the LICENSE file for details.

## Acknowledgements

This project is a Go port of the Python [edge-tts](https://github.com/rany2/edge-tts) package by [rany2](https://github.com/rany2).
