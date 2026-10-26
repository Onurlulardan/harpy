package notificationhttp

import "testing"

func TestInboxQueryUsesClosedRESTQueryAllowlist(t *testing.T) {
	raw := "filter[0][field]=unread&filter[0][op]=eq&filter[0][value]=true&" +
		"filter[1][field]=category&filter[1][op]=eq&filter[1][value]=security&" +
		"sort[0][field]=inbox_revision&sort[0][direction]=desc&page[size]=25"
	filter, err := decodeInboxQuery(raw)
	if err != nil || !filter.UnreadOnly || filter.Category == nil ||
		*filter.Category != "security" || filter.Limit != 25 {
		t.Fatalf("closed query was rejected: %#v %v", filter, err)
	}
	if _, err = decodeInboxQuery("search[term]=secret"); err == nil {
		t.Fatal("unsupported search was accepted")
	}
	if _, err = decodeInboxQuery("filter[0][field]=recipient_id&filter[0][op]=eq&filter[0][value]=x"); err == nil {
		t.Fatal("recipient authority was accepted from query")
	}
}

func TestInboxCursorIsOpaqueAndCanonical(t *testing.T) {
	encoded, err := encodeInboxCursor("42")
	if err != nil || encoded == "42" {
		t.Fatal("cursor was not encoded")
	}
	decoded, err := decodeInboxCursor(encoded)
	if err != nil || decoded != "42" {
		t.Fatal("cursor did not round trip")
	}
	if _, err = decodeInboxCursor("42"); err == nil {
		t.Fatal("raw storage cursor was accepted")
	}
}
