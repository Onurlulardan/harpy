package notificationhttp

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRoutesExposeFinalAndOnlyRealLegacyCompatibilitySurface(t *testing.T) {
	access, _ := NewAccess(httpTestUser, httpTestTenant, 1, 1,
		PermissionInboxRead, OperationInboxList, "", false, false)
	routes := newHTTPTestHandler(access, &httpApplicationFake{}).Routes()
	if len(routes) != 10 {
		t.Fatalf("route count=%d", len(routes))
	}
	for _, route := range routes {
		if strings.Contains(route.Pattern, "sse") {
			t.Fatal("invented notification SSE route exposed")
		}
	}
}

func TestTicketAccessRequiresSessionCSRFSecureOrigin(t *testing.T) {
	access, _ := NewAccess(httpTestUser, httpTestTenant, 1, 7,
		PermissionInboxRead, OperationTicketCreate, "https://app.harpy.test", true, true)
	verifier := &accessVerifierFake{access: access}
	app := &httpApplicationFake{}
	handler, _ := NewHandler(HandlerDependencies{Access: verifier, Application: app,
		Realtime: &realtimeFake{}, Report: func(error) {},
		Correlation: func(*http.Request) string { return "10000000-0000-4000-8000-000000000099" }})
	request := httptest.NewRequest(http.MethodPost, "/api/notification-realtime-tickets", nil)
	request.Header.Set("Origin", "https://app.harpy.test")
	recorder := httptest.NewRecorder()
	handler.createTicket(recorder, request)
	if recorder.Code != http.StatusCreated || app.ticketCalls != 1 ||
		!verifier.request.RequireSession || !verifier.request.RequireCSRF ||
		!verifier.request.RequireHTTPSOrigin || !verifier.request.RequireSecureTransport {
		t.Fatalf("ticket security request=%+v status=%d", verifier.request, recorder.Code)
	}
}

func TestMalformedUpgradeCannotBurnTicket(t *testing.T) {
	access, _ := NewAccess(httpTestUser, httpTestTenant, 1, 7,
		PermissionInboxRead, OperationSocketConnect, "https://app.harpy.test", true, false)
	app := &httpApplicationFake{}
	handler := newHTTPTestHandler(access, app)
	request := httptest.NewRequest(http.MethodGet, "/ws/notifications", nil)
	recorder := httptest.NewRecorder()
	handler.connect(recorder, request)
	if app.consumeCalls != 0 || recorder.Code != 422 {
		t.Fatalf("malformed handshake consumed=%d status=%d", app.consumeCalls, recorder.Code)
	}
}

func TestUpgradeNeverEchoesTicketProtocol(t *testing.T) {
	raw := make([]byte, 32)
	_, _ = rand.Read(raw)
	ticket := base64.RawURLEncoding.EncodeToString(raw)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := upgradeSocket(w, r)
		if err == nil {
			connection.close()
		}
	}))
	defer server.Close()
	address := strings.TrimPrefix(server.URL, "http://")
	connection, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))
	request := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\n"+
		"Connection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: %s\r\n"+
		"Sec-WebSocket-Protocol: harpy.notifications.v1, fc-ticket.%s\r\n\r\n",
		address, key, ticket)
	_, _ = io.WriteString(connection, request)
	response := make([]byte, 2048)
	read, _ := connection.Read(response)
	document := string(response[:read])
	if !strings.Contains(document, "Sec-WebSocket-Protocol: harpy.notifications.v1\r\n") ||
		strings.Contains(document, ticket) {
		t.Fatalf("unsafe upgrade response: %q", document)
	}
}
