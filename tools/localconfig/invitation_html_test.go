package main

import (
	"strings"
	"testing"
)

func TestInvitationGraphiteHTMLKeepsSinglePlaceholder(t *testing.T) {
	if strings.Count(invitationGraphiteHTML, "{{INVITATION_URL}}") != 1 {
		t.Fatal("invitation HTML must contain {{INVITATION_URL}} once")
	}
	if !strings.Contains(invitationGraphiteHTML, "#18181b") {
		t.Fatal("invitation HTML is missing the graphite theme")
	}
}
