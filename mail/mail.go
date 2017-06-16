package mail

import (
	"gopkg.in/gomail.v2"
)

type MailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	SSL      bool
}

type MessageId struct {
}

type SendResult struct {
	MailingId string
	To        string
	Err       error
	Delivered bool
}

type Message struct {
	MailingID string
	From      string
	To        string
	Subject   string
	HtmlBody  string
	TextBody  string
}

type Mailer struct {
	config MailConfig
}

func (m *Mailer) Send(messages <-chan Message, results chan<- SendResult) error {
	d := gomail.NewDialer(m.config.Host, m.config.Port, m.config.Username, m.config.Password)
	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()

	for msg := range messages {
		err := gomail.Send(s, gomailMessage(msg))
		results <- SendResult{
			msg.MailingID,
			msg.To,
			err,
			err != nil,
		}
	}

	return nil
}

func gomailMessage(message Message) *gomail.Message {
	m := gomail.NewMessage()
	m.SetHeader("From", message.From)
	m.SetHeader("To", message.To)
	m.SetHeader("Subject", message.Subject)
	if message.TextBody != "" {
		m.SetBody("text/plain", message.TextBody)
	}
	if message.HtmlBody != "" {
		m.AddAlternative("text/html", message.HtmlBody)
	}
	return m
}
