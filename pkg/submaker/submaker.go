// Package submaker is used to generate subtitles from WordBoundary and SentenceBoundary events.
package submaker

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/difyz9/edge-tts-go/pkg/types"
)

// SubMaker is used to generate subtitles from WordBoundary and SentenceBoundary messages.
type SubMaker struct {
	subtitles []Subtitle
	current   *Subtitle
	index    int
}

// Subtitle represents a subtitle cue.
type Subtitle struct {
	Index   int
	Start   time.Duration
	End     time.Duration
	Content string
}

// NewSubMaker creates a new SubMaker.
func NewSubMaker() *SubMaker {
	return &SubMaker{
		subtitles: make([]Subtitle, 0),
		index:     1,
	}
}

// Feed feeds a WordBoundary or SentenceBoundary message to the SubMaker.
func (sm *SubMaker) Feed(msg types.TTSChunk) error {
	if msg.Type != "WordBoundary" && msg.Type != "SentenceBoundary" {
		return fmt.Errorf("invalid message type, expected 'WordBoundary' or 'SentenceBoundary', got '%s'", msg.Type)
	}

	offset := time.Duration(msg.Offset * float64(time.Nanosecond))
	duration := time.Duration(msg.Duration * float64(time.Nanosecond))

	if sm.current == nil {
		sm.current = &Subtitle{
			Index:   sm.index,
			Start:   offset,
			Content: msg.Text,
		}
		sm.index++
	} else {
		sm.current.End = offset + duration
		sm.current.Content += msg.Text
	}

	// For SentenceBoundary, finalize the current subtitle
	if msg.Type == "SentenceBoundary" {
		sm.subtitles = append(sm.subtitles, *sm.current)
		sm.current = nil
	}

	return nil
}

// MergeCues merges cues to reduce the number of cues.
func (sm *SubMaker) MergeCues(words int) error {
	if words <= 0 {
		return fmt.Errorf("invalid number of words to merge, expected > 0, got %d", words)
	}

	if len(sm.subtitles) == 0 {
		return nil
	}

	newCues := []Subtitle{}
	currentCue := sm.subtitles[0]

	for _, cue := range sm.subtitles[1:] {
		if len(strings.Fields(currentCue.Content)) < words {
			currentCue.End = cue.End
			currentCue.Content = currentCue.Content + " " + cue.Content
		} else {
			newCues = append(newCues, currentCue)
			currentCue = cue
		}
	}

	newCues = append(newCues, currentCue)

	// Update indices
	for i := range newCues {
		newCues[i].Index = i + 1
	}

	sm.subtitles = newCues
	return nil
}

// GetSRT returns the SRT formatted subtitles from the SubMaker.
func (sm *SubMaker) GetSRT() string {
	// Finalize any pending subtitle
	if sm.current != nil {
		sm.subtitles = append(sm.subtitles, *sm.current)
		sm.current = nil
	}

	// Sort subtitles by start time
	sort.Slice(sm.subtitles, func(i, j int) bool {
		return sm.subtitles[i].Start < sm.subtitles[j].Start
	})

	// Reindex
	for i := range sm.subtitles {
		sm.subtitles[i].Index = i + 1
	}

	var sb strings.Builder

	for _, cue := range sm.subtitles {
		// Format: "00:00:00,000 --> 00:00:00,000"
		startStr := formatDuration(cue.Start)
		endStr := formatDuration(cue.End)

		sb.WriteString(fmt.Sprintf("%d\n", cue.Index))
		sb.WriteString(fmt.Sprintf("%s --> %s\n", startStr, endStr))
		sb.WriteString(fmt.Sprintf("%s\n\n", cue.Content))
	}

	return sb.String()
}

// formatDuration formats a duration as "00:00:00,000".
func formatDuration(d time.Duration) string {
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	d -= s * time.Second
	ms := d / time.Millisecond

	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms)
}

// String returns the SRT formatted subtitles from the SubMaker.
func (sm *SubMaker) String() string {
	return sm.GetSRT()
}
