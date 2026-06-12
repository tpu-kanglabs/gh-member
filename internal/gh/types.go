package gh

// Member は Organization のメンバーを表す。
type Member struct {
	Name       string // 空の場合は Login を使う（表示層で処理）
	Login      string
	Role       string // "ADMIN" or "MEMBER"
	DatabaseID int
	URL        string
}

// Org は viewer が参加している Organization を表す。
type Org struct {
	Login string
	Name  string
}
