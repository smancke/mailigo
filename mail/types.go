package mail

type Sender interface {
	Send(messages <-chan Message, results chan<- MessageProcessingResult) error
}

type MessageProcessingResult struct {
	MailingId string
	To        string
	Err       error
}

type Message struct {
	MailingID string
	From      string
	To        string
	Subject   string
	HtmlBody  string
	TextBody  string
}
