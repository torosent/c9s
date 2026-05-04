package acr

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestHostname(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"myreg", "myreg.azurecr.io"},
		{" myreg ", "myreg.azurecr.io"},
		{"myreg.azurecr.io", "myreg.azurecr.io"},
		{"myreg.azurecr.us", "myreg.azurecr.us"},
		{"myreg.azurecr.cn", "myreg.azurecr.cn"},
		{"localhost:5000", "localhost:5000"},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got := Hostname(c.in)
			if got != c.want {
				t.Errorf("Hostname(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestShortName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"myreg", "myreg"},
		{"myreg.azurecr.io", "myreg"},
		{"myreg.azurecr.us", "myreg"},
		{" myreg ", "myreg"},
	}
	for _, c := range cases {
		got := shortName(c.in)
		if got != c.want {
			t.Errorf("shortName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFetchToken_NoAz(t *testing.T) {
	orig := lookPath
	t.Cleanup(func() { lookPath = orig })
	lookPath = func(string) (string, error) { return "", errors.New("not found") }

	_, err := FetchToken(context.Background(), "myreg")
	if !errors.Is(err, ErrAzCLINotFound) {
		t.Fatalf("expected ErrAzCLINotFound, got %v", err)
	}
}

func TestFetchToken_Empty(t *testing.T) {
	_, err := FetchToken(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "registry is empty") {
		t.Fatalf("expected empty-registry error, got %v", err)
	}
}

func TestFetchToken_Success(t *testing.T) {
	origLookPath := lookPath
	origRunAz := runAz
	t.Cleanup(func() { lookPath = origLookPath; runAz = origRunAz })

	lookPath = func(string) (string, error) { return "/usr/local/bin/az", nil }
	runAz = func(ctx context.Context, registry string) ([]byte, []byte, error) {
		if registry != "myreg" {
			t.Errorf("expected --name myreg (short), got %q", registry)
		}
		return []byte("eyJfakeToken_payload\n"), nil, nil
	}

	tok, err := FetchToken(context.Background(), "myreg.azurecr.io")
	if err != nil {
		t.Fatalf("FetchToken: %v", err)
	}
	if tok != "eyJfakeToken_payload" {
		t.Errorf("token = %q, want %q", tok, "eyJfakeToken_payload")
	}
}

func TestFetchToken_AzFails(t *testing.T) {
	origLookPath := lookPath
	origRunAz := runAz
	t.Cleanup(func() { lookPath = origLookPath; runAz = origRunAz })

	lookPath = func(string) (string, error) { return "/usr/local/bin/az", nil }
	runAz = func(ctx context.Context, registry string) ([]byte, []byte, error) {
		return nil, []byte("ERROR: registry 'bogus' not found"), errors.New("exit 1")
	}

	_, err := FetchToken(context.Background(), "bogus")
	if err == nil || !strings.Contains(err.Error(), "registry 'bogus' not found") {
		t.Fatalf("expected wrapped registry-not-found error, got %v", err)
	}
}

func TestFetchToken_EmptyToken(t *testing.T) {
	origLookPath := lookPath
	origRunAz := runAz
	t.Cleanup(func() { lookPath = origLookPath; runAz = origRunAz })

	lookPath = func(string) (string, error) { return "/usr/local/bin/az", nil }
	runAz = func(ctx context.Context, registry string) ([]byte, []byte, error) {
		return []byte("\n"), nil, nil
	}

	_, err := FetchToken(context.Background(), "myreg")
	if err == nil || !strings.Contains(err.Error(), "empty token") {
		t.Fatalf("expected empty-token error, got %v", err)
	}
}
