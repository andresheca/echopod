package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
)

const templ = `
Hostname: {{ .Hostname}}
IP: {{ .IP}}
{{- if .Namespace}}
Namespace: {{ .Namespace}}
{{- end}}
URI: {{ .URI}}
Method: {{ .Method}}
RemoteAddr: {{ .RemoteAddr}}
`

func main() {
	http.HandleFunc("/", handleRoot)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	type PodInfo struct {
		Hostname   string
		IP         string
		Namespace  string
		URI        string
		Method     string
		RemoteAddr string
		Response   string
	}
	var info PodInfo
	var err error

	info.Hostname, err = os.Hostname()
	if err != nil {
		log.Fatal("Error getting hostname:", err)
	}

	info.IP, err = getIP()
	if err != nil {
		log.Fatal("Error getting IP address:", err)
	}

	info.Namespace, err = getNamespace()
	if err != nil {
		log.Fatal("Error getting Pod namespace:", err)
	}

	info.URI = r.RequestURI
	info.RemoteAddr = r.RemoteAddr
	info.Method = r.Method
	report := template.Must(template.New("podinfo").Parse(templ))

	log.Printf("%s - %s %s %s", r.Host, r.RemoteAddr, r.Method, r.URL)
	report.Execute(w, info)
}

func getIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("Error getting Interface Addrs:\n %v", err)
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func getNamespace() (string, error) {
	var namespace string
	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	if _, err := os.Stat(namespaceFile); err == nil {
		file, err := os.Open(namespaceFile)
		if err != nil {
			return "", err
		}
		defer file.Close()

		scaner := bufio.NewScanner(file)
		for scaner.Scan() {
			namespace += scaner.Text()
		}
		return namespace, nil
	}
	return "", nil
}
