package email

import (
	"bufio"
	"context"
	"errors"
	"io"
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

func writeLine(w io.Writer, line string) {
	_, _ = io.WriteString(w, line+"\r\n")
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
