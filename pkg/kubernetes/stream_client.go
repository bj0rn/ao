package kubernetes

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"

	"net/url"

	"github.com/gorilla/websocket"
)

var (
	streamClient = websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
)

type CommandParameter struct {
	Pod       string
	Namespace string
	Tty       string
	Stdin     string
	Stdout    string
	Command   string
	Stderr    string
}

//ExecuteCommand in the running container
func ExecuteCommand(masterUrl, token string, command *CommandParameter) {
	headers := make(http.Header)
	tokenValue := fmt.Sprintf("Bearer %s", token)
	headers.Add("Authorization", tokenValue)
	URL := buildCommandQuery(masterUrl, command)
	fmt.Println(URL)
	conn, _, err := streamClient.Dial(URL, headers)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		fmt.Printf("%s\n", string(message))
	}
}

func buildCommandQuery(masterUrl string, command *CommandParameter) string {
	var buffer bytes.Buffer
	buffer.WriteString("wss://" + masterUrl)
	buffer.WriteString("/api/v1/namespaces/" + command.Namespace + "/pods/" + command.Pod + "/exec?")

	u, err := url.Parse(buffer.String())
	if err != nil {
		fmt.Println("Feil format på url")
	}
	q := u.Query()
	q.Add("command", command.Command)
	if command.Tty != "" {
		q.Add("tty", command.Tty)
	}
	if command.Stdin != "" {
		q.Add("stdin", command.Stdin)
	}
	if command.Stderr != "" {
		q.Add("stderr", command.Stderr)
	}
	if command.Stdout != "" {
		q.Add("stdout", command.Stdout)
	}
	u.RawQuery = q.Encode()

	return u.String()
}
