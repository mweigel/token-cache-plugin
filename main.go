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
	flag.StringVar(&cfg.tokenRequestEndpoint, "token-request-endpoint", "", "URL of endpoint responsible for issuing tokens")
	flag.StringVar(&cfg.tokenReviewEndpoint, "token-review-endpoint", "", "URL of endpoint responsible for reviewing tokens")
	flag.StringVar(&cfg.caCert, "ca-cert", "", "Path to CA certificate used to verify token request and review endpoints")
	flag.StringVar(&cfg.tokenPath, "token-path", "", "Fully qualified path to save and load locally cached tokens")
	flag.BoolVar(&cfg.skipTLSVerification, "skip-tls-verification", false, "Skip TLS verification of token request and review endpoint certificates")
	flag.BoolVar(&cfg.cacheTokens, "cache-tokens", true, "Whether to cache tokens returned by the token request endpoint locally")
}

func main() {
	flag.Parse()

	// Log messages must be written to stderr as kubectl is expecting execCredential on stdout.
	logger := log.New(os.Stderr, "", 0)

	// If a path to a token file is not specified and caching is requested set a default.
	if cfg.tokenPath == "" && cfg.cacheTokens {
		currentUser, err := user.Current()
		if err != nil {
			logger.Fatalf("Error setting token-path: %s\n", err)
		}
		cfg.tokenPath = filepath.Join(currentUser.HomeDir, ".k8s-last-token")
	}

	client, err := getHTTPClient()
	if err != nil {
		logger.Fatalf("Error creating HTTP client: %s\n", err)
	}

	// Attempt to read and use a previously cached token before prompting for a username and password.
	token, err := ioutil.ReadFile(cfg.tokenPath)
	if err != nil {
		logger.Println(err)
	}

	tokenResponse, err := reviewToken(client, token)
	if err != nil {
		logger.Println(err)
	}
	if !tokenResponse.Status.Authenticated {
		var username, password string
		if err = readCredentials(&username, &password); err != nil {
			logger.Fatalf("Error reading credentials: %s\n", err)
		}
		if token, err = requestToken(client, username, password); err != nil {
			logger.Fatalf("Error requesting token: %s\n", err)
		}

		// Write token to file to be used next time kubectl is run unless caching is disabled.
		if cfg.cacheTokens {
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
func reviewToken(client *http.Client, token []byte) (tokenReviewResponse, error) {
	postBody, err := json.Marshal(newTokenReviewRequest(token))
	if err != nil {
		return tokenReviewResponse{}, err
	}

	req, err := http.NewRequest("POST", cfg.tokenReviewEndpoint, bytes.NewReader(postBody))
	if err != nil {
		return tokenReviewResponse{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return tokenReviewResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tokenReviewResponse{}, err
	}

	res := tokenReviewResponse{}
	err = json.Unmarshal(respBody, &res)
	return res, err
}

// Request a token from token service.
func requestToken(client *http.Client, username, password string) ([]byte, error) {
	req, err := http.NewRequest("GET", cfg.tokenRequestEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Return token to kubectl on stdout.
// https://kubernetes.io/docs/admin/authentication/#input-and-output-formats
func outputToken(token []byte) error {
	output, err := json.Marshal(newExecCredential(token))
	if err != nil {
		return err
	}

	_, err = fmt.Printf("%s", output)
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
