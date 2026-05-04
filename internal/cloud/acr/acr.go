// Package acr wraps the small subset of Azure CLI (`az`) we need to
// short-circuit Azure Container Registry login for c9s users.
//
// The flow we automate is the official Microsoft recipe for ad-hoc
// token-based ACR login:
//
//	# 1. Get a short-lived AAD-backed ACR refresh token (~3h validity).
//	az acr login --name <registry> --expose-token \
//	    --output tsv --query accessToken
//
//	# 2. Hand it to the container runtime as the "anonymous" zero-GUID user.
//	container registry login <registry>.azurecr.io \
//	    --username 00000000-0000-0000-0000-000000000000 \
//	    --password-stdin
//
// See https://learn.microsoft.com/azure/container-registry/container-registry-authentication
// for the security model.
package acr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// AnonymousUser is the username ACR expects when authenticating with an
// AAD bearer token instead of basic auth credentials. ACR treats the
// literal zero GUID as a sentinel for "the password field carries an
// AAD-backed identity token, parse it accordingly".
const AnonymousUser = "00000000-0000-0000-0000-000000000000"

// ErrAzCLINotFound is returned when `az` is not on PATH. The error
// includes a hint pointing at the Microsoft install docs.
var ErrAzCLINotFound = errors.New("Azure CLI (az) not found on PATH; install from https://learn.microsoft.com/cli/azure/install-azure-cli")

// lookPath is a swappable indirection over exec.LookPath so unit tests
// can simulate a missing `az` binary.
//
// NOTE: this is package-level state and not safe for concurrent test
// mutation. c9s only invokes FetchToken from a single tea.Cmd at a
// time, so this is fine in production. If you ever add t.Parallel()
// to acr_test.go, switch to a per-FetchToken option struct first.
var lookPath = exec.LookPath

// runAz is a swappable indirection over the actual `az` invocation so unit
// tests can simulate token retrieval (or failure) without having Azure CLI
// installed and authenticated.
//
// NOTE: same concurrent-mutation caveat as lookPath above.
var runAz = func(ctx context.Context, registry string) ([]byte, []byte, error) {
	//nolint:gosec // 'az' is fixed; registry comes from c9s config / palette input.
	cmd := exec.CommandContext(ctx, "az", "acr", "login",
		"--name", registry,
		"--expose-token",
		"--output", "tsv",
		"--query", "accessToken",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// Hostname returns the canonical ACR hostname for the given input. Both
// "myregistry" and "myregistry.azurecr.io" are accepted; anything that
// already contains a dot or a port (`:`) is treated as a fully-qualified
// hostname and returned as-is. (Other ACR clouds like .azurecr.us /
// .azurecr.cn are preserved by this rule, and a stray non-Azure host
// like localhost:5000 won't get .azurecr.io tacked onto it.)
func Hostname(registry string) string {
	registry = strings.TrimSpace(registry)
	if strings.ContainsAny(registry, ".:") {
		return registry
	}
	return registry + ".azurecr.io"
}

// shortName returns the registry name without the `.azurecr.io` suffix —
// `az acr login --name` wants the bare name, not the FQDN.
func shortName(registry string) string {
	registry = strings.TrimSpace(registry)
	if i := strings.Index(registry, "."); i >= 0 {
		return registry[:i]
	}
	return registry
}

// FetchToken returns a short-lived ACR refresh token for the given
// registry (either the bare name or full hostname). Returns
// ErrAzCLINotFound if `az` isn't on PATH.
func FetchToken(ctx context.Context, registry string) (string, error) {
	registry = strings.TrimSpace(registry)
	if registry == "" {
		return "", errors.New("registry is empty")
	}
	if _, err := lookPath("az"); err != nil {
		return "", ErrAzCLINotFound
	}

	stdout, stderr, err := runAz(ctx, shortName(registry))
	if err != nil {
		msg := strings.TrimSpace(string(stderr))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("az acr login failed: %s", msg)
	}
	token := strings.TrimSpace(string(stdout))
	if token == "" {
		return "", errors.New("az returned empty token (is `az login` current?)")
	}
	return token, nil
}
