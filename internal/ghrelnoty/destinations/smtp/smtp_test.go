package smtp

import (
	"strconv"
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
		Project: "dummy-project",
		Author:  "dummy-author",
		Version: "v1.2.3",
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
}
