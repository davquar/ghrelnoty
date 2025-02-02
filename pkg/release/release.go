package release

import "fmt"

// Release holds data that describe a release.
type Release struct {
	Project     string
	Author      string
	Version     string
	Description string
	URL         string
}

func (r Release) Repo() string {
	return fmt.Sprintf("%s/%s", r.Author, r.Project)
}
