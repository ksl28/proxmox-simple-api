package main

import (
	"encoding/json"
	"io"
	"net/http"
)

func sendRequest(request *http.Request, resp interface{}, apiToken string) error {
	request.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(response, resp); err != nil {
		return err
	}

	return nil
}
