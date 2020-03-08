package slack

// EmojiListResponse describes the Slack response for the emoji listing
// request.
type EmojiListResponse struct {
	CustomEmojiTotalCount int64   `json:"custom_emoji_total_count"`
	DisabledEmojis        []Emoji `json:"disabled_emoji"`
	Emojis                []Emoji `json:"emoji"`
	IsOk                  bool    `json:"ok"`
	Paging                Paging  `json:"paging"`
}
