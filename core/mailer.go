package core

import "context"

// MailAttachment represents a file attached to an email.
type MailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Mail holds all data needed to send an email.
type Mail struct {
	From        string
	To          []string
	CC          []string
	BCC         []string
	Subject     string
	HTMLBody    string
	TextBody    string
	Attachments []MailAttachment
}

// Mailer is the contract for email sending backends (e.g. ss-keel-mail).
type Mailer interface {
	Send(ctx context.Context, mail Mail) error
}
