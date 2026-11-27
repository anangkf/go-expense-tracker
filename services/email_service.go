package services

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"

	"go-expense-tracker-api/config"
)

type EmailService struct {
	cfg config.EmailConfig
}

func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{cfg: cfg.Email}
}

// SEND THE VERIFICATION CODE. WITHOUT SMTP_HOST THE MESSAGE IS ONLY LOGGED (DEBUG MODE).
func (s *EmailService) SendVerificationOTP(to, name, code string) error {
	subject := "Your verification code"
	body := fmt.Sprintf(
		"Hi %s,\r\n\r\nYour verification code is %s.\r\nIt expires in %d minutes. If you did not create an account, ignore this email.\r\n",
		stripCRLF(name), code, s.cfg.OTPExpireMinutes,
	)

	if s.cfg.SMTPHost == "" {
		log.Printf("[email disabled] to=%s subject=%q body=%q", to, subject, body)
		return nil
	}
	return s.send(to, subject, body)
}

func (s *EmailService) send(to, subject, body string) error {
	to = stripCRLF(to)
	from := stripCRLF(s.cfg.From)

	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := net.JoinHostPort(s.cfg.SMTPHost, s.cfg.SMTPPort)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	// BOUND THE WHOLE SMTP CONVERSATION
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	client, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return err
	}
	defer client.Close()

	// SMTP CLIENT NEGOTIATES STARTTLS WHEN THE SERVER OFFERS IT; PlainAuth REFUSES UNENCRYPTED REMOTE CONNECTIONS
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: s.cfg.SMTPHost, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if s.cfg.SMTPUser != "" {
		auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func stripCRLF(v string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(v)
}
