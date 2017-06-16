package mail

import (
	"github.com/smancke/mailigo/logging"
	"io"
	"math/rand"
	"path/filepath"
	"time"
)

type MailingManager struct {
	templateBaseDir string
	sender          Sender
}

func NewMailingManager(templateBaseDir string, sender Sender) *MailingManager {
	return &MailingManager{
		templateBaseDir: templateBaseDir,
		sender:          sender,
	}
}

func (mm *MailingManager) CreateMailing(templateName string, gobalDataJson string) (mailingId string, err error) {
	return
}

func (mm *MailingManager) DoMailing(templateName string, jsonStream io.Reader) (string, error) {
	id := randStringBytes(20)
	templateDir := filepath.Join(mm.templateBaseDir, templateName)
	mailing, err := NewMailingWithTemplates(id, templateDir)
	if err != nil {
		return id, err
	}

	out := make(chan Message, 0)
	populateResult := make(chan MessageProcessingResult, 0)
	go func() {
		for res := range populateResult {
			if res.Err != nil {
				logging.Logger.WithError(res.Err).Errorf("error processing message %v:%v", res.MailingId, res.To)
			} else {
				logging.Logger.Infof("processed message %v:%v", res.MailingId, res.To)
			}
		}
	}()

	err = mailing.PopulateMessages(out, populateResult, jsonStream)
	if err != nil {
		return id, err
	}

	go func() {
		sendResult := make(chan MessageProcessingResult, 10)
		err = mm.sender.Send(out, sendResult)
		if err != nil {
			logging.Logger.WithError(err).Errorf("error sending messages")
		}

		for res := range sendResult {
			if res.Err != nil {
				logging.Logger.WithError(res.Err).Errorf("error processing message %v:%v", res.MailingId, res.To)
			} else {
				logging.Logger.Infof("processed message %v:%v", res.MailingId, res.To)
			}
		}
	}()

	return id, nil
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
