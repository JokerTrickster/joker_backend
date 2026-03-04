package response

import "time"

type ResRankingItem struct {
	Rank        int       `json:"rank"`
	Nickname    string    `json:"nickname"`
	ClearTimeMs uint      `json:"clearTimeMs"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ResRankingList struct {
	Rankings []ResRankingItem `json:"rankings"`
}

type ResMyRankingEntry struct {
	Rank        int       `json:"rank"`
	Nickname    string    `json:"nickname"`
	ClearTimeMs uint      `json:"clearTimeMs"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ResMyRanking struct {
	Rank  int               `json:"rank"`
	Entry ResMyRankingEntry `json:"entry"`
}

type ResSubmitRanking struct {
	Rank        int  `json:"rank"`
	IsNewRecord bool `json:"isNewRecord"`
}
