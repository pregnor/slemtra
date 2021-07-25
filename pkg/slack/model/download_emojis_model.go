package model // nolint:dupl // The download and rename models are structurally similar but semantically different.

import "github.com/pregnor/slemtra/pkg/errors"

// DownloadEmojisModel encapsulates information required for downloading emojis.
type DownloadEmojisModel interface {
	// Name returns the name of the emoji to download from the workspace.
	Name() (name string)

	// Path returns the file path to download the emoji content to from the
	// workspace.
	Path() (path string)
}

// NewDownloadEmojisModel returns a model for downloading a custom emoji.
func NewDownloadEmojisModel(name, path string) (model DownloadEmojisModel, err error) {
	if name == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewDownloadEmojisModel", "key", "name", "reason", "empty value")
	} else if path == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewDownloadEmojisModel", "key", "path", "reason", "empty value")
	}

	return &downloadEmojisModel{
		name: name,
		path: path,
	}, nil
}

// downloadEmojisModel provides a basic implementation of the corresponding
// model to download emojis.
//
// Implements the model.DownloadEmojisModel interface.
type downloadEmojisModel struct {
	name string
	path string
}

// Name returns the name of the emoji to download from the workspace.
//
// Implements the model.DownloadEmojisModel interface.
func (model *downloadEmojisModel) Name() (name string) {
	if model == nil {
		return ""
	}

	return model.name
}

// Path returns the file path to download the emoji content to from the
// workspace.
func (model *downloadEmojisModel) Path() (path string) {
	if model == nil {
		return ""
	}

	return model.path
}
