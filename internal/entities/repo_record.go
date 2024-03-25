package entities

import (
	"fmt"
	"time"
)

type RepoRecord struct {
	ID        string     `yaml:"id,omitempty"`
	ShortID   string     `yaml:"shortId,omitempty"`
	Title     string     `yaml:"title,omitempty"`
	Message   string     `yaml:"message,omitempty"`
	CreatedAt *time.Time `yaml:"createdAt,omitempty"`
	WebURL    string     `yaml:"webURL,omitempty"`
	Origin    string     `yaml:"origin,omitempty"`
}

func (r RepoRecord) String() string {
	return fmt.Sprintf("%s (%s): %s", r.Origin, r.ShortID, r.Title)
}
