package slack

// Paging describes the page information of a Slack response.
type Paging struct {
	Page           int `json:"page"`
	PageCount      int `json:"pages"`
	PageSize       int `json:"count"`
	TotalItemCount int `json:"total"`
}
