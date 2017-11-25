package kubernetes

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

const (
	PODS_PATTERN_ALL    = "%s/api/v1/namespaces/%s/pods"
	COMMAND_PATTERN     = "%s/api/v1/namespaces/%s/pods/%s/exec?command=env"
	PORTFORWARD_PATTERN = "%s/api/v1/namespaces/{namespace}/pods/{name}/portforward"
)

var (
	transport = http.Transport{
		Dial:            dialTimeout,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client = http.Client{
		Transport: &transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

var timeout = time.Duration(1 * time.Second)

func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

//GetVersions get applications version of all running pods
func GetVersions(cmd, token, namespace string) string {
	//get every pod (Names only)
	podList := fmt.Sprintf(PODS_PATTERN_ALL, "https://master-api.theopsh.net:8443", namespace)

	resp, err := get(podList, token)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	command := fmt.Sprintf(COMMAND_PATTERN, "https://master-api.theopsh.net:8443", namespace, "node-1-tkx5f")
	resp, err = get(command, token)
	if err != nil {
		return ""
	}

	c := &CommandParameter{
		Pod:       "node-1-tkx5f",
		Namespace: "test",
		Command:   "ls",
		Stdout:    "true",
		Stderr:    "true",
		Stdin:     "true",
	}

	ExecuteCommand("master-api.theopsh.net:8443", token, c)

	//body, _ := ioutil.ReadAll(resp.Body)

	return "Yo"
}

func GetHealthStatus(namespace, token string) {

	//get every pod
	//
	//response := get(url, token)

	//filter out the names (Only unique applications)

	//Portforward

	//get health and collect results

	//return results
}

func post(url, payload, token string) (string, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	tokenValue := fmt.Sprintf("Bearer %s", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", tokenValue)

	return "", nil
}

func get(url, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	tokenValue := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", tokenValue)
	return client.Do(req)
}

func patch(url, payload, token string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return nil, err
	}

	tokenValue := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", tokenValue)

	return client.Do(req)
}

func put(url, token, payload string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return nil, err
	}
	tokenValue := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", tokenValue)

	return client.Do(req)
}
