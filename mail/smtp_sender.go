package mail

import (
	"gopkg.in/gomail.v2"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	SSL      bool
}

type SMTPSender struct {
	config SMTPConfig
}

func NewSMTPSender(config SMTPConfig) *SMTPSender {
	return &SMTPSender{config}
}

func (m *SMTPSender) Send(messages <-chan Message, results chan<- MessageProcessingResult) error {
	d := gomail.NewDialer(m.config.Host, m.config.Port, m.config.Username, m.config.Password)
	s, err := d.Dial()
	if err != nil {
		return err
	}
	defer s.Close()

	go func() {
		for msg := range messages {
			err := gomail.Send(s, gomailMessage(msg))
			results <- MessageProcessingResult{
				msg.MailingID,
				msg.To,
				err,
			}
		}

		close(results)
	}()

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
