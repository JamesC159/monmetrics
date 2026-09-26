package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// Message is a plain-text email
type Message struct {
	To      string
	Subject string
	Body    string
}

// Sender delivers email messages
type Sender interface {
	Send(ctx context.Context, m Message) error
}

// Config holds SMTP settings; an empty Host selects the log-only sender.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// NewSender returns an SMTP sender, or a LogSender when SMTP is not configured.
func NewSender(cfg Config) Sender {
	if cfg.Host == "" {
		return LogSender{}
	}
	if cfg.Port == "" {
		cfg.Port = "587"
	}
	return &SMTPSender{cfg: cfg}
}

// LogSender prints messages instead of sending them (development)
type LogSender struct{}

func (LogSender) Send(_ context.Context, m Message) error {
	log.Printf("📧 [email not sent: SMTP_HOST unset] to=%s subject=%q\n%s", headerSafe(m.To), headerSafe(m.Subject), m.Body)
	return nil
}

// SMTPSender sends mail via an SMTP server using STARTTLS when offered
type SMTPSender struct {
	cfg Config
}

func (s *SMTPSender) Send(ctx context.Context, m Message) error {
	to, err := mail.ParseAddress(headerSafe(m.To))
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	from, err := mail.ParseAddress(headerSafe(s.cfg.From))
	if err != nil {
		return fmt.Errorf("invalid SMTP_FROM: %w", err)
	}

	d := net.Dialer{Timeout: 10 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(s.cfg.Host, s.cfg.Port))
	if err != nil {
		return err
	}
	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		conn.Close()
		return err
	}
	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if s.cfg.User != "" {
		// PlainAuth refuses to send credentials without TLS except to localhost
		if err := c.Auth(smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)); err != nil {
			return err
		}
	}
	if err := c.Mail(from.Address); err != nil {
		return err
	}
	if err := c.Rcpt(to.Address); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(buildMessage(from.String(), to.String(), m.Subject, m.Body))); err != nil {
		w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	b.WriteString("From: " + headerSafe(from) + "\r\n")
	b.WriteString("To: " + headerSafe(to) + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("UTF-8", headerSafe(subject)) + "\r\n")
	b.WriteString("Date: " + time.Now().UTC().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
	b.WriteString(strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\n", "\r\n"))
	return b.String()
}

// headerSafe strips CR/LF to prevent header injection
func headerSafe(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}
