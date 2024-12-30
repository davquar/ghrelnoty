package ghrelnoty

import "testing"

func TestSeparateName(t *testing.T) {
	r := Repository{
		Name: "theauthor/thereponame",
	}

	author, reponame := r.SeparateName()
	if author != "theauthor" && reponame != "thereponame" {
		t.Fatalf("expected {theauthor, thereponame}, got: {%s, %s}", author, reponame)
	}
}
