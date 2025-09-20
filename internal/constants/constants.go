// Package constants contains all the constants used in the edge-tts-go project.
package constants

const (
	// BaseURL is the base URL for the Microsoft Edge TTS service.
	BaseURL = "api.msedgeservices.com/tts/cognitiveservices"

	// TrustedClientToken is the token used to authenticate with the Microsoft Edge TTS service.
	TrustedClientToken = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"

	// WSSURL is the WebSocket URL for the Microsoft Edge TTS service.
	WSSURL = "wss://" + BaseURL + "/websocket/v1?Ocp-Apim-Subscription-Key=" + TrustedClientToken

	// VoiceList is the URL for the voice list.
	VoiceList = "https://" + BaseURL + "/voices/list?Ocp-Apim-Subscription-Key=" + TrustedClientToken

	// DefaultVoice is the default voice used for TTS.
	DefaultVoice = "en-US-EmmaMultilingualNeural"

	// ChromiumFullVersion is the full version of Chromium used in the User-Agent header.
	ChromiumFullVersion = "140.0.3485.14"

	// ChromiumMajorVersion is the major version of Chromium used in the User-Agent header.
	ChromiumMajorVersion = "140" // Extracted from ChromiumFullVersion

	// SecMSGECVersion is the version used in the Sec-MS-GEC header.
	SecMSGECVersion = "1-" + ChromiumFullVersion

	// WinEpoch is the Windows epoch time (January 1, 1601 UTC) in seconds.
	WinEpoch = 11644473600

	// SToNS is the number of nanoseconds in a second.
	SToNS = 1e9
)

// BaseHeaders are the base headers used in all HTTP requests.
var BaseHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" +
		" (KHTML, like Gecko) Chrome/" + ChromiumMajorVersion + ".0.0.0 Safari/537.36" +
		" Edg/" + ChromiumMajorVersion + ".0.0.0",
	"Accept-Encoding": "gzip, deflate, br",
	"Accept-Language": "en-US,en;q=0.9",
}

// WSSHeaders are the headers used in WebSocket requests.
var WSSHeaders = map[string]string{
	"Pragma":               "no-cache",
	"Cache-Control":        "no-cache",
	"Origin":               "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold",
	"Sec-WebSocket-Protocol": "synthesize",
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" +
		" (KHTML, like Gecko) Chrome/" + ChromiumMajorVersion + ".0.0.0 Safari/537.36" +
		" Edg/" + ChromiumMajorVersion + ".0.0.0",
	"Accept-Encoding": "gzip, deflate, br",
	"Accept-Language": "en-US,en;q=0.9",
}

// VoiceHeaders are the headers used in voice list requests.
var VoiceHeaders = map[string]string{
	"Authority": "speech.platform.bing.com",
	"Sec-CH-UA": "\" Not;A Brand\";v=\"99\", \"Microsoft Edge\";v=\"" + ChromiumMajorVersion + "\"," +
		" \"Chromium\";v=\"" + ChromiumMajorVersion + "\"",
	"Sec-CH-UA-Mobile": "?0",
	"Accept":           "*/*",
	"Sec-Fetch-Site":   "none",
	"Sec-Fetch-Mode":   "cors",
	"Sec-Fetch-Dest":   "empty",
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" +
		" (KHTML, like Gecko) Chrome/" + ChromiumMajorVersion + ".0.0.0 Safari/537.36" +
		" Edg/" + ChromiumMajorVersion + ".0.0.0",
	"Accept-Encoding": "gzip, deflate, br",
	"Accept-Language": "en-US,en;q=0.9",
}
