package cli

import (
	"context"
	"testing"
)

func TestRunJSONDecodesArray(t *testing.T) {
	type item struct {
		ID string `json:"id"`
	}

	c := NewDefaultClient(
		WithBinary("echo"),
	)

	result, err := runJSON[[]item](
		context.Background(),
		c,
		"cli.test",
		`[{"id":"a"},{"id":"b"}]`,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[0].ID != "a" {
		t.Errorf("expected first id to be 'a', got '%s'", result[0].ID)
	}
	if result[1].ID != "b" {
		t.Errorf("expected second id to be 'b', got '%s'", result[1].ID)
	}
}

func TestRunJSONNonZeroExit(t *testing.T) {
	type item struct {
		ID string `json:"id"`
	}

	c := NewDefaultClient(
		WithBinary("/bin/sh"),
	)

	_, err := runJSON[[]item](
		context.Background(),
		c,
		"cli.test",
		"-c",
		"echo nope >&2; exit 3",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	cliErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if cliErr.Op != "cli.test" {
		t.Errorf("expected Op='cli.test', got '%s'", cliErr.Op)
	}

	if cliErr.Hint == "" {
		t.Error("expected Hint to contain stderr, got empty string")
	}
}

func TestRunJSONInvalidJSON(t *testing.T) {
	type item struct {
		ID string `json:"id"`
	}

	c := NewDefaultClient(
		WithBinary("echo"),
	)

	_, err := runJSON[[]item](
		context.Background(),
		c,
		"cli.test",
		"not json",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	cliErr, ok := err.(*Error)
	if !ok {
		t.Fatalf("expected *Error, got %T", err)
	}

	if cliErr.Op != "cli.test" {
		t.Errorf("expected Op='cli.test', got '%s'", cliErr.Op)
	}
}
