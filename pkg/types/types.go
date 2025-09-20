// Package types contains all the type definitions used in the edge-tts-go project.
package types

// TTSConfig represents the internal TTS configuration for edge-tts-go's Communicate struct.
type TTSConfig struct {
	Voice    string
	Rate     string
	Volume   string
	Pitch    string
	Boundary string // "WordBoundary" or "SentenceBoundary"
}

// TTSChunk represents a chunk of data from the TTS service.
type TTSChunk struct {
	Type     string // "audio", "WordBoundary", or "SentenceBoundary"
	Data     []byte // only for audio
	Duration float64 // only for WordBoundary and SentenceBoundary
	Offset   float64 // only for WordBoundary and SentenceBoundary
	Text     string  // only for WordBoundary and SentenceBoundary
}

// VoiceTag represents the voice tag data.
type VoiceTag struct {
	ContentCategories  []string
	VoicePersonalities []string
}

// Voice represents a voice and its attributes.
type Voice struct {
	Name           string
	ShortName      string
	Gender         string // "Female" or "Male"
	Locale         string
	SuggestedCodec string
	FriendlyName   string
	Status         string
	VoiceTag       VoiceTag
}

// VoicesManagerVoice represents a voice for the VoicesManager.
type VoicesManagerVoice struct {
	Voice
	Language string
}

// CommunicateState represents the state of the Communicate struct.
type CommunicateState struct {
	PartialText        []byte
	OffsetCompensation float64
	LastDurationOffset float64
	StreamWasCalled    bool
}

// VoicesManagerFind represents the criteria for finding voices.
type VoicesManagerFind struct {
	Gender   string
	Locale   string
	Language string
}
