package system

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

func sampleDNS() []cli.DNSDomain {
	return []cli.DNSDomain{
		{Name: "internal.local", Default: true},
		{Name: "dev.local"},
	}
}

func feedDNS(t *testing.T, m DNSModel, ds []cli.DNSDomain) DNSModel {
	t.Helper()
	snap := state.Snapshot[cli.DNSDomain]{Items: ds, FetchedAt: time.Now()}
	s, _ := m.Update(state.RefreshedMsg[cli.DNSDomain]{
		Resource: dnsResource,
		Snapshot: snap,
	})
	return s.(DNSModel)
}

func TestNewDNS(t *testing.T) {
	m := NewDNS(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "System DNS" {
		t.Errorf("Title=%q", m.Title())
	}
	if m.Init() == nil {
		t.Error("Init nil")
	}
}

func TestDNSRender(t *testing.T) {
	m := NewDNS(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedDNS(t, m, sampleDNS())
	v := m.View(120, 30)
	if !strings.Contains(v, "internal.local") {
		t.Errorf("expected internal.local in view: %q", v)
	}
	if !strings.Contains(m.Summary(), "default internal.local") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestDNSCreateFlow(t *testing.T) {
	f := cli.NewFake()
	m := NewDNS(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
	if cmd == nil {
		t.Fatal("c nil")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected OpenModalMsg from c")
	}

	_, cmd = m.Update(modals.TextInputResultMsg{Result: modals.TextInputResult{Label: "create-dns", Value: "new.local"}})
	if cmd == nil {
		t.Fatal("nil cmd from text input result")
	}
	walk(t, cmd)
	if !contains(f.Calls, "CreateDNSDomain") {
		t.Errorf("expected CreateDNSDomain: %v", f.Calls)
	}
}

func TestDNSCreateEmptyToast(t *testing.T) {
	f := cli.NewFake()
	m := NewDNS(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(modals.TextInputResultMsg{Result: modals.TextInputResult{Label: "create-dns", Value: "  "}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	msg := cmd()
	st, ok := msg.(screens.StatusMsg)
	if !ok {
		t.Fatalf("expected StatusMsg, got %T", msg)
	}
	if !strings.Contains(st.Toast, "empty") {
		t.Errorf("Toast=%q", st.Toast)
	}
	if contains(f.Calls, "CreateDNSDomain") {
		t.Errorf("CreateDNSDomain should not run on empty name")
	}
}

func TestDNSDeleteFlow(t *testing.T) {
	f := cli.NewFake()
	m := NewDNS(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedDNS(t, m, sampleDNS())

	_, cmd := m.Update(tea.KeyPressMsg{Code: 'D', Text: "D"})
	if cmd == nil {
		t.Fatal("D nil")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected confirm modal")
	}
	_, cmd = m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "delete-dns"}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	walk(t, cmd)
	if !contains(f.Calls, "DeleteDNSDomain") {
		t.Errorf("expected DeleteDNSDomain: %v", f.Calls)
	}
}

func TestDNSSetDefault(t *testing.T) {
	f := cli.NewFake()
	m := NewDNS(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedDNS(t, m, sampleDNS())
	_, cmd := m.Update(tea.KeyPressMsg{Code: '*', Text: "*"})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	walk(t, cmd)
	if !contains(f.Calls, "SetDefaultDNSDomain") {
		t.Errorf("expected SetDefaultDNSDomain: %v", f.Calls)
	}
}

func TestDNSRefreshAndFilter(t *testing.T) {
	f := cli.NewFake()
	m := NewDNS(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'r', Text: "r"})
	if cmd == nil {
		t.Fatal("r nil")
	}
	cmd()
	if !contains(f.Calls, "ListDNSDomains") {
		t.Errorf("expected ListDNSDomains: %v", f.Calls)
	}
	m = feedDNS(t, m, sampleDNS())
	s, _ := m.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	m = s.(DNSModel)
	for _, r := range "dev" {
		s, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		m = s.(DNSModel)
	}
	s, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = s.(DNSModel)
	v := m.View(120, 30)
	if !strings.Contains(v, "dev.local") {
		t.Errorf("expected dev.local in filtered view: %q", v)
	}
}

func walk(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		return
	}
	msg := cmd()
	if msg == nil {
		return
	}
	if b, ok := msg.(tea.BatchMsg); ok {
		for _, c := range b {
			walk(t, c)
		}
	}
}
