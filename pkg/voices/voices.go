// Package voices contains functions to list all available voices and a struct to find the
// correct voice based on their attributes.
package voices

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/difyz9/edge-tts-go/internal/constants"
	"github.com/difyz9/edge-tts-go/internal/drm"
	"github.com/difyz9/edge-tts-go/pkg/types"
)

// ListVoices lists all available voices and their attributes.
func ListVoices(ctx context.Context, proxy string) ([]types.Voice, error) {
	// Create HTTP client
	client := &http.Client{}
	if proxy != "" {
		// In a real implementation, you would set up a proxy here
		// For simplicity, we'll ignore the proxy parameter
	}

	// Create request
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		constants.VoiceList+"&Sec-MS-GEC="+drm.GenerateSecMSGEC()+"&Sec-MS-GEC-Version="+constants.SecMSGECVersion,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range constants.VoiceHeaders {
		req.Header.Set(k, v)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle 403 error (clock skew)
	if resp.StatusCode == http.StatusForbidden {
		err = drm.HandleClientResponseError(resp)
		if err != nil {
			return nil, err
		}

		// Retry the request
		req, err = http.NewRequestWithContext(
			ctx,
			"GET",
			constants.VoiceList+"&Sec-MS-GEC="+drm.GenerateSecMSGEC()+"&Sec-MS-GEC-Version="+constants.SecMSGECVersion,
			nil,
		)
		if err != nil {
			return nil, err
		}

		// Set headers
		for k, v := range constants.VoiceHeaders {
			req.Header.Set(k, v)
		}

		// Send request
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var voices []types.Voice
	err = json.Unmarshal(body, &voices)
	if err != nil {
		return nil, err
	}

	// Clean up voice data
	for i := range voices {
		// Remove leading and trailing whitespace from categories and personalities
		for j, category := range voices[i].VoiceTag.ContentCategories {
			voices[i].VoiceTag.ContentCategories[j] = strings.TrimSpace(category)
		}
		for j, personality := range voices[i].VoiceTag.VoicePersonalities {
			voices[i].VoiceTag.VoicePersonalities[j] = strings.TrimSpace(personality)
		}
	}

	return voices, nil
}

// VoicesManager is a struct to find the correct voice based on their attributes.
type VoicesManager struct {
	voices      []types.VoicesManagerVoice
	calledCreate bool
	mu          sync.RWMutex
}

// NewVoicesManager creates a new VoicesManager.
func NewVoicesManager() *VoicesManager {
	return &VoicesManager{
		voices:      []types.VoicesManagerVoice{},
		calledCreate: false,
	}
}

// Create populates the VoicesManager with all available voices.
func (vm *VoicesManager) Create(ctx context.Context, customVoices []types.Voice, proxy string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	var voices []types.Voice
	var err error

	if customVoices != nil {
		voices = customVoices
	} else {
		voices, err = ListVoices(ctx, proxy)
		if err != nil {
			return err
		}
	}

	// Convert Voice to VoicesManagerVoice
	vm.voices = make([]types.VoicesManagerVoice, len(voices))
	for i, voice := range voices {
		vm.voices[i] = types.VoicesManagerVoice{
			Voice:    voice,
			Language: strings.Split(voice.Locale, "-")[0],
		}
	}

	vm.calledCreate = true
	return nil
}

// Find finds all matching voices based on the provided attributes.
func (vm *VoicesManager) Find(criteria types.VoicesManagerFind) ([]types.VoicesManagerVoice, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if !vm.calledCreate {
		return nil, ErrNotCreated
	}

	var matchingVoices []types.VoicesManagerVoice
	for _, voice := range vm.voices {
		if (criteria.Gender == "" || voice.Gender == criteria.Gender) &&
			(criteria.Locale == "" || voice.Locale == criteria.Locale) &&
			(criteria.Language == "" || voice.Language == criteria.Language) {
			matchingVoices = append(matchingVoices, voice)
		}
	}

	return matchingVoices, nil
}
