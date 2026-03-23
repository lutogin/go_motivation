package email

import (
	"fmt"
	"html"
	"net/smtp"
	"regexp"

	"github.com/aluto/go-motivation/internal/config"
	"github.com/aluto/go-motivation/internal/entity"
	log "github.com/sirupsen/logrus"
)

// QuoteEmailSubject is the production subject line for quote emails.
const QuoteEmailSubject = "GO MOTIVATION BOT"

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

type Sender struct {
	host    string
	port    int
	user    string
	pass    string
	from    string
	enabled bool
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{
		host:    cfg.SMTPHost,
		port:    cfg.SMTPPort,
		user:    cfg.SMTPUser,
		pass:    cfg.SMTPPass,
		from:    cfg.SMTPFrom,
		enabled: cfg.SMTPEnabled(),
	}
}

func (s *Sender) Enabled() bool {
	return s.enabled
}

func (s *Sender) SendQuote(to string, q *entity.Quote) error {
	if !s.enabled {
		return nil
	}

	raw := QuoteEmailRFC822(s.from, to, q)

	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, raw); err != nil {
		log.Errorf("send email to %s: %v", to, err)
		return err
	}

	log.Infof("email sent to %s", to)
	return nil
}

// QuoteEmailRFC822 builds the raw SMTP message body (HTML quote email).
func QuoteEmailRFC822(from, to string, q *entity.Quote) []byte {
	body := FormatQuoteHTML(q)
	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"Content-Transfer-Encoding: 8bit\r\n"+
		"\r\n"+
		"%s",
		from, to, QuoteEmailSubject, body)
	return []byte(msg)
}

// FormatQuoteHTML renders the same HTML layout as production quote emails.
func FormatQuoteHTML(q *entity.Quote) string {
	out := `<div style="font-family:Georgia,serif;max-width:500px;margin:0 auto;padding:30px;border:1px solid #e0e0e0;border-radius:8px">`
	out += fmt.Sprintf(`<p style="font-size:18px;font-style:italic;color:#333;line-height:1.6">"%s"</p>`, html.EscapeString(q.Text))

	if q.Author != "" {
		out += fmt.Sprintf(`<p style="text-align:right;color:#666;font-weight:bold">— %s</p>`, html.EscapeString(q.Author))
	}

	if q.Notes != "" {
		out += fmt.Sprintf(`<p style="color:#888;font-size:14px;border-top:1px solid #eee;padding-top:10px;margin-top:15px">📝 %s</p>`, html.EscapeString(q.Notes))
	}

	out += `</div>`
	return out
}
