package api

import (
	"github.com/eparis/bugzilla"
)

type BugAction struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Default     bool               `json:"default"`
	Query       bugzilla.Query     `json:"query"`
	Update      bugzilla.BugUpdate `json:"update"`
}
