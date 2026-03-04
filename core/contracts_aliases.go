package core

import "github.com/slice-soft/ss-keel-core/core/contracts"

// These aliases keep the public core API stable while organizing contracts
// under core/contracts.
type (
	Guard = contracts.Guard

	Cache = contracts.Cache

	StorageObject = contracts.StorageObject
	Storage       = contracts.Storage

	MailAttachment = contracts.MailAttachment
	Mail           = contracts.Mail
	Mailer         = contracts.Mailer

	Message        = contracts.Message
	MessageHandler = contracts.MessageHandler
	Publisher      = contracts.Publisher
	Subscriber     = contracts.Subscriber

	Job       = contracts.Job
	Scheduler = contracts.Scheduler

	RequestMetrics   = contracts.RequestMetrics
	MetricsCollector = contracts.MetricsCollector

	Span   = contracts.Span
	Tracer = contracts.Tracer

	Translator = contracts.Translator
)
