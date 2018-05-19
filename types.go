package main

// ExecCredenital which will be printed to stdout. k8s.io/client-go will then use the
// returned bearer token in the status when authenticating against the Kubernetes API.
// https://kubernetes.io/docs/admin/authentication/#client-go-credential-plugins
type execCredential struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Status     map[string]string `json:"status"`
}

// TokenReviewRequest which will be sent when verifying a cached token.
// https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication
type tokenReviewRequest struct {
	APIVersion string
	Kind       string
	Spec       map[string]string
}

// TokenReviewResponse which will be received when verifying a cached token.
// https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication
type tokenReviewResponse struct {
	APIVersion string
	Kind       string
	Status     status
}

type status struct {
	Authenticated bool
	User          k8suser
}

type k8suser struct {
	Username string
	UID      string
	Groups   []string
	Extra    map[string][]string
}

// Config populated by environment variables from kubeconfig file.
// https://kubernetes.io/docs/admin/authentication/#configuration
type config struct {
	tokenServerURL      string
	tokenPath           string
	caCert              string
	skipTLSVerification bool
}
