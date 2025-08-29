package internal

import (
	"bytes"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
	"log/slog"
	"os"
	"slices"
	"strings"
)

type Emailer struct {
	host      string
	port      int
	from      string
	whitelist []string
}

func NewEmailer(host string, port int, from string) *Emailer {
	whitelist := os.Getenv("EMAIL_WHITELIST")
	return &Emailer{
		host:      host,
		port:      port,
		from:      from,
		whitelist: strings.Split(whitelist, ","),
	}
}

func (e *Emailer) Send(to []string, shoppingList []Item) error {
	if len(shoppingList) == 0 {
		return fmt.Errorf("shopping list is empty")
	}
	var approvedTo []string
	for _, addr := range to {
		if slices.Contains(e.whitelist, addr) {
			approvedTo = append(approvedTo, addr)
		} else {
			slog.Warn("Email not approved", "email", addr)
		}
	}
	pwd, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
	if !ok {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS env var not set")
	}

	body, err := e.prepareEmailBody(shoppingList)
	if err != nil {
		return fmt.Errorf("error when preparing email body: %v", err)
	}
	fromAddr := e.from
	m := gomail.NewMessage()

	m.SetHeader("From", fromAddr)
	m.SetHeader("To", approvedTo...)
	m.SetHeader("Subject", "New Shopping List")
	m.SetBody("text/html", body)

	// Use app password here
	d := gomail.NewDialer(e.host, e.port, fromAddr, pwd)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error when sending email: %v", err)
	}
	return nil
}

func (e *Emailer) prepareEmailBody(list []Item) (string, error) {
	htmlTpl := `
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: Arial, sans-serif; color: #333; line-height: 1.4; }
    h2 { color: #2c3e50; }
    table { border-collapse: collapse; width: 100%; margin-top: 10px; }
    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
    th { background-color: #f2f2f2; }
    tr:nth-child(even) { background-color: #f9f9f9; }
  </style>
</head>
<body>
  <h2>Shopping List</h2>
  <table>
    <thead>
      <tr>
        <th>Name</th>
        <th>Count</th>
      </tr>
    </thead>
    <tbody>
      {{range .}}
      <tr>
        <td>{{.Name}}</td>
        <td>{{.Count}}</td>
      </tr>
      {{end}}
    </tbody>
</html>
`
	var buf bytes.Buffer
	t := template.Must(template.New("emailHTML").Parse(htmlTpl))
	if err := t.Execute(&buf, list); err != nil {
		return "", err
	}
	return buf.String(), nil
}
