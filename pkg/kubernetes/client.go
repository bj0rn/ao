package kubernetes

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

//GetVersions get applications version of all running pods
func GetVersions(cmd, token, namespace string) string {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	return ""
}

//GetHealthStatus of every application
func GetHealthStatus(namespace, token string) {

	//get every pod
	//
	//response := get(url, token)

	//filter out the names (Only unique applications)

	//Portforward

	//get health and collect results

	//return results
}

//func post(url, payload, token string) (string, error) {
//	req, err := http.NewRequest("POST", url, nil)
//	if err != nil {
//		return "", err
//	}

//	tokenValue := fmt.Sprintf("Bearer %s", token)
//	req.Header.Set("Content-Type", "application/json")
//	req.Header.Add("Authorization", tokenValue)

//	return "", nil
//}

//func get(url, token string) (*http.Response, error) {
//	req, err := http.NewRequest(http.MethodGet, url, nil)
//	if err != nil {
//		return nil, err
//	}

//	tokenValue := fmt.Sprintf("Bearer %s", token)
//	req.Header.Add("Authorization", tokenValue)
//	return client.Do(req)
//}

//func patch(url, payload, token string) (*http.Response, error) {
//	req, err := http.NewRequest(http.MethodPatch, url, nil)
//	if err != nil {
//		return nil, err
//	}

//	tokenValue := fmt.Sprintf("Bearer %s", token)
//	req.Header.Add("Authorization", tokenValue)

//	return client.Do(req)
//}

//func put(url, token, payload string) (*http.Response, error) {
//	req, err := http.NewRequest(http.MethodPut, url, nil)
//	if err != nil {
//		return nil, err
//	}
//	tokenValue := fmt.Sprintf("Bearer %s", token)
//	req.Header.Add("Authorization", tokenValue)

//	return client.Do(req)
//}
