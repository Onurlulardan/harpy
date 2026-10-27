package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/kms/apiv1/kmspb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestTLSHandlerMultiplexesSTSAndKMSGRPC(t *testing.T) {
	grpcServer := grpc.NewServer()
	kmspb.RegisterKeyManagementServiceServer(grpcServer, newKMSService(fixtureKey))
	server := httptest.NewUnstartedServer(newHTTPHandler(grpcServer))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()
	defer grpcServer.Stop()

	assertSTSOverHTTP2(t, server)
	assertKMSOverHTTP2(t, server)
}

func assertSTSOverHTTP2(t *testing.T, server *httptest.Server) {
	t.Helper()
	request, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/v1/token",
		strings.NewReader(validTokenValues().Encode()),
	)
	if err != nil {
		t.Fatalf("new STS request: %v", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatalf("STS request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.ProtoMajor != 2 {
		t.Fatalf("STS status = %d over HTTP/%d", response.StatusCode, response.ProtoMajor)
	}
}

func assertKMSOverHTTP2(t *testing.T, server *httptest.Server) {
	t.Helper()
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	serverName := server.Certificate().DNSNames[0]
	connection, err := grpc.NewClient(
		server.Listener.Addr().String(),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
			RootCAs:    roots,
			ServerName: serverName,
		})),
	)
	if err != nil {
		t.Fatalf("new gRPC client: %v", err)
	}
	defer connection.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := kmspb.NewKeyManagementServiceClient(connection).RawEncrypt(
		ctx, validEncryptFixture(),
	)
	if err != nil {
		t.Fatalf("RawEncrypt over gRPC: %v", err)
	}
	if !validChecksum(response.Ciphertext, response.CiphertextCrc32C) {
		t.Fatal("gRPC response failed CRC32C verification")
	}
}
