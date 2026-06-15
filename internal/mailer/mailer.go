package mailer

import (
	"bytes"
	"embed"
	ht "html/template"
	tt "text/template"
	"time"

	"github.com/wneessen/go-mail"
)

// the variable of type embed.FS holds our email templates
// the comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.
//
//go:embed "templates"
var templateFS embed.FS

// Mailer struct contains an instance of mail.Client to connect to an SMTP server
// & the From email address details eg "Frank Mwale <frank@example.com>"
type Mailer struct {
	client *mail.Client
	sender string
}

func New(host string, port int, username, password, sender string) (*Mailer, error) {
	// initialize a new mail.Dialer instance with the given SMTP server settings
	// we'll also have a 5-second timeout whenever we send an email
	client, err := mail.NewClient(
		host,
		mail.WithSMTPAuth(mail.SMTPAuthLogin),
		mail.WithPort(port),
		mail.WithUsername(username),
		mail.WithPassword(password),
		mail.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, err
	}

	// return the mailer instance
	mailer := &Mailer{
		client: client,
		sender: sender,
	}

	return mailer, nil
}

// the Sender() method takes the recipient email address as the first parameter
// the name of the file containing the templates, & the dynamic data in the template as any parameter
func (m *Mailer) Send(recipient string, templateFile string, data any) error {
	// use ParseFS from text/template to parse the required
	// template file from the embedded file system
	textTmpl, err := tt.New("").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing
	// the result in a bytes.Buffer variable
	subject := new(bytes.Buffer)
	err = textTmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Execute the named template "plainBody" and store the result in the plainBody variable
	plainBody := new(bytes.Buffer)
	err = textTmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// use ParseFS method from html/template to parse the required
	// template file from the embedded file system
	htmlTmpl, err := ht.New("").ParseFS(templateFS, "/templates/"+templateFile)
	if err != nil {
		return err
	}

	// execute the HTML body template and store the result in the htmlBody variable
	htmlBody := new(bytes.Buffer)
	err = htmlTmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// initialize a new mail.Msg instance
	msg := mail.NewMsg()

	// set recipient
	err = msg.To(recipient)
	if err != nil {
		return err
	}

	// set sender
	err = msg.From(m.sender)
	if err != nil {
		return err
	}

	// set subject headers, the plain text body, and the html body as altenative
	msg.Subject(subject.String())
	msg.SetBodyString(mail.TypeTextPlain, plainBody.String())
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody.String())

	// pass the message to DialAndSend.
	// This opens a connection to the SMTP server, sends the message, then closes the
	// connection.
	return m.client.DialAndSend(msg)
}
