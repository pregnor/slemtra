package upload

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// Configuration describes the necessary information for operating the
// upload tool.
type Configuration struct {
	SlackEmojiAliasPrefix      string `json:"slack_emoji_alias_prefix"`
	SlackEmojiAliasSuffix      string `json:"slack_emoji_alias_suffix"`
	SlackEmojiAliasTakenPrefix string `json:"slack_emoji_alias_taken_prefix"`
	SlackEmojiAliasTakenSuffix string `json:"slack_emoji_alias_taken_suffix"`
	SlackEmojiCookie           string `json:"slack_emoji_cookie"`
	SlackEmojiDirectory        string `json:"slack_emoji_directory"`
	SlackTeamName              string `json:"slack_team_name"`
}

// NewConfigurationFromCLI instantiates a configuration object read from the CLI
// argument `-configuration-file-path`.
func NewConfigurationFromCLI(rawArguments []string) (configuration *Configuration, err error) {
	if len(rawArguments) != 0 &&
		rawArguments[0] == os.Args[0] {
		rawArguments = rawArguments[1:]
	}

	configurationFilePath := ""
	cliFlags := flag.NewFlagSet("cli-arguments", flag.ContinueOnError)
	cliFlags.StringVar(&configurationFilePath, "configuration-file-path", "", "Path to the (JSON) configuration file.")

	err = cliFlags.Parse(rawArguments)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing configuration CLI arguments failed, raw arguments: '%+v'", rawArguments)
	}

	if configurationFilePath == "" {
		return nil, fmt.Errorf("required configuration CLI argument `-configuration-file-path` is empty, raw arguments: '%+v'", rawArguments)
	}

	configuration, err = NewConfigurationFromFile(configurationFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading configuration from file failed, path: '%+v'", configurationFilePath)
	}

	return configuration, nil
}

// NewConfigurationFromFile instantiates a configuration object read from a
// file with automatic encoding discovery through extension checking.
func NewConfigurationFromFile(configurationFilePath string) (configuration *Configuration, err error) {
	if configurationFilePath == "" {
		return configuration, fmt.Errorf("configuration file path is empty")
	}

	split := strings.Split(configurationFilePath, ".")
	if len(split) < 2 {
		return configuration, fmt.Errorf("file path misses extension, configuration file path: '%+v'", configurationFilePath)
	}

	configurationData, err := ioutil.ReadFile(configurationFilePath)
	if err != nil {
		return configuration, errors.Wrapf(err, "reading configuration file failed, configuration file path: '%+v'", configurationFilePath)
	}

	extension := filepath.Ext(configurationFilePath)
	switch extension {
	case ".json":
		return NewConfigurationFromJSON(configurationData)
	default:
		return configuration, fmt.Errorf("unsupported configuration file path extension, extension: '%+v'", extension)
	}
}

// NewConfigurationFromJSON instantiates a configuration object read from
// JSON encoded text binary data.
func NewConfigurationFromJSON(jsonConfiguration []byte) (configuration *Configuration, err error) {
	if len(jsonConfiguration) == 0 {
		return configuration, fmt.Errorf("configuration data is empty")
	}

	err = json.Unmarshal(jsonConfiguration, &configuration)
	if err != nil {
		return configuration, errors.Wrapf(err, "unmarshalling configuration JSON failed, configuration JSON: '%+v'", string(jsonConfiguration))
	}

	return configuration, nil
}
