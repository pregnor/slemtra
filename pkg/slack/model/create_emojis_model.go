package model

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/pregnor/slemtra/pkg/errors"
	"gopkg.in/resty.v1"
)

// CreateEmojisModel encapsulates information required to create a new custom
// emoji.
type CreateEmojisModel interface {
	// Aliases returns the slice of alias names for the emoji.
	Aliases() (aliases []string)

	// Name returns the name of the emoji to be created.
	Name() (name string)

	// URL returns the URL to the emoji's graphical content.
	URL() (emojiURL string)
}

// NewCreateEmojisModel returns a model for creating a custom emoji.
func NewCreateEmojisModel(name, emojiURL string, aliases []string) (model CreateEmojisModel, err error) {
	if name == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewCreateEmojiModel", "key", "name", "reason", "empty value")
	} else if emojiURL == "" {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewCreateEmojiModel", "key", "emojiURL", "reason", "empty value")
	}

	parsedEmojiURL, err := url.Parse(emojiURL)
	if err != nil {
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewCreateEmojiModel", "error", err, "key", "emojiURL",
			"reason", "url cannot be parsed", "value", emojiURL)
	}

	switch parsedEmojiURL.Scheme {
	case "", "file": // Note: no scheme means file path.
		_, err = ioutil.ReadFile(emojiURL) // nolint:gosec // File read is just a basic check and content is discarded.
		if err != nil {
			return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
				"context", "NewCreateEmojiModel", "error", err, "key", "emojiURL",
				"reason", "reading file failed", "value", emojiURL)
		}
	case "http", "https":
		timeout := 30 * time.Second // nolint:gomnd // Scalar value for timeout.

		ctx, cancelFunction := context.WithTimeout(context.Background(), timeout)
		defer cancelFunction()

		restClient := resty.NewWithClient(
			&http.Client{
				Timeout: timeout,
			},
		)
		request := restClient.R().
			SetContext(ctx)

		response, err := request.Get(emojiURL)
		if err != nil ||
			response.StatusCode() < 200 ||
			response.StatusCode() > 299 {
			return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
				"context", "NewCreateEmojiModel", "error", err, "key", "emojiURL",
				"reason", "requesting emoji URL failed", "value", emojiURL)
		}
	default:
		return nil, errors.NewErrorWithDetails(errors.ErrorInvalidValue,
			"context", "NewCreateEmojiModel", "key", "emojiURL",
			"reason", fmt.Sprintf("unknown scheme: %+v", parsedEmojiURL.Scheme), "value", emojiURL)
	}

	return &createEmojisModel{
		aliases:  aliases,
		emojiURL: emojiURL,
		name:     name,
	}, nil
}

// createEmojisModel provides a basic implementation for collecting information
// required for creating an emoji.
//
// Implements the model.CreateEmojiModel interface.
type createEmojisModel struct {
	aliases  []string
	emojiURL string
	name     string
}

// Aliases returns the slice of alias names for the emoji.
//
// Implements the model.CreateEmojiModel interface.
func (model *createEmojisModel) Aliases() (aliases []string) {
	if model == nil {
		return nil
	}

	return model.aliases
}

// Name returns the name of the emoji to be created.
//
// Implements the model.CreateEmojiModel interface.
func (model *createEmojisModel) Name() (name string) {
	if model == nil {
		return ""
	}

	return model.name
}

// URL returns the URL to the emoji's graphical content.
//
// Implements the model.CreateEmojiModel interface.
func (model *createEmojisModel) URL() (emojiURL string) {
	if model == nil {
		return ""
	}

	return model.emojiURL
}
