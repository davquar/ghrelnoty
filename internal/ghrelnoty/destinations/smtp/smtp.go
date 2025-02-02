package smtp

import (
	"fmt"
	"net/smtp"

	"it.davquar/gitrelnoty/pkg/release"
)

// Destination holds the configuration for the SMTP destination.
type Destination struct {
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Notify sends an email to Destination, to announce a new Release of the given repo.
func (d Destination) Notify(release release.Release) error {
	subject := fmt.Sprintf("New release: %s %s", release.Repo(), release.Version)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", d.From, d.To, subject, plainText(release)))

	err := smtp.SendMail(d.Host+":"+d.Port, d.auth(), d.From, []string{d.To}, msg)

	return err
}

// auth returns smtp.Auth if username and password are set, otherwise nil indicating NOAUTH
func (d Destination) auth() smtp.Auth {
	if d.Username == "" && d.Password == "" {
		return nil
	}
	return smtp.PlainAuth("", d.From, d.Password, d.Host)
}

// plainText returns the plain text email body for the given Release.
func plainText(r release.Release) string {
	return fmt.Sprintf(`GHRelNoty
---------

New release for %s/%s: %s

%s

URL: %s`, r.Author, r.Project, r.Version,
		r.Description,
		r.URL)
}
