// Package communicate provides the main functionality for communicating with the TTS service.
package communicate

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/difyz9/edge-tts-go/internal/constants"
	"github.com/difyz9/edge-tts-go/internal/websocket"
	"github.com/difyz9/edge-tts-go/pkg/errors"
	"github.com/difyz9/edge-tts-go/pkg/types"
	"github.com/difyz9/edge-tts-go/pkg/util"
)

// Communicate is the main struct for communicating with the TTS service.
type Communicate struct {
	texts          [][]byte
	ttsConfig      types.TTSConfig
	proxy          string
	connectTimeout int
	receiveTimeout int
	state          types.CommunicateState
	mu             sync.Mutex
}

// NewCommunicate creates a new Communicate instance.
func NewCommunicate(
	text string,
	voice string,
	rate string,
	volume string,
	pitch string,
	proxy string,
	connectTimeout int,
	receiveTimeout int,
	boundary ...string,
) (*Communicate, error) {
	// Set default values
	if voice == "" {
		voice = constants.DefaultVoice
	}
	if rate == "" {
		rate = "+0%"
	}
	if volume == "" {
		volume = "+0%"
	}
	if pitch == "" {
		pitch = "+0Hz"
	}
	if connectTimeout <= 0 {
		connectTimeout = 10
	}
	if receiveTimeout <= 0 {
		receiveTimeout = 60
	}

	// Set boundary (default to SentenceBoundary if not provided)
	boundaryValue := "SentenceBoundary"
	if len(boundary) > 0 && boundary[0] != "" {
		boundaryValue = boundary[0]
	}

	// Create and validate TTS config
	ttsConfig := types.TTSConfig{
		Voice:    voice,
		Rate:     rate,
		Volume:   volume,
		Pitch:    pitch,
		Boundary: boundaryValue,
	}
	err := util.ValidateTTSConfig(&ttsConfig)
	if err != nil {
		return nil, err
	}

	// Clean and escape the text
	cleanText := util.RemoveIncompatibleCharacters(text)
	escapedText := util.EscapeXML(cleanText)

	// Split the text into multiple strings
	texts := util.SplitTextByByteLength(escapedText, util.CalcMaxMesgSize(ttsConfig))

	// Create the Communicate instance
	return &Communicate{
		texts:          texts,
		ttsConfig:      ttsConfig,
		proxy:          proxy,
		connectTimeout: connectTimeout,
		receiveTimeout: receiveTimeout,
		state: types.CommunicateState{
			PartialText:        []byte{},
			OffsetCompensation: 0,
			LastDurationOffset: 0,
			StreamWasCalled:    false,
		},
	}, nil
}

// Stream streams audio and metadata from the service.
func (c *Communicate) Stream(ctx context.Context) (<-chan types.TTSChunk, <-chan error) {
	chunkChan := make(chan types.TTSChunk)
	errChan := make(chan error, 1)

	c.mu.Lock()
	// Check if stream was called before
	if c.state.StreamWasCalled {
		c.mu.Unlock()
		errChan <- fmt.Errorf("stream can only be called once")
		close(chunkChan)
		return chunkChan, errChan
	}
	c.state.StreamWasCalled = true
	c.mu.Unlock()

	go func() {
		defer close(chunkChan)
		defer close(errChan)

		// Stream the audio and metadata from the service
		for _, partialText := range c.texts {
			c.mu.Lock()
			c.state.PartialText = partialText
			c.mu.Unlock()

			err := c.streamPartialText(ctx, chunkChan)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	return chunkChan, errChan
}

// streamPartialText streams a partial text to the service.
func (c *Communicate) streamPartialText(ctx context.Context, chunkChan chan<- types.TTSChunk) error {
	// Create a new WebSocket client
	client := websocket.NewClient(c.proxy, c.connectTimeout, c.receiveTimeout)

	// Connect to the service
	err := client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// Send the command request
	err = client.SendCommandRequest(c.ttsConfig)
	if err != nil {
		return err
	}

	// Send the SSML request
	c.mu.Lock()
	err = client.SendSSMLRequest(c.state.PartialText, c.ttsConfig)
	c.mu.Unlock()
	if err != nil {
		return err
	}

	// Receive messages from the service
	audioWasReceived := false
	for {
		chunk, err := client.ReceiveMessage()
		if err != nil {
			return err
		}

		if chunk.Type == "audio" {
			audioWasReceived = true
			chunkChan <- chunk
		} else if chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary" {
			chunkChan <- chunk

			// Update the last duration offset for use by the next SSML request
			c.mu.Lock()
			c.state.LastDurationOffset = chunk.Offset + chunk.Duration
			c.mu.Unlock()
		} else if chunk.Type == "turn.end" {
			// Update the offset compensation for the next SSML request
			c.mu.Lock()
			c.state.OffsetCompensation = c.state.LastDurationOffset

			// Use average padding typically added by the service
			// to the end of the audio data. This seems to work pretty
			// well for now, but we might ultimately need to use a
			// more sophisticated method like using ffmpeg to get
			// the actual duration of the audio data.
			c.state.OffsetCompensation += 8_750_000
			c.mu.Unlock()

			// Exit the loop so we can send the next SSML request
			break
		}
	}

	if !audioWasReceived {
		return errors.NewNoAudioReceivedError("no audio was received. Please verify that your parameters are correct.")
	}

	return nil
}

// Save saves the audio and metadata to the specified files.
func (c *Communicate) Save(ctx context.Context, audioFname string, metadataFname string) error {
	// Open the audio file
	audioFile, err := os.Create(audioFname)
	if err != nil {
		return err
	}
	defer audioFile.Close()

	// Open the metadata file if specified
	var metadataFile *os.File
	if metadataFname != "" {
		metadataFile, err = os.Create(metadataFname)
		if err != nil {
			return err
		}
		defer metadataFile.Close()
	}

	// Stream the audio and metadata
	chunkChan, errChan := c.Stream(ctx)

	// Process the chunks
	for chunk := range chunkChan {
		if chunk.Type == "audio" {
			_, err := audioFile.Write(chunk.Data)
			if err != nil {
				return err
			}
		} else if (chunk.Type == "WordBoundary" || chunk.Type == "SentenceBoundary") && metadataFile != nil {
			// Write the metadata to the file
			_, err := fmt.Fprintf(metadataFile, "Type: %s, Offset: %f, Duration: %f, Text: %s\n",
				chunk.Type, chunk.Offset, chunk.Duration, chunk.Text)
			if err != nil {
				return err
			}
		}
	}

	// Check for errors
	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// StreamSync streams audio and metadata synchronously.
func (c *Communicate) StreamSync() <-chan types.TTSChunk {
	chunkChan := make(chan types.TTSChunk)

	go func() {
		defer close(chunkChan)

		// Get the async stream
		asyncChunkChan, errChan := c.Stream(context.Background())

		// Process chunks
		for {
			select {
			case chunk, ok := <-asyncChunkChan:
				if !ok {
					return
				}
				chunkChan <- chunk
			case err, ok := <-errChan:
				if !ok {
					return
				}
				// In sync version, we ignore errors and just close the channel
				// The caller should handle errors through the async version
				if err != nil {
					close(chunkChan)
					return
				}
			}
		}
	}()

	return chunkChan
}
