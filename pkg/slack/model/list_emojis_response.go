package model

// ListEmojisResponse describes the Slack response for the emoji listing
// request.
type ListEmojisResponse struct {
	CustomEmojiTotalCount int64           `json:"custom_emoji_total_count"`
	DisabledEmojis        []EmojiResponse `json:"disabled_emoji"`
	Emojis                []EmojiResponse `json:"emoji"`
	IsOk                  bool            `json:"ok"`
	Paging                PagingResponse  `json:"paging"`
}
