package model

import "github.com/pregnor/slemtra/pkg/errors"

// DeleteEmojisModel encapsulates information required for deleting emojis.
type DeleteEmojisModel interface {
	// Name returns the name of the emoji to delete from the workspace.
	Name() (name string)
}

// NewDeleteEmojisModel returns a model for deleting a custom emoji.
func NewDeleteEmojisModel(name string) (model DeleteEmojisModel, err error) {
	if name == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewDeleteEmojisModel", "key", "name", "reason", "empty value")
	}

	return &deleteEmojisModel{
		name: name,
	}, nil
}

// deleteEmojisModel provides a basic implementation of the corresponding model
// to delete emojis.
//
// Implements the model.DeleteEmojisModel interface.
type deleteEmojisModel struct {
	name string
}

// Name returns the name of the emoji to delete from the workspace.
//
// Implements the model.DeleteEmojisModel interface.
func (model *deleteEmojisModel) Name() (name string) {
	if model == nil {
		return ""
	}

	return model.name
}
