package model

import (
	"strings"
	"time"

	"github.com/pregnor/slemtra/pkg/errors"
)

// ListEmojisModel encapsulates information returned when listing emojis.
type ListEmojisModel interface {
	// Aliases returns the slice of alias names for the emoji.
	Aliases() (aliases []string)

	// CreationTime returns the point in time when the emoji was created in the
	// workspace.
	CreationTime() (creationTime time.Time)

	// CreatorID returns the userID of the user who created the emoji in the
	// workspace.
	CreatorID() (userID string)

	// IsRemovable returns whether the emoji can be deleted from the workspace.
	IsRemovable() (isRemovable bool)

	// IsUsable returns whether the emoji is intact with a valid name and can be
	// used for messages and reactions in the workspace.
	IsUsable() (isUsable bool)

	// Name returns the name of the emoji.
	Name() (name string)

	// URL returns the URL to the emoji's graphical content.
	URL() (emojiURL string)
}

// NewListEmojisModel returns a model for listing custom emojis.
func NewListEmojisModel(response *EmojiResponse) (model ListEmojisModel, err error) {
	if response.AliasFor != "" ||
		response.IsAlias {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewListEmojisModel", "key", "model",
			"reason", "aliases must be processed as part of their original emoji", "value", model)
	}

	return &listEmojisModel{
		aliases:      response.Synonyms,
		creationTime: time.Unix(response.Created, 0),
		creatorID:    response.UserID,
		emojiURL:     strings.ReplaceAll(response.URL, "\\", ""), // Note: the emoji URL is backslash escaped.
		isRemovable:  response.CanDelete,
		isUsable:     !response.IsBad,
		name:         response.Name,
	}, nil
}

// listEmojisModel provides a basic implementation for collecting information
// about emojis.
//
// Implements the model.ListEmojiModel interface.
type listEmojisModel struct {
	aliases      []string
	creationTime time.Time
	creatorID    string
	emojiURL     string
	isRemovable  bool
	isUsable     bool
	name         string
}

// Aliases returns the slice of alias names for the emoji.
//
// Implements the model.ListEmojiModel interface.
func (model *listEmojisModel) Aliases() (aliases []string) {
	if model == nil {
		return nil
	}

	return model.aliases
}

// CreationTime returns the point in time when the emoji was created in the
// workspace.
//
// Implements the model.ListEmojiModel interface.
func (model *listEmojisModel) CreationTime() (creationTime time.Time) {
	if model == nil {
		return time.Time{}
	}

	return model.creationTime
}

// CreatorID returns the userID of the user who created the emoji in the
// workspace.
//
// Implements the model.ListEmojiModel interface.
func (model *listEmojisModel) CreatorID() (userID string) {
	if model == nil {
		return ""
	}

	return model.creatorID
}

// IsRemovable returns whether the emoji can be deleted from the workspace.
//
// Implements the model.ListEmojiModel interface.
func (model *listEmojisModel) IsRemovable() (isRemovable bool) {
	if model == nil {
		return false
	}

	return model.isRemovable
}

// IsUsable returns whether the emoji is intact with a valid name and can be
// used for messages and reactions in the workspace.
//
// Implements the model.ListEmojiModel interface.
func (model *listEmojisModel) IsUsable() (isUsable bool) {
	if model == nil {
		return false
	}

	return model.isUsable
}

// Name returns the name of the emoji to be listd.
//
// Implements the model.ListEmojiModel interface.
func (model *listEmojisModel) Name() (name string) {
	if model == nil {
		return ""
	}

	return model.name
}

// URL returns the URL to the emoji's graphical content.
//
// Implements the model.ListEmojiModel interface.
func (model *listEmojisModel) URL() (emojiURL string) {
	if model == nil {
		return ""
	}

	return model.emojiURL
}
