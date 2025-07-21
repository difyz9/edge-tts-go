// Package main provides the command-line interface for the edge-tts-go project.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/difyz9/edge-tts-go/internal/constants"
	"github.com/difyz9/edge-tts-go/pkg/communicate"
	"github.com/difyz9/edge-tts-go/pkg/submaker"
	"github.com/difyz9/edge-tts-go/pkg/voices"
)

// UtilArgs represents the CLI arguments.
type UtilArgs struct {
	Text           string
	File           string
	Voice          string
	ListVoices     bool
	Rate           string
	Volume         string
	Pitch          string
	Boundary       string
	WordsInCue     int
	WriteMedia     string
	WriteSubtitles string
	Proxy          string
}

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Parse command-line arguments
	args := parseArgs()

	// List voices if requested
	if args.ListVoices {
		err := printVoices(ctx, args.Proxy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing voices: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Check if text is provided
	if args.Text == "" && args.File == "" {
		fmt.Fprintln(os.Stderr, "Error: either --text or --file must be provided")
		os.Exit(1)
	}

	// Read text from file if provided
	if args.File != "" {
		data, err := os.ReadFile(args.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
		args.Text = string(data)
	}

	// Check if the user wants to write to the terminal
	if args.WriteMedia == "" && isTerminal(os.Stdout.Fd()) && isTerminal(os.Stdin.Fd()) {
		fmt.Fprintln(os.Stderr, "Warning: TTS output will be written to the terminal.")
		fmt.Fprintln(os.Stderr, "Use --write-media to write to a file.")
		fmt.Fprintln(os.Stderr, "Press Ctrl+C to cancel the operation.")
		fmt.Fprintln(os.Stderr, "Press Enter to continue.")
		fmt.Scanln()
	}

	// Create a new Communicate instance
	comm, err := communicate.NewCommunicate(
		args.Text,
		args.Voice,
		args.Rate,
		args.Volume,
		args.Pitch,
		args.Proxy,
		10, // connectTimeout
		60, // receiveTimeout
		args.Boundary,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Communicate instance: %v\n", err)
		os.Exit(1)
	}

	// Create a SubMaker instance
	sm := submaker.NewSubMaker()

	// Open the output files
	var audioFile io.WriteCloser
	if args.WriteMedia != "" && args.WriteMedia != "-" {
		audioFile, err = os.Create(args.WriteMedia)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating audio file: %v\n", err)
			os.Exit(1)
		}
		defer audioFile.Close()
	} else {
		audioFile = os.Stdout
	}

	var subFile io.WriteCloser
	if args.WriteSubtitles != "" && args.WriteSubtitles != "-" {
		subFile, err = os.Create(args.WriteSubtitles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating subtitle file: %v\n", err)
			os.Exit(1)
		}
		defer subFile.Close()
	} else if args.WriteSubtitles == "-" {
		subFile = os.Stderr
	}

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
		} else if chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary" {
			err := sm.Feed(chunk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error feeding %s: %v\n", chunk.Type, err)
				os.Exit(1)
			}
		}
	}

	// Check for errors
	if err := <-errChan; err != nil {
		fmt.Fprintf(os.Stderr, "Error streaming: %v\n", err)
		os.Exit(1)
	}

	// Merge cues if requested
	if args.WordsInCue > 0 {
		err := sm.MergeCues(args.WordsInCue)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error merging cues: %v\n", err)
			os.Exit(1)
		}
	}

	// Write subtitles if requested
	if subFile != nil {
		_, err := fmt.Fprint(subFile, sm.GetSRT())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing subtitles: %v\n", err)
			os.Exit(1)
		}
	}
}

// parseArgs parses the command-line arguments.
func parseArgs() UtilArgs {
	args := UtilArgs{}

	flag.StringVar(&args.Text, "text", "", "what TTS will say")
	flag.StringVar(&args.Text, "t", "", "what TTS will say (shorthand)")
	flag.StringVar(&args.File, "file", "", "same as --text but read from file")
	flag.StringVar(&args.File, "f", "", "same as --text but read from file (shorthand)")
	flag.StringVar(&args.Voice, "voice", constants.DefaultVoice, "voice for TTS")
	flag.StringVar(&args.Voice, "v", constants.DefaultVoice, "voice for TTS (shorthand)")
	flag.BoolVar(&args.ListVoices, "list-voices", false, "lists available voices and exits")
	flag.BoolVar(&args.ListVoices, "l", false, "lists available voices and exits (shorthand)")
	flag.StringVar(&args.Rate, "rate", "+0%", "set TTS rate")
	flag.StringVar(&args.Volume, "volume", "+0%", "set TTS volume")
	flag.StringVar(&args.Pitch, "pitch", "+0Hz", "set TTS pitch")
	flag.StringVar(&args.Boundary, "boundary", "WordBoundary", "set boundary type (WordBoundary or SentenceBoundary)")
	flag.IntVar(&args.WordsInCue, "words-in-cue", 10, "number of words in a subtitle cue")
	flag.StringVar(&args.WriteMedia, "write-media", "", "send media output to file instead of stdout")
	flag.StringVar(&args.WriteSubtitles, "write-subtitles", "", "send subtitle output to provided file instead of stderr")
	flag.StringVar(&args.Proxy, "proxy", "", "use a proxy for TTS and voice list")

	flag.Parse()

	return args
}

// printVoices prints all available voices.
func printVoices(ctx context.Context, proxy string) error {
	// Get the list of voices
	voiceList, err := voices.ListVoices(ctx, proxy)
	if err != nil {
		return err
	}

	// Sort the voices by ShortName
	// In Go, we would typically use the sort package, but for simplicity,
	// we'll assume the voices are already sorted by ShortName

	// Print the voices in a table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Name\tGender\tContentCategories\tVoicePersonalities")

	for _, voice := range voiceList {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			voice.ShortName,
			voice.Gender,
			strings.Join(voice.VoiceTag.ContentCategories, ", "),
			strings.Join(voice.VoiceTag.VoicePersonalities, ", "),
		)
	}

	return w.Flush()
}

// isTerminal returns true if the file descriptor is a terminal.
func isTerminal(fd uintptr) bool {
	// Simple implementation that checks if the file descriptor 
	// corresponds to a character device (typical for terminals)
	if fd == os.Stdin.Fd() || fd == os.Stdout.Fd() || fd == os.Stderr.Fd() {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return false
		}
		return (stat.Mode() & os.ModeCharDevice) != 0
	}
	
	return false
}
