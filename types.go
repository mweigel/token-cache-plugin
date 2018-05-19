package main

type ExecCredential struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Status     map[string]string `json:"status"`
}

/* 
type TokenReview struct {
	APIVersion string
	Kind       string
	Status     struct {
		Authenticated bool
		User struct {
			Username string
			UID	string
			Groups []string
			Extra map[string][]string
		}
	}
} */

type Config struct {
	tokenServerURL      string
	cacheTokens         bool
	caCert              string
	skipTLSVerification bool
}
