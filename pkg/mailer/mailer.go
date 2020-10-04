package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	config "github.com/valensto/api_apbp/configs"
)

type mailer struct {
	Email string
	PWD   string
	Host  string
	Port  string
}

func NewMailer(c config.Mailer) mailer {
	return mailer{
		Email: c.Email,
		PWD:   c.PWD,
		Host:  c.Host,
		Port:  c.Port,
	}
}

type Mail struct {
	To      []string
	Body    *bytes.Buffer
	Subject string
}

func NewMail() Mail {
	return Mail{}
}

func (m *Mail) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	m.Body = buf
	return nil
}

func (m mailer) Send(mail Mail) error {
	auth := smtp.PlainAuth("", m.Email, m.PWD, m.Host)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := fmt.Sprintf("Subject: %v\n", mail.Subject)
	msg := []byte(subject + mime + mail.Body.String())

	err := smtp.SendMail(m.Host+":"+m.Port, auth, m.Email, mail.To, msg)
	if err != nil {
		return err
	}

	return nil
}

type Sender interface {
	Send(mail Mail) error
}
