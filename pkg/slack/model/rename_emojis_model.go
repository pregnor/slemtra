package model // nolint:dupl // The download and rename models are structurally similar but semantically different.

import "github.com/pregnor/slemtra/pkg/errors"

// RenameEmojisModel encapsulates information required for renaming emojis.
type RenameEmojisModel interface {
	// NewName returns the new name of the emoji to change to in the workspace.
	NewName() (newName string)

	// OriginalName returns the original name of the emoji to change from in the
	// workspace.
	OriginalName() (originalName string)
}

// NewRenameEmojisModel returns a model for renaming a custom emoji.
func NewRenameEmojisModel(originalName, newName string) (model RenameEmojisModel, err error) {
	if originalName == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewRenameEmojisModel", "key", "originalName", "reason", "empty value")
	} else if newName == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewRenameEmojisModel", "key", "newName", "reason", "empty value")
	}

	return &renameEmojisModel{
		newName:      newName,
		originalName: originalName,
	}, nil
}

// renameEmojisModel provides a basic implementation of the corresponding model
// to rename emojis.
//
// Implements the model.RenameEmojisModel interface.
type renameEmojisModel struct {
	newName      string
	originalName string
}

// NewName returns the new name of the emoji to change to in the workspace.
//
// Implements the model.RenameEmojisModel interface.
func (model *renameEmojisModel) NewName() (newName string) {
	if model == nil {
		return ""
	}

	return model.newName
}

// OriginalName returns the original name of the emoji to change from in the
// workspace.
//
// Implements the model.RenameEmojisModel interface.
func (model *renameEmojisModel) OriginalName() (originalName string) {
	if model == nil {
		return ""
	}

	return model.originalName
}
