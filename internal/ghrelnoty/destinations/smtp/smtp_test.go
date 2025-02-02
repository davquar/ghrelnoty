package smtp

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	smtpmock "github.com/mocktools/go-smtp-mock/v2"
	"it.davquar/gitrelnoty/pkg/release"
)

func TestNotifySMTP(t *testing.T) {
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	err := server.Start()
	if err != nil {
		t.Fatalf("cannot start smtp mock server: %v", err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Logf("smtp mock server error while closing: %v", err)
		}
	}()

	d := Destination{
		From:     "from@test.test",
		To:       "to@test.test",
		Host:     "127.0.0.1",
		Username: "",
		Password: "",
		Port:     strconv.Itoa(server.PortNumber()),
	}

	err = d.Notify(release.Release{
		Project:     "dummy-project",
		Author:      "dummy-author",
		Version:     "v1.2.3",
		Description: "This is a dummy release just for testing.",
		URL:         "some-url",
	})
	if err != nil {
		t.Fatalf("unexpected error sending email: %v", err)
	}

	msgs := server.MessagesAndPurge()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	if len(msgs[0].RcpttoRequestResponse()) != 1 {
		t.Fatalf("expected 1 receiver, got %d", len(msgs[0].RcpttoRequestResponse()))
	}

	expmsg := makeRawMsg(d.From, d.To)
	if msgs[0].MsgRequest() != expmsg {
		t.Fatalf("expected msg '%s', got '%s'", expmsg, msgs[0].MsgRequest())
	}
}

func makeRawMsg(from string, to string) string {
	s := fmt.Sprintf(`From: %s
To: %s
Subject: New release: dummy-author/dummy-project v1.2.3

GHRelNoty
---------

New release for dummy-author/dummy-project: v1.2.3

This is a dummy release just for testing.

URL: some-url
`, from, to)

	// set \r\n as newline sequence, because that's what is used in the msg field
	return strings.ReplaceAll(s, "\n", "\r\n")
}
