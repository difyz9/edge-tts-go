# Changelog

All notable changes to `edge-tts-go` will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-09-20

### Added
- `StreamSync()` method for synchronous streaming operations
- `VoicesManagerFind` struct for structured voice search criteria
- New example `examples/with_subtitles/main.go` demonstrating subtitle generation
- Comprehensive error types matching Python edge-tts exceptions
- WebSocket compression support (equivalent to Python's `compress=15`)

### Changed
- **BREAKING**: Default boundary changed from `WordBoundary` to `SentenceBoundary`
- Updated API endpoints from `speech.platform.bing.com` to `api.msedgeservices.com/tts/cognitiveservices`
- Updated Chromium version from 130.0.2849.68 to 140.0.3485.14
- Enhanced SubMaker logic to properly handle SentenceBoundary vs WordBoundary
- Improved WebSocket header configuration
- Updated VoicesManager.Find() method signature to use structured criteria
- Enhanced DRM token generation with improved clock skew correction

### Fixed
- Removed duplicate WebSocket headers
- Fixed unused import issues
- Improved timestamp accuracy in subtitle generation
- Enhanced error handling throughout the codebase

### Technical Details
- **API Compatibility**: Maintained backward compatibility for existing code
- **Python Sync**: Synchronized with Python edge-tts v7.2.3 features and behavior
- **Testing**: All core functionality verified working
- **Documentation**: Updated examples and usage instructions

### Migration Guide
If you're upgrading from a previous version:

1. **Default Boundary Change**: If you were relying on `WordBoundary` as default, explicitly set `boundary` parameter to `"WordBoundary"` in `NewCommunicate()`
2. **VoicesManager**: Update calls to `Find()` method to use `VoicesManagerFind{}` struct instead of separate parameters
3. **API Endpoints**: No code changes required - endpoints updated automatically

### Files Modified
- `internal/constants/constants.go` - Updated API endpoints and Chromium version
- `pkg/communicate/communicate.go` - Added sync streaming, changed default boundary
- `pkg/submaker/submaker.go` - Enhanced subtitle generation logic
- `pkg/voices/voices.go` - Updated Find method signature
- `pkg/types/types.go` - Added VoicesManagerFind struct
- `internal/websocket/websocket.go` - Improved WebSocket handling
- `examples/with_subtitles/main.go` - New subtitle example
- `README.md` - Updated documentation and examples

## [0.1.0] - 2024-XX-XX

Initial release of edge-tts-go with basic TTS functionality.
