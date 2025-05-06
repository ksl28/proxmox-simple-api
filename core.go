package main

import (
	"encoding/json"
	"fmt"
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

	switch res.StatusCode / 100 {
	case 2:

	case 3:
		return fmt.Errorf("redirect: %s", res.Status)
	case 4:
		return fmt.Errorf("client error: %s", res.Status)
	case 5:
		return fmt.Errorf("server error: %s", res.Status)
	default:
		return fmt.Errorf("unexpected status: %s", res.Status)
	}

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(response, resp); err != nil {
		return err
	}

	return nil
}
