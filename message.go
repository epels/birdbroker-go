package birdbroker

import (
	"unicode"
	"unicode/utf8"
)

type Message struct {
	Body       string
	Originator string
	Recipient  string
}

func (m *Message) Validate() error {
	if m.Body == "" {
		return ClientError{Reason: "Missing body"}
	}
	if rc := utf8.RuneCountInString(m.Recipient); rc < 1 || rc > 15 {
		return ClientError{Reason: "Recipient length must be 1-15 chars"}
	}
	if !isNumeric(m.Originator) {
		if rc := utf8.RuneCountInString(m.Originator); rc > 11 {
			return ClientError{Reason: "Alphanumeric originator must not be <11 chars"}
		}
	}
	if m.Originator == "" {
		return ClientError{Reason: "Missing originator"}
	}
	return nil
}

func isNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
