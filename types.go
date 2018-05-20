package main

// ExecCredenital which will be printed to stdout. k8s.io/client-go will then use the
// returned bearer token in the status when authenticating against the Kubernetes API.
// https://kubernetes.io/docs/admin/authentication/#client-go-credential-plugins
type execCredential struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Status     map[string]string `json:"status"`
}

func newExecCredential(token []byte) *execCredential {
	return &execCredential{
		APIVersion: "client.authentication.k8s.io/v1alpha1",
		Kind:       "ExecCredential",
		Status: map[string]string{
			"token": string(token),
		},
	}
}

// TokenReviewRequest which will be sent when verifying a cached token.
// https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication
type tokenReviewRequest struct {
	APIVersion string
	Kind       string
	Spec       map[string]string
}

func newTokenReviewRequest(token []byte) *tokenReviewRequest {
	return &tokenReviewRequest{
		APIVersion: "client.authentication.k8s.io/v1beta",
		Kind:       "TokenReview",
		Spec: map[string]string{
			"token": string(token),
		},
	}
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

// Config populated by arguments from kubeconfig file.
// https://kubernetes.io/docs/admin/authentication/#configuration
type config struct {
	tokenRequestEndpoint string
	tokenReviewEndpoint  string
	caCert               string
	skipTLSVerification  bool
	cacheTokens          bool
	tokenPath            string
}
