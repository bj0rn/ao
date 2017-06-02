package serverapi_v2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skatteetaten/aoc/pkg/jsonutil"
	"github.com/skatteetaten/aoc/pkg/openshift"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// Structs to represent return data from the API interface

type ApplicationId struct {
	EnvironmentName string `json:"environmentName"`
	ApplicationName string `json:"applicationName"`
}

type OpenShiftResponse struct {
	Kind          string `json:"kind"`
	OperationType string `json:"operationType"` // CREATED, UPDATE eller NONE
	Payload       struct {
		Kind string `json:"payload"`
	} `json:"payload"`
	ResponseBody json.RawMessage `json:"responseBody"`
}

type DeploymentDescriptor struct {
	TemplateFile string            `json:"templateFile"`
	Template     string            `json:"template"`
	Parameters   map[string]string `json:"parameters"`
}

type AuroraDeploymentConfig struct {
	SchemaVersion        string               `json:"schemaVersion"`
	Affiliation          string               `json:"affiliation"`
	Cluster              string               `json:"cluster"`
	Type                 string               `json:"type"`
	Name                 string               `json:"name"`
	EnvName              string               `json:"envName"`
	Groups               []string             `json:"groups"`
	Users                []string             `json:"users"`
	Replicas             int                  `json:"replicas"`
	Secrets              map[string]string    `json:"secrets"`
	Config               map[string]string    `json:"config"`
	GroupId              string               `json:"groupId"`
	ArtifactId           string               `json:"artifactId"`
	Version              string               `json:"version"`
	Route                bool                 `json:"route"`
	DeploymentStrategy   string               `json:"deploymentStrategy"`
	DeploymentDescriptor DeploymentDescriptor `json:"deploymentDescriptor"`
}

type ApplicationResult struct {
	ApplicationId     ApplicationId          `json:"applicationId"`
	AuroraDc          AuroraDeploymentConfig `json:"auroraDc"`
	OpenShiftResponse OpenShiftResponse      `json:"openShiftResponse"`
}

type Response struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Items   []json.RawMessage `json:"items"`
	Count   int               `json:"count"`
}

type ResponseItemError struct {
	Application string `json:"application"`
	Environment string `json:"environment"`
	Messages    []struct {
		Message string `json:"message"`
		Field   struct {
			Path   string `json:"path"`
			Value  string `json:"value"`
			Source string `json:"source"`
		} `json:"field"`
	} `json:"messages"`
}

type AuroraConfig struct {
	Files   map[string]json.RawMessage `json:"files"`
	Secrets map[string]json.RawMessage `json:"secrets"`
}

type PingResult struct {
	Items []struct {
		Result struct {
			Status     string `json:"status"`
			Dnsname    string `json:"dnsname"`
			ResolvedIp string `json:"resolvedIp"`
			Port       string `json:"port"`
			Message    string `json:"message"`
		} `json:"result"`
		PodIp    string `json:"podIp"`
		HostIp   string `json:"hostIp"`
		HostName string
	} `json:"items"`
}

const apiNotInstalledResponse = "Application is not available"
const localhostAddress = "localhost"
const localhostPort = "8080"

func ParsePingResult(responseString string) (PingResult PingResult, err error) {
	var responseData []byte
	responseData = []byte(responseString)
	err = json.Unmarshal(responseData, &PingResult)
	if err != nil {
		return
	}

	return
}

func ParseResponse(responseString string) (response Response, err error) {
	var responseData []byte
	responseData = []byte(responseString)
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		return
	}

	return
}

func ResponseItems2ApplicationResults(response Response) (applicationResults []ApplicationResult, err error) {
	applicationResults = make([]ApplicationResult, len(response.Items))

	for item := range response.Items {
		err = json.Unmarshal([]byte(response.Items[item]), &applicationResults[item])
	}
	return
}

func ResponseItems2AuroraConfig(response Response) (auroraConfig AuroraConfig, err error) {

	if response.Count > 1 {
		err = errors.New("Internal error: Multiple items not supported in AOC")
		return
	}
	for item := range response.Items {
		err = json.Unmarshal([]byte(response.Items[item]), &auroraConfig)
	}
	return
}

func ApplicationResult2MessageString(applicationResult ApplicationResult) (output string, err error) {

	output +=
		//applicationResult.ApplicationId.ApplicationName +
		applicationResult.AuroraDc.GroupId + "/" + applicationResult.AuroraDc.ArtifactId + "-" + applicationResult.AuroraDc.Version +
			" deployed in " + applicationResult.AuroraDc.Cluster + "/" + applicationResult.AuroraDc.EnvName
	return
}

func ResponsItems2MessageString(response Response) (output string, err error) {
	if response.Message != "" {
		output = response.Message + ": "
	}

	for item := range response.Items {
		var responseItemError ResponseItemError
		err = json.Unmarshal([]byte(response.Items[item]), &responseItemError)
		if err != nil {
			return
		}
		output = output + "\n\t" + responseItemError.Environment + "/" + responseItemError.Application + ":"

		for message := range responseItemError.Messages {
			output = output + "\n\t\t" + responseItemError.Messages[message].Field.Path + " (" +
				responseItemError.Messages[message].Field.Value + ") in " + responseItemError.Messages[message].Field.Source
			output = output + "\n\t\t\t" + responseItemError.Messages[message].Message
		}
	}
	return
}

func getConsoleAddress(clusterName string) (consoleAddress string) {
	//consoleAddress = "http://console-aurora." + clusterName + ".paas.skead.no"
	consoleAddress = "http://console-paas-espen-dev." + clusterName + ".paas.skead.no"
	return
}

func CallConsole(apiEndpoint string, arguments string, verbose bool, debug bool, openshiftConfig *openshift.OpenshiftConfig) (result json.RawMessage, err error) {
	apiCluster, err := openshiftConfig.GetApiCluster()
	consoleAddress := getConsoleAddress(apiCluster.Name)
	token := apiCluster.Token

	url := consoleAddress + "/public/" + apiEndpoint
	if arguments != "" {
		url += "?" + arguments
	}
	if debug {
		fmt.Print("Sending request to Console at " + url + "...")
	}
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			fmt.Println("FAIL.  Error connecting to Console service")
		}
		err = errors.New(fmt.Sprintf("Error connecting to the Console service on %v: %v", url, err))
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	output := string(body)

	if resp.StatusCode == http.StatusOK {
		if debug {
			fmt.Println("OK")
		}
	} else {
		if debug {
			fmt.Println("ERROR: " + resp.Status)
		}
		if resp.StatusCode == http.StatusGatewayTimeout {
			return nil, errors.New("Ping request timed out")
		} else {
			return nil, errors.New(resp.Status)
		}
	}

	if debug {
		fmt.Println("Response status: " + strconv.Itoa(resp.StatusCode))
		if jsonutil.IsLegalJson(output) {
			fmt.Println(jsonutil.PrettyPrintJson(output))
		} else {
			fmt.Println(output)
		}

	}
	result = json.RawMessage(output)
	return
}

// Check for valid login, that is we have a configuration with at least one reachable cluster
func ValidateLogin(openshiftConfig *openshift.OpenshiftConfig) (output bool) {
	var openshiftCluster *openshift.OpenshiftCluster
	openshiftCluster, _ = openshiftConfig.GetApiCluster()
	if openshiftCluster != nil {
		if !openshiftCluster.HasValidToken() {
			return false
		}
	}
	return true
}

func GetApiAddress(clusterName string, localhost bool) (apiAddress string) {
	if localhost {
		apiAddress = "http://" + localhostAddress + ":" + localhostPort
	} else {
		apiAddress = "http://boober-aos-bas-dev." + clusterName + ".paas.skead.no"
	}
	return
}
func GetApiSetupUrl(clusterName string, apiEndpont string, localhost bool, dryrun bool, apiAddress string) string {
	if apiAddress == "" {
		apiAddress = GetApiAddress(clusterName, localhost)
	}
	return apiAddress + apiEndpont
}

func CallApi(httpMethod string, apiEndpoint string, combindedJson string, showConfig bool, showObjects bool, api bool, localhost bool, verbose bool,
	openshiftConfig *openshift.OpenshiftConfig, dryRun bool, debug bool, apiAddress string) (outputMap map[string]string, err error) {
	//var openshiftConfig *openshift.OpenshiftConfig
	var apiCluster *openshift.OpenshiftCluster

	outputMap = make(map[string]string)
	if localhost {
		apiAddress = GetApiAddress("", true)
	}
	if apiAddress != "" {
		var token string = ""
		apiCluster, err = openshiftConfig.GetApiCluster()
		if apiCluster != nil {
			token = apiCluster.Token
		}
		output, err := callApiInstance(httpMethod, combindedJson, verbose,
			GetApiSetupUrl(apiAddress, apiEndpoint, localhost, dryRun, apiAddress), token, dryRun, debug)
		outputMap[openshiftConfig.Clusters[0].Name] = output
		if err != nil {
			return outputMap, err
		}
	} else {
		var errorString string
		var newlineErr string
		for i := range openshiftConfig.Clusters {
			if openshiftConfig.Clusters[i].Reachable {
				if !api || openshiftConfig.Clusters[i].Name == openshiftConfig.APICluster {
					output, err := callApiInstance(httpMethod, combindedJson, verbose,
						GetApiSetupUrl(openshiftConfig.Clusters[i].Name, apiEndpoint, localhost, dryRun, apiAddress),
						openshiftConfig.Clusters[i].Token, dryRun, debug)
					if output != "" {
						outputMap[openshiftConfig.Clusters[i].Name] = output

						if err != nil {
							errorString += newlineErr + err.Error()
							newlineErr = "\n"
						}
					}
				}
			}
		}
		if errorString != "" {
			err = errors.New(errorString)
			fmt.Println("ERROR: " + errorString)
			return
		}
	}
	return
}

func callApiInstance(httpMethod string, combindedJson string, verbose bool, url string, token string, dryRun bool, debug bool) (output string, err error) {

	if verbose {
		fmt.Print("Sending config to Boober at " + url + "... ")
	}

	var jsonStr = []byte(combindedJson)

	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return "", errors.New(fmt.Sprintf("Internal error in NewRequest: %v", err))
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Add("dryrun", fmt.Sprintf("%v", dryRun))
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		if verbose {
			fmt.Println("FAIL.  Error connecting to Boober service")
		}
		fmt.Println("Debug: Error in client.Do")
		return "", errors.New(fmt.Sprintf("Error connecting to the Boober service on %v: %v", url, err))
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	output = string(body)

	if debug {
		fmt.Println("Response status: " + strconv.Itoa(resp.StatusCode))
		if jsonutil.IsLegalJson(output) {
			fmt.Println(jsonutil.PrettyPrintJson(output))
		} else {
			fmt.Println(output)
		}

	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusBadRequest) {

		var errorstring string
		if !strings.Contains(output, apiNotInstalledResponse) {
			errorstring = fmt.Sprintf("Internal error on %v: %v", url, output)
		}
		if verbose {
			if strings.Contains(output, apiNotInstalledResponse) {
				fmt.Println("WARN.  Boober not available")
			} else {
				fmt.Println("FAIL.  Internal error")
			}
		}
		err = errors.New(fmt.Sprintf(errorstring))
		return
	}

	if resp.StatusCode == http.StatusBadRequest {
		// We have a validation situation, give error
		if verbose {
			fmt.Println("FAIL.  Error in configuration")
		}

		err = errors.New(fmt.Sprintf(output))

		return
	}

	if verbose {
		fmt.Println("OK")
	}

	return
}
