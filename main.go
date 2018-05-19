package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

var cfg = config{}

func init() {
	flag.StringVar(&cfg.tokenRequestEndpoint, "token-request-endpoint", "", "help message for flagname")
	flag.StringVar(&cfg.tokenReviewEndpoint, "token-review-endpoint", "", "help message for flagname")
	flag.StringVar(&cfg.caCert, "ca-cert", "", "help message for flagname")
	flag.StringVar(&cfg.tokenPath, "token-path", "", "help message for flagname")
	flag.BoolVar(&cfg.skipTLSVerification, "skip-tls-verification", false, "help message for flagname")
	flag.BoolVar(&cfg.cacheTokens, "cache-tokens", true, "help message for flagname")

	if cfg.tokenPath == "" {
		currentUser, err := user.Current()
		if err != nil {
			os.Exit(1)
		}
		cfg.tokenPath = filepath.Join(currentUser.HomeDir, ".k8s-last-token")
	}
}

func main() {
	flag.Parse()

	// Log messages must be written to stderr as kubectl is expecting execCredential on stdout.
	logger := log.New(os.Stderr, "", 0)

	client, err := getHTTPClient()
	if err != nil {
		logger.Fatalf("Error creating HTTP client: %s\n", err)
	}

	// Attempt to read and use a previously cached token before prompting for a username and password.
	token, err := ioutil.ReadFile(cfg.tokenPath)
	if err != nil {
		logger.Println(err)
	}

	valid, err := reviewToken(client, token)
	if err != nil {
		logger.Println(err)
	}
	if !valid {
		var username, password string
		if err = readCredentials(&username, &password); err != nil {
			logger.Fatalf("Error reading credentials: %s\n", err)
		}
		if token, err = requestToken(client, username, password); err != nil {
			logger.Fatalf("Error requesting token: %s\n", err)
		}

		// Write token to file to be used next time kubectl is run unless caching is disabled.
		if cfg.tokenPath != "" {
			if err = ioutil.WriteFile(cfg.tokenPath, token, os.FileMode(0600)); err != nil {
				logger.Println(err)
			}
		}
	}

	// Write token to stdout to be used by kubectl.
	if err = outputToken(token); err != nil {
		logger.Fatalf("Unable to output token: %s\n", err)
	}
}

// Review token using the same endpoint that K8s will also use.
// https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication
func reviewToken(client *http.Client, token []byte) (bool, error) {
	tokenReviewRequest := &tokenReviewRequest{
		APIVersion: "client.authentication.k8s.io/v1beta",
		Kind:       "TokenReview",
		Spec: map[string]string{
			"token": string(token),
		},
	}
	output, err := json.Marshal(tokenReviewRequest)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", cfg.tokenReviewEndpoint, bytes.NewReader(output))
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	tokenResponse := tokenReviewResponse{}
	err = json.Unmarshal(data, &tokenResponse)
	if err != nil {
		return false, err
	}

	if !tokenResponse.Status.Authenticated {
		return false, err
	}

	return true, nil
}

// Request a token from token service.
func requestToken(client *http.Client, username, password string) (token []byte, err error) {
	req, err := http.NewRequest("GET", cfg.tokenRequestEndpoint, nil)
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return token, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// Return token to kubectl on stdout.
// https://kubernetes.io/docs/admin/authentication/#input-and-output-formats
func outputToken(token []byte) error {
	execCredential := execCredential{
		APIVersion: "client.authentication.k8s.io/v1alpha1",
		Kind:       "ExecCredential",
		Status: map[string]string{
			"token": string(token),
		},
	}

	output, err := json.Marshal(execCredential)
	if err != nil {
		return err
	}

	fmt.Printf("%s", output)
	return err
}

func readCredentials(username, password *string) error {
	fmt.Fprintf(os.Stderr, "username: ")
	fmt.Fscanf(os.Stdin, "%s", username)

	fmt.Fprintf(os.Stderr, "password: ")
	p, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	*password = string(p)

	return nil
}

func getHTTPClient() (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.skipTLSVerification,
	}

	if cfg.caCert != "" {
		caCert, err := ioutil.ReadFile(cfg.caCert)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}
