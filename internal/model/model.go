package model

// PRMeta holds minimal pull request metadata needed for context.
type PRMeta struct {
	Number int
	Title  string
	URL    string
	Goal   string
	Repo   string
}

// Comment represents a single review comment.
type Comment struct {
	ID        string
	Body      string
	Author    string
	CreatedAt string
	URL       string
}

// Thread groups review comments for a single location.
type Thread struct {
	ID       string
	FilePath string
	Line     int
	Resolved bool
	Comments []Comment
	DiffHunk string
	URL      string
}

// InboxItem is the compacted, prioritized representation of a thread.
type InboxItem struct {
	ThreadID   string `json:"threadId"`
	Priority   string `json:"priority"`
	FilePath   string `json:"filePath"`
	LineNumber int    `json:"lineNumber"`
	Author     string `json:"author"`
	Summary    string `json:"summary"`
	Latest     string `json:"latest"`
	URL        string `json:"url"`
	Resolved   bool   `json:"resolved"`
}
