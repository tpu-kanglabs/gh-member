package gh

// Member represents a member of a GitHub Organization.
type Member struct {
	Name       string // empty when not set; presentation layer falls back to Login
	Login      string
	Role       string // "ADMIN" or "MEMBER"
	DatabaseID int
	URL        string
}

// Org represents a GitHub Organization the authenticated user belongs to.
type Org struct {
	Login string
	Name  string
}
