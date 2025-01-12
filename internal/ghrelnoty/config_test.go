package ghrelnoty

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
	"it.davquar/gitrelnoty/internal/ghrelnoty/destinations/smtp"
)

func TestSeparateName(t *testing.T) {
	r := RepositoryConfig{
		Name: "theauthor/thereponame",
	}

	author, reponame := r.SeparateName()
	if author != "theauthor" && reponame != "thereponame" {
		t.Fatalf("expected {theauthor, thereponame}, got: {%s, %s}", author, reponame)
	}
}

func TestDestinationsYAMLUnmarshalOK(t *testing.T) {
	y := `
destinations:
  mydstname:
    type: smtp
    config:
`

	var c Config
	err := yaml.Unmarshal([]byte(y), &c)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(c.Destinations) != 1 {
		t.Fatalf("expected 1 destination, got %d", len(c.Destinations))
	}

	d, ok := c.Destinations["mydstname"]
	if !ok {
		t.Fatal("key 'mydstname' not found")
	}

	n, err := d.Notifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok = n.(*smtp.Destination)
	if !ok {
		t.Fatalf("expected notifier of type SMTP, got %s", reflect.TypeOf(n))
	}
}

func TestDestinationsYAMLUnmarshalUnknownNotifier(t *testing.T) {
	y := `
destinations:
  mydstname:
    type: shouldfailhere
    config:
`

	var c Config
	err := yaml.Unmarshal([]byte(y), &c)
	if err == nil {
		t.Fatalf("expected unmarshal error due to unknown destination type, got Config=%v", c)
	}
}
