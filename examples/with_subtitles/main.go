// This is an example of how to use the edge-tts-go library with subtitles.
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
	text := "Hello, World! This is an example of how to use the edge-tts-go library with subtitles."

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

	// Create a SubMaker for subtitles
	subMaker := submaker.NewSubMaker()

	// Stream the audio and metadata
	chunkChan, errChan := comm.Stream(ctx)

	// Open the audio file
	audioFile, err := os.Create("output_with_subs.mp3")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating audio file: %v\n", err)
		os.Exit(1)
	}
	defer audioFile.Close()

	// Process the chunks
	for chunk := range chunkChan {
		if chunk.Type == "audio" {
			_, err := audioFile.Write(chunk.Data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing audio: %v\n", err)
				os.Exit(1)
			}
		} else if chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary" {
			subMaker.Feed(chunk)
		}
	}

	// Check for errors
	if err := <-errChan; err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming: %v\n", err)
		os.Exit(1)
	}

	// Save the subtitles
	subFile, err := os.Create("output_with_subs.srt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating subtitle file: %v\n", err)
		os.Exit(1)
	}
	defer subFile.Close()

	_, err = subFile.WriteString(subMaker.GetSRT())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing subtitles: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Audio saved to output_with_subs.mp3")
	fmt.Println("Subtitles saved to output_with_subs.srt")
}
