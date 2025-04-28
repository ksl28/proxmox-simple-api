package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func convertJSON() ([]PVEConnectionObject, error) {
	envVar, ok := os.LookupEnv("OBJECTS_JSON")
	if !ok {
		log.Println("No data found in OBJECTS_JSON variable")
		return nil, fmt.Errorf("no data found in OBJECTS_JSON variable")
	}

	var objects []PVEConnectionObject
	if err := json.Unmarshal([]byte(envVar), &objects); err != nil {
		var singleObject PVEConnectionObject
		if err := json.Unmarshal([]byte(envVar), &singleObject); err != nil {
			return objects, err
		} else {
			objects = []PVEConnectionObject{singleObject}
		}
	}
	return objects, nil
}

var httpClient *http.Client

func main() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	allNodes, err := convertJSON()
	if err != nil {
		log.Fatalf("Failed to decode the JSON content - %v", err)
	}

	for _, obj := range allNodes {
		v, _ := testHostPort(obj.Name, obj.Port)
		if !v {
			log.Printf("%v is not listening on port %d", obj.Name, obj.Port)
		} else {
			log.Printf("%v is listening on port %d", obj.Name, obj.Port)
		}

	}

	apiPort, ok := os.LookupEnv("apiport")
	if !ok {
		apiPort = "8080"
	}

	router := gin.Default()
	router.GET("/api/v1/infrastructure/nodes/summary", quickHostOverview)
	router.GET("/api/v1/infrastructure/nodes/detailed/:name", detailedHostOverview)
	router.GET("/api/v1/infrastructure/nodes/detailed/:name/storage", getNodeStorageOverview)
	router.GET("/api/v1/infrastructure/nodes/detailed/:name/disks", getNodeDiskOverview)
	router.GET("/api/v1/virtualization/vm/summary", vmSummary)
	router.GET("/api/v1/virtualization/vm/detailed/:parent/:id", vmDetailedOverview)
	router.GET("/api/v1/virtualization/lxc/summary", lxcSummary)

	apiListener := fmt.Sprintf("0.0.0.0:%v", apiPort)
	router.Run(apiListener)
}
