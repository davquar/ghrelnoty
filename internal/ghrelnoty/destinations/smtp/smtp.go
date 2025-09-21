package smtp

import (
	"bytes"
	"fmt"
	"net/smtp"

	"github.com/yuin/goldmark"
	goldmarkext "github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"it.davquar/gitrelnoty/pkg/release"
)

const boundary = "knef9j23nr82"

// Destination holds the configuration for the SMTP destination.
type Destination struct {
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	HTML     bool   `yaml:"html"`
}

// Notify sends an email to Destination, to announce a new Release of the given repo.
func (d Destination) Notify(release release.Release) error {
	subject := fmt.Sprintf("New release: %s %s", release.Repo(), release.Version)
	msg := makeMsg(d.From, d.To, subject, release, d.HTML)
	err := smtp.SendMail(d.Host+":"+d.Port, d.auth(), d.From, []string{d.To}, msg)

	return err
}

func makeMsg(from string, to string, subject string, r release.Release, isMultipart bool) []byte {
	if isMultipart {
		return fmt.Appendf(nil,
			"From: %s\r\n"+
				"To: %s\r\n"+
				"Subject: %s\r\n"+
				"MIME-Version: 1.0\r\n"+
				"Content-Type: multipart/alternative; boundary=%q\r\n"+
				"\r\n%s",
			from, to, subject, boundary, multipart(htmlContent(r), plaintextContent(r)))
	}

	return fmt.Appendf(nil,
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s\r\n", from, to, subject, plaintextContent(r))
}

// auth returns smtp.Auth if username and password are set, otherwise nil indicating NOAUTH
func (d Destination) auth() smtp.Auth {
	if d.Username == "" && d.Password == "" {
		return nil
	}
	return smtp.PlainAuth("", d.From, d.Password, d.Host)
}

func plaintextContent(r release.Release) string {
	return fmt.Sprintf(`GHRelNoty
---------

New release for %s/%s: %s

%s

URL: %s`, r.Author, r.Project, r.Version,
		r.Description,
		r.URL)
}

func htmlContent(r release.Release) string {
	md := goldmark.New(
		goldmark.WithExtensions(goldmarkext.GFM),
		goldmark.WithRendererOptions(
			goldmarkhtml.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert([]byte(r.Description), &buf); err != nil {
		return plaintextContent(r)
	}

	return fmt.Sprintf(`<h1>New release for %s/%s: %s</h1>
</hr>
%s
</hr>
<p>URL: <a href="%s">%s</a></p>`, r.Author, r.Project, r.Version, buf.String(), r.URL, r.URL)
}

func multipart(html string, plain string) string {
	mime := fmt.Sprintf("MIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
	plainPart := fmt.Sprintf("--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", boundary, plain)
	htmlPart := fmt.Sprintf("--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n<html><body>%s</body></html>\r\n", boundary, html)
	end := fmt.Sprintf("--%s--\r\n", boundary)

	return mime + plainPart + htmlPart + end
}
