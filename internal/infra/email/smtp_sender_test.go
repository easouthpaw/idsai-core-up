package email

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSMTPSenderSendValidation(t *testing.T) {
	err := NewSMTPSender("", "25", "", "", "sender@example.local").Send(context.Background(), "to@example.local", "Subject", "Body")
	require.EqualError(t, err, "smtp is not configured")

	err = NewSMTPSender("127.0.0.1", "25", "", "", "sender@example.local").Send(context.Background(), " ", "Subject", "Body")
	require.EqualError(t, err, "email recipient is required")
}

func TestSMTPSenderSend(t *testing.T) {
	addr, received := startSMTPServer(t, false)
	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	sender := NewSMTPSenderWithTimeout(host, port, "", "", "IDS AI <sender@example.local>", time.Second)
	require.NoError(t, sender.Send(context.Background(), "to@example.local", " Hello ", "Body"))

	msg := <-received
	require.Contains(t, msg, "From: \"IDS AI\" <sender@example.local>")
	require.Contains(t, msg, "To: to@example.local")
	require.Contains(t, msg, "Subject: Hello")
	require.Contains(t, msg, "Body")
}

func TestSMTPSenderSendRequiresAuthExtensionWhenConfigured(t *testing.T) {
	addr, _ := startSMTPServer(t, false)
	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	sender := NewSMTPSenderWithTimeout(host, port, "user", "pass", "sender@example.local", time.Second)
	err = sender.Send(context.Background(), "to@example.local", "Subject", "Body")
	require.EqualError(t, err, "smtp: server doesn't support AUTH")
}

func TestSMTPSenderSendWithImplicitTLS(t *testing.T) {
	addr, received, certPEM := startTLSSMTPServer(t, false)
	host, port, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	prevPool := systemCertPool
	systemCertPool = func() (*x509.CertPool, error) {
		pool := x509.NewCertPool()
		require.True(t, pool.AppendCertsFromPEM(certPEM))
		return pool, nil
	}
	t.Cleanup(func() {
		systemCertPool = prevPool
	})

	sender := NewSMTPSenderWithOptions(host, port, "", "", "sender@example.local", time.Second, true)
	require.NoError(t, sender.Send(context.Background(), "to@example.local", "Subject", "TLS Body"))

	msg := <-received
	require.Contains(t, msg, "Subject: Subject")
	require.Contains(t, msg, "TLS Body")
}

func TestMessageAndFromNormalization(t *testing.T) {
	require.Equal(t, "\"IDS AI\" <sender@example.local>", normalizeFromHeader("IDS AI <sender@example.local>"))
	require.Equal(t, "sender@example.local", normalizeFromEnvelope("IDS AI <sender@example.local>"))
	require.Empty(t, normalizeFromEnvelope("not an address <broken>"))
	require.Equal(t, "plain@example.local", normalizeFromEnvelope("plain@example.local"))

	msg := buildMessage("sender@example.local", "to@example.local", "Subject", "")
	require.Contains(t, msg, "\r\n\r\n(empty)\r\n")
}

func startSMTPServer(t *testing.T, advertiseAuth bool) (string, <-chan string) {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = ln.Close()
	})

	messages := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		reader := bufio.NewReader(conn)
		writeLine(conn, "220 localhost ESMTP")

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if !errors.Is(err, io.EOF) {
					t.Logf("smtp test server read failed: %v", err)
				}
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				if advertiseAuth {
					writeLine(conn, "250-localhost")
					writeLine(conn, "250 AUTH PLAIN")
				} else {
					writeLine(conn, "250 localhost")
				}
			case strings.HasPrefix(cmd, "HELO"):
				writeLine(conn, "250 localhost")
			case strings.HasPrefix(cmd, "AUTH"):
				writeLine(conn, "235 authenticated")
			case strings.HasPrefix(cmd, "MAIL FROM:"):
				writeLine(conn, "250 sender ok")
			case strings.HasPrefix(cmd, "RCPT TO:"):
				writeLine(conn, "250 recipient ok")
			case cmd == "DATA":
				writeLine(conn, "354 end with dot")
				var body strings.Builder
				for {
					dataLine, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if strings.TrimRight(dataLine, "\r\n") == "." {
						break
					}
					body.WriteString(dataLine)
				}
				messages <- body.String()
				writeLine(conn, "250 queued")
			case cmd == "QUIT":
				writeLine(conn, "221 bye")
				return
			default:
				writeLine(conn, "250 ok")
			}
		}
	}()

	return ln.Addr().String(), messages
}

func startTLSSMTPServer(t *testing.T, advertiseAuth bool) (string, <-chan string, []byte) {
	t.Helper()

	certPEM, keyPEM := generateSelfSignedCert(t)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = ln.Close()
	})

	messages := make(chan string, 1)
	go serveSMTPConnection(t, ln, advertiseAuth, messages)
	return ln.Addr().String(), messages, certPEM
}

func serveSMTPConnection(t *testing.T, ln net.Listener, advertiseAuth bool, messages chan<- string) {
	conn, err := ln.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writeLine(conn, "220 localhost ESMTP")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Logf("smtp test server read failed: %v", err)
			}
			return
		}
		cmd := strings.ToUpper(strings.TrimSpace(line))
		switch {
		case strings.HasPrefix(cmd, "EHLO"):
			if advertiseAuth {
				writeLine(conn, "250-localhost")
				writeLine(conn, "250 AUTH PLAIN")
			} else {
				writeLine(conn, "250 localhost")
			}
		case strings.HasPrefix(cmd, "HELO"):
			writeLine(conn, "250 localhost")
		case strings.HasPrefix(cmd, "AUTH"):
			writeLine(conn, "235 authenticated")
		case strings.HasPrefix(cmd, "MAIL FROM:"):
			writeLine(conn, "250 sender ok")
		case strings.HasPrefix(cmd, "RCPT TO:"):
			writeLine(conn, "250 recipient ok")
		case cmd == "DATA":
			writeLine(conn, "354 end with dot")
			var body strings.Builder
			for {
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				if strings.TrimRight(dataLine, "\r\n") == "." {
					break
				}
				body.WriteString(dataLine)
			}
			messages <- body.String()
			writeLine(conn, "250 queued")
		case cmd == "QUIT":
			writeLine(conn, "221 bye")
			return
		default:
			writeLine(conn, "250 ok")
		}
	}
}

func writeLine(w io.Writer, line string) {
	_, _ = io.WriteString(w, line+"\r\n")
}

func generateSelfSignedCert(t *testing.T) ([]byte, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "127.0.0.1",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	return certPEM, keyPEM
}

func TestSMTPSenderContextTimeout(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err == nil {
			defer conn.Close()
			time.Sleep(50 * time.Millisecond)
		}
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	_, err = strconv.Atoi(port)
	require.NoError(t, err)

	sender := NewSMTPSenderWithTimeout(host, port, "", "", "sender@example.local", 10*time.Millisecond)
	require.Error(t, sender.Send(context.Background(), "to@example.local", "Subject", "Body"))
}
