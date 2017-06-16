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

var HtmlTemplateName = "body.html"
var TextTemplateName = "body.txt"

type Mailing struct {
	MailingID        string
	HtmlBodyTemplate string
	TextBodyTemplate string
}

func NewMailingWithTemplates(mailingID string, templateDir string) (*Mailing, error) {
	m := &Mailing{
		MailingID: mailingID,
	}

	htmlTemplate := filepath.Join(templateDir, HtmlTemplateName)
	if exists(htmlTemplate) {
		t, err := ioutil.ReadFile(htmlTemplate)
		m.HtmlBodyTemplate = string(t)
		if err != nil {
			return nil, err
		}
	}

	textTemplate := filepath.Join(templateDir, TextTemplateName)
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

func (m *Mailing) PopulateMessages(out chan<- Message, errors chan<- MessageProcessingResult, jsonStream io.Reader) error {
	d := json.NewDecoder(jsonStream)

	var globalData interface{}
	err := d.Decode(&globalData)
	if err != nil {
		return err
	}

	mf := &messageFunctions{}

	var htmlT, textT *template.Template
	if m.HtmlBodyTemplate != "" {
		htmlT, err = template.New("htmlBody").
			Funcs(mf.Funcs()).
			Parse(m.HtmlBodyTemplate)
		if err != nil {
			println("Decode z")

			return err
		}
	}

	if m.TextBodyTemplate != "" {
		textT, err = template.New("textBody").
			Funcs(mf.Funcs()).
			Parse(m.TextBodyTemplate)
		if err != nil {
			return err
		}
	}

	go m.doTemplating(out, errors, d, htmlT, textT, globalData, mf)
	return nil
}

func (m *Mailing) doTemplating(out chan<- Message, errors chan<- MessageProcessingResult, d *json.Decoder, htmlT, textT *template.Template, globalData interface{}, mf *messageFunctions) {
	defer close(out)
	defer close(errors)

	data := struct {
		Global  interface{}
		Message map[string]interface{}
	}{Global: globalData}

	offset := 0
	for {
		var parsedata interface{}
		err := d.Decode(&parsedata)
		if err == io.EOF {
			break
		}
		if err != nil {
			errors <- MessageProcessingResult{
				m.MailingID,
				"",
				fmt.Errorf("error parsing json stream at offset %v with %v", offset, err),
			}
			return
		}

		var messageList []interface{}
		switch d := parsedata.(type) {
		case []interface{}:
			messageList = d
		case map[string]interface{}:
			messageList = append(messageList, d)
		default:
			errors <- MessageProcessingResult{
				m.MailingID,
				"",
				fmt.Errorf("message data of wrong type at offset %v", offset),
			}
			continue
		}

		for _, item := range messageList {
			offset++
			var ok bool
			if data.Message, ok = item.(map[string]interface{}); !ok {
				errors <- MessageProcessingResult{
					m.MailingID,
					"",
					fmt.Errorf("message data of wrong type at offset %v", offset),
				}
				continue
			}
			data.Message = item.(map[string]interface{})
			msg := Message{MailingID: m.MailingID}
			mf.SetTarget(&msg)
			if htmlT != nil {
				w := bytes.NewBuffer(nil)
				err := htmlT.Execute(w, data)

				if err != nil {
					errors <- MessageProcessingResult{
						m.MailingID,
						"",
						fmt.Errorf("html template execution failed for message %v:%v with %v", m.MailingID, msg.To, err),
					}
					continue
				}
				msg.HtmlBody = w.String()
			}
			if textT != nil {
				w := bytes.NewBuffer(nil)
				err := textT.Execute(w, data)
				if err != nil {
					errors <- MessageProcessingResult{
						m.MailingID,
						"",
						fmt.Errorf("text template execution failed for message %v:%v with %v", m.MailingID, msg.To, err),
					}
					continue
				}
				msg.TextBody = w.String()
			}

			if msg.To == "" {
				errors <- MessageProcessingResult{
					m.MailingID,
					"",
					fmt.Errorf("message no address set at offset %v", offset),
				}
				continue
			}

			out <- msg
		}
	}
}

func exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

type messageFunctions struct {
	target *Message
}

func (mf *messageFunctions) SetTarget(message *Message) {
	mf.target = message
}

func (mf *messageFunctions) Funcs() template.FuncMap {
	return template.FuncMap{
		"subject": func(subject string, dropArguments ...interface{}) string {
			mf.target.Subject = subject
			return ""
		},
		"to": func(to string, dropArguments ...interface{}) string {
			mf.target.To = to
			return ""
		},
		"from": func(from string, dropArguments ...interface{}) string {
			mf.target.From = from
			return ""
		},
	}
}
