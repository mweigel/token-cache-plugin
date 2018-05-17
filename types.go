package main

type ExecCredential struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Status     map[string]string `json:"status"`
}

type Config struct {
	tokenServerURL      string
	cacheTokens         bool
	caCert              string
	skipTLSVerification bool
}
