package pulses

import (
	"strings"
	"testing"
	"time"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNew(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{{ID: "c1", ShortID: "c1"}},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	if m == nil {
		t.Fatal("expected non-nil model")
	}

	if m.Title() != "pulses" {
		t.Errorf("expected title 'pulses', got %s", m.Title())
	}
}

func TestView(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{{ID: "c1", ShortID: "c1"}, {ID: "c2", ShortID: "c2"}},
		ListImagesResp:     []cli.Image{{ID: "i1", ShortID: "i1"}},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Update(DataRefreshedMsg{Containers: 2, Images: 1, Volumes: 0, Networks: 0})

	view := m.View(80, 24)
	if !strings.Contains(view, "Containers: 2") {
		t.Errorf("expected view to show 2 containers, got: %s", view)
	}
	if !strings.Contains(view, "Images:     1") {
		t.Errorf("expected view to show 1 image, got: %s", view)
	}
}

func TestRefreshTick(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{{ID: "c1", ShortID: "c1"}},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)

	// Send RefreshMsg
	_, cmd := m.Update(RefreshMsg{})
	if cmd == nil {
		t.Fatal("expected refresh command")
	}

	// Execute the cmd to get DataRefreshedMsg
	msg := cmd()
	dataMsg, ok := msg.(DataRefreshedMsg)
	if !ok {
		t.Fatalf("expected DataRefreshedMsg, got %T", msg)
	}

	if dataMsg.Containers != 1 {
		t.Errorf("expected 1 container, got %d", dataMsg.Containers)
	}
}

func TestSummary(t *testing.T) {
	fake := &cli.Fake{}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Update(DataRefreshedMsg{Containers: 5, Images: 10})

	summary := m.Summary()
	if !strings.Contains(summary, "5 containers") {
		t.Errorf("expected summary to contain '5 containers', got %s", summary)
	}
	if !strings.Contains(summary, "10 images") {
		t.Errorf("expected summary to contain '10 images', got %s", summary)
	}
}
