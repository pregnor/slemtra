package model

import (
	"github.com/pregnor/slemtra/pkg/errors"
)

// CreateEmojisAliasModel collects all information required for creating an
// emoji alias.
type CreateEmojisAliasModel interface {
	// Alias returns the alias name to create.
	Alias() string

	// Name returns the name of the existing emoji to create the alias for.
	Name() string
}

// NewCreateEmojisAliasModel returns a model for creating an emoji alias.
func NewCreateEmojisAliasModel(name, alias string) (model CreateEmojisAliasModel, err error) {
	if name == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewCreateEmojiAliasModel", "key", "name", "reason", "empty value")
	} else if alias == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewCreateEmojiAliasModel", "key", "alias", "reason", "empty value")
	}

	return &createEmojisAliasModel{
		alias: alias,
		name:  name,
	}, nil
}

// createEmojisAliasModel provides a basic implementation for collecting
// information required for creating an emoji alias.
//
// Implements the model.CreateEmojiAliasModel interface.
type createEmojisAliasModel struct {
	alias string
	name  string
}

// Alias returns the alias name to create.
//
// Implements the model.CreateEmojiAliasModel interface.
func (model *createEmojisAliasModel) Alias() string {
	if model == nil {
		return ""
	}

	return model.alias
}

// Name returns the name of the existing emoji to create the alias for.
//
// Implements the model.CreateEmojiAliasModel interface.
func (model *createEmojisAliasModel) Name() string {
	if model == nil {
		return ""
	}

	return model.name
}
