package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Mailing struct {
	MailingID        string
	HtmlBodyTemplate string
	TextBodyTemplate string
}

func NewMailingWithTemplates(mailingID string, templateDir string) (*Mailing, error) {
	m := &Mailing{
		MailingID: mailingID,
	}

	htmlTemplate := filepath.Join(templateDir, "body.html")
	if exists(htmlTemplate) {
		t, err := ioutil.ReadFile(htmlTemplate)
		m.HtmlBodyTemplate = string(t)
		if err != nil {
			return nil, err
		}
	}

	textTemplate := filepath.Join(templateDir, "body.txt")
	if exists(htmlTemplate) {
		t, err := ioutil.ReadFile(textTemplate)
		m.TextBodyTemplate = string(t)
		if err != nil {
			return nil, err
		}
	}

	if len(m.HtmlBodyTemplate) == 0 && len(m.TextBodyTemplate) == 0 {
		return nil, fmt.Errorf("no templates fount at %q and %q", htmlTemplate, textTemplate)
	}

	return m, nil
}

func (m *Mailing) PopulateMessages(out chan<- Message, jsonStream io.Reader) error {
	d := json.NewDecoder(jsonStream)

	data := struct {
		Global  interface{}
		Message interface{}
	}{}

	err := d.Decode(&data.Global)
	if err != nil {
		return err
	}

	var htmlT, textT *template.Template
	if m.HtmlBodyTemplate != "" {
		htmlT, err = template.New("htmlBody").Parse(m.HtmlBodyTemplate)
		if err != nil {
			return err
		}
	}

	if m.TextBodyTemplate != "" {
		textT, err = template.New("textBody").Parse(m.TextBodyTemplate)
		if err != nil {
			return err
		}
	}

	for {
		var messagesData []interface{}
		err = d.Decode(&messagesData)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		for data.Message = range messagesData {
			msg := Message{MailingID: m.MailingID}
			funcs := templateFuncsForMessage(msg)
			if htmlT != nil {
				textT.Funcs(funcs)
				w := bytes.NewBuffer(nil)
				err := htmlT.Execute(w, data)
				if err != nil {
					return err
				}
				msg.HtmlBody = w.String()
			}
			if textT != nil {
				textT.Funcs(funcs)
				w := bytes.NewBuffer(nil)
				err := textT.Execute(w, data)
				if err != nil {
					return err
				}
				msg.TextBody = w.String()
			}
			out <- msg
		}
	}
	return nil
}

func exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func templateFuncsForMessage(message Message) template.FuncMap {
	return template.FuncMap{
		"subject": func(subject string) {
			message.Subject = subject
		},
		"to": func(to string) {
			message.To = to
		},
		"from": func(from string) {
			message.From = from
		},
	}
}
