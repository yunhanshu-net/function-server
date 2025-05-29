package model

import "gorm.io/gorm"

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

	FullNamePath string `json:"full_name_path" gorm:"-"`
	User         string `json:"user"`
}

func (r *Runner) TableName() string {
	return "runner"
}

func (r *Runner) AfterFind(db *gorm.DB) error {
	r.FullNamePath = r.User + "/" + r.Name
	return nil
}
