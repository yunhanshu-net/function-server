package model

type Runner struct {
	Base
	Title       string `json:"title"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`

	Language  string `json:"language"`
	Status    int8   `json:"status"`
	RuncherID *int64 `json:"runcher_id"`
	IsPublic  bool   `json:"is_public"`
	Tags      string `json:"tags"`

	TreeID          int64  `json:"tree_id"`
	ForkFromUser    string `json:"fork_from_user,omitempty"`
	ForkFromVersion string `json:"fork_from_version"`
	ForkFromID      *int64 `json:"fork_from_id"`

	User string `json:"user"`
}

func (r *Runner) TableName() string {
	return "runner"
}
