package mail

import (
	"bytes"
	. "github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_PopulateMessages(t *testing.T) {
	dir, cleanup := createTemplateDir(
		`{{ to .Message.to | subject .Global.subject | from "from@example.org" -}}
<html><body>Hello {{.Message.name}}</body></html>`,
		`Hello {{.Message.name}}`,
	)
	defer cleanup()

	m, err := NewMailingWithTemplates("id42", dir)
	NoError(t, err)

	jsonStream := bytes.NewBufferString(
		`{"subject": "The Subject"}
[{"to": "to1@example.org", "name": "Arthur"},{"to": "to2@example.org", "name": "Zappod"}]
{}
{"to": "to3@example.org", "name": "Marvin"}`)

	out := make(chan Message, 10)
	errors := make(chan MessageProcessingResult, 10)

	err = m.PopulateMessages(out, errors, jsonStream)
	NoError(t, err)
	time.Sleep(time.Millisecond * 10)

	if len(out) != 3 {
		Failf(t, "wrong number of messages", "expected: 3, but was: %v", len(out))
		return
	}

	msg := <-out
	Equal(t, "id42", msg.MailingID)
	Equal(t, "from@example.org", msg.From)
	Equal(t, "The Subject", msg.Subject)

	Equal(t, "to1@example.org", msg.To)
	Equal(t, "<html><body>Hello Arthur</body></html>", msg.HtmlBody)
	Equal(t, "Hello Arthur", msg.TextBody)

	msg = <-out
	Equal(t, "id42", msg.MailingID)
	Equal(t, "from@example.org", msg.From)
	Equal(t, "The Subject", msg.Subject)

	Equal(t, "to2@example.org", msg.To)
	Equal(t, "<html><body>Hello Zappod</body></html>", msg.HtmlBody)
	Equal(t, "Hello Zappod", msg.TextBody)

	msg = <-out
	Equal(t, "id42", msg.MailingID)
	Equal(t, "from@example.org", msg.From)
	Equal(t, "The Subject", msg.Subject)

	Equal(t, "to3@example.org", msg.To)
	Equal(t, "<html><body>Hello Marvin</body></html>", msg.HtmlBody)
	Equal(t, "Hello Marvin", msg.TextBody)

	Equal(t, len(errors), 1)
	error := <-errors
	Contains(t, error.Err.Error(), "html template execution failed")

}

func createTemplateDir(htmlTemplate, textTemplate string) (dir string, cleanup func()) {
	dir, err := ioutil.TempDir("", "mailigotest")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filepath.Join(dir, HtmlTemplateName), []byte(htmlTemplate), 0666)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filepath.Join(dir, TextTemplateName), []byte(textTemplate), 0666)
	if err != nil {
		panic(err)
	}

	cleanup = func() {
		os.RemoveAll(dir)
	}
	return dir, cleanup
}
