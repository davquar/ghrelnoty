package smtp

import (
	"fmt"
	"net/smtp"
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

// Notify sends an email to Destination, to announce a new release of the given repo.
func (d Destination) Notify(repo string, release string) error {
	subject := fmt.Sprintf("New release: %s:%s", repo, release)
	body := fmt.Sprintf("GHRelNoty detected a new release: %s:%s", repo, release)
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", d.From, d.To, subject, body))
	auth := smtp.PlainAuth("", d.From, d.Password, d.Host)
	err := smtp.SendMail(d.Host+":"+d.Port, auth, d.From, []string{d.To}, msg)

	return err
}
