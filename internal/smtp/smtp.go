package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/relationskat/auth-service/internal/config"
	"go.uber.org/zap"
)

type Mailer struct {
	log  *zap.Logger
	cfg  config.Config
	auth smtp.Auth
}

func New(log *zap.Logger, cfg config.Config) (*Mailer, error) {
	if cfg.SMTP.Host == "" || cfg.SMTP.Port == 0 {
		return nil, fmt.Errorf("smtp: host and port are required")
	}
	if cfg.SMTP.From == "" {
		cfg.SMTP.From = cfg.SMTP.Username
	}

	return &Mailer{
		log:  log.Named("SMTP"),
		cfg:  cfg,
		auth: smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host),
	}, nil
}

func (s *Mailer) Send(ctx context.Context, to []string, subject, htmlBody string) error {
	const op = "smtp.Send"

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTP.Host, s.cfg.SMTP.Port)

	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("%s: dial: %w", op, err)
	}

	if s.cfg.SMTP.Port == 465 {
		conn = tls.Client(conn, &tls.Config{ServerName: s.cfg.SMTP.Host})
	}

	c, err := smtp.NewClient(conn, s.cfg.SMTP.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("%s: new client: %w", op, err)
	}
	defer c.Close()

	if s.cfg.SMTP.Port != 465 {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: s.cfg.SMTP.Host}); err != nil {
				return fmt.Errorf("%s: starttls: %w", op, err)
			}
		}
	}

	if ok, _ := c.Extension("AUTH"); ok {
		if err := c.Auth(s.auth); err != nil {
			return fmt.Errorf("%s: auth: %w", op, err)
		}
	}

	if err := c.Mail(s.cfg.SMTP.From); err != nil {
		return fmt.Errorf("%s: mail from: %w", op, err)
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("%s: rcpt %s: %w", op, rcpt, err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("%s: data: %w", op, err)
	}
	if _, err := w.Write(s.buildMessage(to, subject, htmlBody)); err != nil {
		return fmt.Errorf("%s: write: %w", op, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("%s: close data: %w", op, err)
	}

	if err := c.Quit(); err != nil {
		return fmt.Errorf("%s: quit: %w", op, err)
	}

	s.log.Info("email sent", zap.Strings("to", to), zap.String("subject", subject))
	return nil
}

func (s *Mailer) buildMessage(to []string, subject, htmlBody string) []byte {
	var b strings.Builder
	b.WriteString("From: " + s.cfg.SMTP.From + "\r\n")
	b.WriteString("To: " + strings.Join(to, ", ") + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(htmlBody)
	return []byte(b.String())
}
