package mail

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"time"
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

	go func() {
		d := gomail.NewDialer(m.config.Host, m.config.Port, m.config.Username, m.config.Password)
		s, err := d.Dial()
		if err != nil {
			fmt.Printf("error on connect %v", err)
		}

		counter := 0
		for msg := range messages {
			//fmt.Printf("msg: %+v\n", msg)
			err := gomail.Send(s, gomailMessage(msg))
			results <- MessageProcessingResult{
				msg.MailingID,
				msg.To,
				err,
			}
			counter++
			if err != nil || counter%20 == 0 {
				fmt.Printf("reconnect ...\n")
				s.Close()
				time.Sleep(time.Second)
				d = gomail.NewDialer(m.config.Host, m.config.Port, m.config.Username, m.config.Password)
				s, err = d.Dial()
				if err != nil {
					fmt.Printf("error on reconnect %v\n", err)
				}
			}
			time.Sleep(150 * time.Millisecond)
		}
		close(results)
		s.Close()
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
