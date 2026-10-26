package actorbridge

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const maximumTLSFileBytes = 1024 * 1024

type ClientTLSFiles struct {
	CertificateFile string
	PrivateKeyFile  string
	ServerCAFile    string
}

type ServerTLSFiles struct {
	CertificateFile string
	PrivateKeyFile  string
	ClientCAFile    string
}

type ClientIdentity struct {
	certificate tls.Certificate
	roots       *x509.CertPool
}

func LoadClientIdentity(
	files ClientTLSFiles,
	now time.Time,
) (ClientIdentity, error) {
	certificate, err := loadTLSCertificate(
		files.CertificateFile, files.PrivateKeyFile,
	)
	if err != nil || !validLocalCertificate(
		certificate, actorCorePeerURI, x509.ExtKeyUsageClientAuth, now,
	) {
		return ClientIdentity{}, NewFault(FaultUnavailable)
	}
	roots, err := loadCertPool(files.ServerCAFile)
	if err != nil {
		return ClientIdentity{}, NewFault(FaultUnavailable)
	}
	return ClientIdentity{certificate: certificate, roots: roots}, nil
}

func loadTLSCertificate(certificatePath, keyPath string) (tls.Certificate, error) {
	certificatePEM, err := readTLSFile(certificatePath, false)
	if err != nil {
		return tls.Certificate{}, err
	}
	privateKeyPEM, err := readTLSFile(keyPath, true)
	if err != nil {
		return tls.Certificate{}, err
	}
	certificate, err := tls.X509KeyPair(certificatePEM, privateKeyPEM)
	clear(privateKeyPEM)
	if err != nil || len(certificate.Certificate) < 1 {
		return tls.Certificate{}, errors.New("actor bridge TLS identity is invalid")
	}
	leaf, err := x509.ParseCertificate(certificate.Certificate[0])
	if err != nil {
		return tls.Certificate{}, errors.New("actor bridge TLS leaf is invalid")
	}
	certificate.Leaf = leaf
	return certificate, nil
}

func loadCertPool(path string) (*x509.CertPool, error) {
	document, err := readTLSFile(path, false)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(document) {
		return nil, errors.New("actor bridge TLS trust roots are invalid")
	}
	return pool, nil
}

func readTLSFile(path string, private bool) ([]byte, error) {
	if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("actor bridge TLS path is invalid")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, errors.New("actor bridge TLS path cannot be resolved")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return nil, errors.New("actor bridge TLS file cannot be opened")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() ||
		(private && runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0) {
		return nil, errors.New("actor bridge TLS file permissions are invalid")
	}
	document, err := io.ReadAll(io.LimitReader(file, maximumTLSFileBytes+1))
	if err != nil || len(document) < 1 || len(document) > maximumTLSFileBytes {
		return nil, errors.New("actor bridge TLS file size is invalid")
	}
	return document, nil
}
