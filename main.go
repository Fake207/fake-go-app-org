package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Metadata estructura básica para Cloud Run
type Metadata struct {
	Service  string `json:"service"`
	Revision string `json:"revision"`
	Project  string `json:"project"`
	Region   string `json:"region"`
}

func main() {
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Servidor iniciado en puerto %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	metadata, err := fetchMetadata()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	publicIP, err := getPublicIP()
	if err != nil {
		publicIP = "Unavailable"
	}

	response := fmt.Sprintf("Hello World!\nPublic IP: %s\nService Metadata:\n- Service: %s\n- Revision: %s\n- Project: %s\n- Region: %s\n",
		publicIP, metadata.Service, metadata.Revision, metadata.Project, metadata.Region)

	fmt.Fprint(w, response)
}

func fetchMetadata() (Metadata, error) {
	// Simple metadata (Cloud Run)
	meta := Metadata{}

	service, err := fetchMetaField("K_SERVICE")
	if err != nil {
		service = "local"
	}
	revision, err := fetchMetaField("K_REVISION")
	if err != nil {
		revision = "local"
	}
	project, err := fetchMetaField("GOOGLE_CLOUD_PROJECT")
	if err != nil {
		project = "local"
	}
	region, err := fetchRegion()
	if err != nil {
		region = "local"
	}

	meta.Service = service
	meta.Revision = revision
	meta.Project = project
	meta.Region = region

	return meta, nil
}

// Obtener variables de entorno de Cloud Run
func fetchMetaField(envVar string) (string, error) {
	val := os.Getenv(envVar)
	if val == "" {
		return "", fmt.Errorf("environment variable %s not set", envVar)
	}
	return val, nil
}

// Obtener región desde Metadata server
func fetchRegion() (string, error) {
	req, _ := http.NewRequest("GET",
		"http://metadata.google.internal/computeMetadata/v1/instance/region", nil)
	req.Header.Set("Metadata-Flavor", "Google")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Obtener IP pública externa
func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var ip struct {
		IP string `json:"ip"`
	}
	err = json.NewDecoder(resp.Body).Decode(&ip)
	if err != nil {
		return "", err
	}
	return ip.IP, nil
}
