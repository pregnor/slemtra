package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	upload "github.com/pregnor/slack-emoji-upload"
	"github.com/pregnor/slack-emoji-upload/slack"
)

func handleFatalError(condition bool, exitCode int, messages ...interface{}) {
	if condition {
		log.Println(messages...)
		log.Println("Aborting on fatal error")
		os.Exit(exitCode)
	}
}

func main() {
	configuration, err := upload.NewConfigurationFromCLI(os.Args)
	handleFatalError(err != nil, 1, errors.Wrapf(err, "loading configuration failed, CLI arguments: '%+v'", os.Args))

	slackClient, err := slack.NewSlackClient(configuration.SlackTeamName, configuration.SlackEmojiCookie)
	handleFatalError(err != nil, 2, errors.Wrapf(err, "initializing Slack client failed, configuration: '%+v'", configuration))

	log.Printf("Existing emojis:\n")
	for _, emoji := range slackClient.Emojis {
		log.Printf(":%s:\n", emoji.Name)
	}
	log.Printf("\n")

	err = slackClient.PostEmojis(configuration.SlackEmojiDirectory, configuration.SlackEmojiAliasPrefix, configuration.SlackEmojiAliasSuffix, configuration.SlackEmojiAliasTakenPrefix, configuration.SlackEmojiAliasTakenSuffix)
	handleFatalError(err != nil, 3, errors.Wrapf(err, "posting emojis failed, directory: '%+v', prefix: '%+v', suffix: '%+v'", configuration.SlackEmojiDirectory, configuration.SlackEmojiAliasPrefix, configuration.SlackEmojiAliasSuffix))
}
