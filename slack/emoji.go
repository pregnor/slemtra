package slack

// Emoji encapsulates the available information on a single Slack emoji.
type Emoji struct {
	AliasFor        string   `json:"alias_for"`
	AvatarHash      string   `json:"avatar_hash"`
	CanDelete       bool     `json:"can_delete"`
	Created         int64    `json:"created"`
	IsAlias         int      `json:"is_alias"`
	IsBad           bool     `json:"is_bad"`
	Name            string   `json:"name"`
	Synonyms        []string `json:"synonyms"`
	TeamID          string   `json:"team_id"`
	URL             string   `json:"url"`
	UserDisplayName string   `json:"user_display_name"`
	UserID          string   `json:"user_id"`
}
