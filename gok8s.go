package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//required to authenticate to GCP
	"crypto/tls"
	"net/http"

	"io"

	"net/url"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig *string
	if !testInternetConnectivity() {
		fmt.Println("No internet connectivity.Exiting")
		os.Exit(1)
	}
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

	namespace := "openfaas-fn"
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("\nThere are %d pods in the cluster\n", len(pods.Items))
	for _, pod := range pods.Items {
		fmt.Printf("Name of the pod : %s\n", pod.Name)
		for _, container := range pod.Spec.Containers {
			fmt.Printf("\t\t\t Container : %s. Image : %s\n", container.Name, container.Image)
			for _, args := range container.Args {
				fmt.Printf("\t\t\tCommandine arguments for container : %s\n", args)
			}

			for _, cmds := range container.Command {
				fmt.Printf("\t\t\tCommand for container: %s\n", cmds)
			}
		}

	}

	/*
		// Examples for error handling:
		// - Use helper functions like e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message

		pod := "example-xxxxx"
		_, err = clientset.CoreV1().Pods(namespace).Get(pod, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %s in namespace %s: %v\n",
				pod, namespace, statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		}*/

	//time.Sleep(10 * time.Second)

}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func testInternetConnectivity() bool {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: transport}
	request := &http.Request{
		URL:    &url.URL{Host: "www.google.com", Path: "/", Scheme: "https"},
		Method: "GET",
	}
	resp, err := client.Do(request)

	if err != nil {
		panic("Internet connectivity could not be tested. Prerequisites missing")
	}
	if resp.StatusCode != 200 {
		return false
	}
	io.Copy(os.Stdout, resp.Body)
	return true
}
