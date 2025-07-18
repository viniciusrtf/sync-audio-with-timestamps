package core

import "errors"

var (
	// ErrInvalidManifest is returned when the manifest file is invalid.
	ErrInvalidManifest = errors.New("invalid manifest")
	// ErrProcessingEntry is returned when an error occurs while processing a manifest entry.
	ErrProcessingEntry = errors.New("processing entry failed")
)
