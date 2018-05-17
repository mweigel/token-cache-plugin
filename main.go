package main

import (
	"fmt"
	"os"
	"strings"
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"encoding/json"
)

type ExecCredential struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Status     map[string]string `json:"status"`
}

func main() {
	tokenServerURL := os.Getenv("TOKEN_SERVER_URL")
	if tokenServerURL == "" {
		fmt.Fprintf(os.Stderr, "TOKEN_SERVER_URL not specified\n")
		os.Exit(1)
	}

	cacheTokens := os.Getenv("CACHE_TOKENS")
	if cacheTokens == "" {
		cacheTokens = "false"
	}

	// Attempt to read a previously cached token before prompting for a username and password.
	// If this doesn't work attempt to acquire a new token.
	token, err := ioutil.ReadFile("token-cache")
	if err != nil {
		username, password := readCredentials()
		token, err = requestToken(tokenServerURL, username, password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error requesting token: %s\n", err)
			os.Exit(1)
		}
	}

	// Write cached token to be used for next time.
	if cacheTokens == "true" {
		err := ioutil.WriteFile("token-cache", []byte(token), os.FileMode(0600))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to cache token %s\n")
		}
	}

	execCredential := ExecCredential{
		APIVersion: "client.authentication.k8s.io/v1alpha1",
	 	Kind: "ExecCredential",
	 	Status: map[string]string{
			"token": string(token),
	 	},
	}
	
	output, err := json.Marshal(execCredential)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshalling json: %s\n")
		os.Exit(1)
	}

	// fmt.Fprintf(os.Stderr, "%s\n", output)
	// fmt.Fprintf(os.Stderr, "%s\n", oldoutput)
	fmt.Printf("%s", output)
	// fmt.Printf("%s", oldoutput)
}

// Plugin must write prompts to stderr.
func readCredentials() (username, password string) {
	fmt.Fprintf(os.Stderr, "Please enter username: \n")
	fmt.Fscanf(os.Stdin, "%s", &username)

	fmt.Fprintf(os.Stderr, "Please enter password: \n")
	fmt.Fscanf(os.Stdin, "%s", &password)

	return strings.TrimSpace(username), strings.TrimSpace(password)
}

func requestToken(server, username, password string) (token []byte, err error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest("GET", server, nil)
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	return bodyText, nil
}
