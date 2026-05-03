package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// k9sStatusBlock holds the k9s status colors.
type k9sStatusBlock struct {
	NewColor       string `yaml:"newColor"`
	ModifyColor    string `yaml:"modifyColor"`
	AddColor       string `yaml:"addColor"`
	ErrorColor     string `yaml:"errorColor"`
	HighlightColor string `yaml:"highlightColor"`
	KillColor      string `yaml:"killColor"`
	CompletedColor string `yaml:"completedColor"`
	PendingColor   string `yaml:"pendingColor"`
}

// k9sFrameBlock holds the k9s frame block.
type k9sFrameBlock struct {
	Border struct {
		FgColor    string `yaml:"fgColor"`
		FocusColor string `yaml:"focusColor"`
	} `yaml:"border"`
	Title struct {
		FgColor          string `yaml:"fgColor"`
		BgColor          string `yaml:"bgColor"`
		HighlightColor   string `yaml:"highlightColor"`
		CounterColor     string `yaml:"counterColor"`
		FilterColor      string `yaml:"filterColor"`
		FocusBorderColor string `yaml:"focusBorderColor"`
	} `yaml:"title"`
	Crumbs struct {
		FgColor     string `yaml:"fgColor"`
		BgColor     string `yaml:"bgColor"`
		ActiveColor string `yaml:"activeColor"`
	} `yaml:"crumbs"`
	Status k9sStatusBlock `yaml:"status"`
	Menu   struct {
		FgColor     string `yaml:"fgColor"`
		KeyColor    string `yaml:"keyColor"`
		NumKeyColor string `yaml:"numKeyColor"`
	} `yaml:"menu"`
}

// k9sBodyBlock holds the k9s body styles.
type k9sBodyBlock struct {
	FgColor   string `yaml:"fgColor"`
	BgColor   string `yaml:"bgColor"`
	LogoColor string `yaml:"logoColor"`
}

// k9sInfoBlock holds the k9s cluster info pane.
type k9sInfoBlock struct {
	FgColor      string `yaml:"fgColor"`
	BgColor      string `yaml:"bgColor"`
	SectionColor string `yaml:"sectionColor"`
}

// k9sViewsBlock holds tabular view styling.
type k9sViewsBlock struct {
	Table struct {
		FgColor string `yaml:"fgColor"`
		BgColor string `yaml:"bgColor"`
		Header  struct {
			FgColor     string `yaml:"fgColor"`
			BgColor     string `yaml:"bgColor"`
			SorterColor string `yaml:"sorterColor"`
		} `yaml:"header"`
		CursorFgColor string `yaml:"cursorFgColor"`
		CursorBgColor string `yaml:"cursorBgColor"`
	} `yaml:"table"`
	Yaml struct {
		KeyColor   string `yaml:"keyColor"`
		ColonColor string `yaml:"colonColor"`
		ValueColor string `yaml:"valueColor"`
	} `yaml:"yaml"`
}

// K9sSkin represents the real k9s YAML skin structure (nested under k9s.body, k9s.frame, etc.).
type K9sSkin struct {
	K9s struct {
		Body  k9sBodyBlock  `yaml:"body"`
		Info  k9sInfoBlock  `yaml:"info"`
		Frame k9sFrameBlock `yaml:"frame"`
		Views k9sViewsBlock `yaml:"views"`
	} `yaml:"k9s"`
}

// firstNonEmpty returns the first non-empty string from the candidates.
func firstNonEmpty(opts ...string) string {
	for _, s := range opts {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

// FromK9sYAML parses a k9s YAML byte slice into a c9s Skin.
func FromK9sYAML(data []byte) (Skin, error) {
	var k9s K9sSkin
	if err := yaml.Unmarshal(data, &k9s); err != nil {
		return Skin{}, fmt.Errorf("parse k9s skin: %w", err)
	}

	skin := Skin{}
	skin.Colors.Fg = firstNonEmpty(k9s.K9s.Body.FgColor, k9s.K9s.Views.Table.FgColor)
	skin.Colors.Bg = firstNonEmpty(k9s.K9s.Body.BgColor, k9s.K9s.Views.Table.BgColor)
	skin.Colors.Border = firstNonEmpty(k9s.K9s.Frame.Border.FocusColor, k9s.K9s.Frame.Border.FgColor)
	skin.Colors.Accent = firstNonEmpty(k9s.K9s.Body.LogoColor, k9s.K9s.Frame.Title.HighlightColor, k9s.K9s.Frame.Menu.KeyColor)
	// Dim is for muted-but-readable text. K9s's `crumbs.fgColor` is often the
	// darkest text color (close to bg) — avoid it. Prefer info.fgColor or
	// menu.fgColor which are designed to be readable subtitles.
	skin.Colors.Dim = firstNonEmpty(k9s.K9s.Info.FgColor, k9s.K9s.Frame.Menu.FgColor, k9s.K9s.Info.SectionColor)
	skin.Colors.Success = firstNonEmpty(k9s.K9s.Frame.Status.NewColor, k9s.K9s.Frame.Status.AddColor)
	skin.Colors.Warning = firstNonEmpty(k9s.K9s.Frame.Status.ModifyColor, k9s.K9s.Frame.Status.PendingColor)
	skin.Colors.Error = firstNonEmpty(k9s.K9s.Frame.Status.ErrorColor, k9s.K9s.Frame.Status.KillColor)
	// Selection must be highly visible. K9s skins often set cursor near-bg
	// (nightfox: cursor = current_line ≈ bg) which is invisible against the
	// row colour. Use the body fg as the selection bg and body bg as the
	// selection fg — an inverted highlight that is guaranteed contrast.
	skin.Colors.SelectionBg = firstNonEmpty(k9s.K9s.Body.LogoColor, k9s.K9s.Frame.Title.HighlightColor, k9s.K9s.Body.FgColor)
	skin.Colors.SelectionFg = firstNonEmpty(k9s.K9s.Body.BgColor, k9s.K9s.Views.Table.BgColor)
	skin.Colors.HeaderFg = firstNonEmpty(k9s.K9s.Frame.Title.FgColor, k9s.K9s.Body.LogoColor)
	skin.Colors.HeaderBg = firstNonEmpty(k9s.K9s.Frame.Title.BgColor, k9s.K9s.Body.BgColor)

	skin.StateColors = map[string]string{
		"running":  firstNonEmpty(k9s.K9s.Frame.Status.NewColor, k9s.K9s.Frame.Status.AddColor),
		"exited":   firstNonEmpty(k9s.K9s.Frame.Crumbs.FgColor, k9s.K9s.Info.SectionColor),
		"paused":   firstNonEmpty(k9s.K9s.Frame.Status.ModifyColor, k9s.K9s.Frame.Status.PendingColor),
		"stopping": firstNonEmpty(k9s.K9s.Frame.Status.ErrorColor, k9s.K9s.Frame.Status.KillColor),
		"created":  firstNonEmpty(k9s.K9s.Frame.Status.HighlightColor, k9s.K9s.Frame.Title.HighlightColor),
	}

	return skin, nil
}

// ImportK9sSkin converts a k9s YAML skin file to c9s TOML format and saves it.
// Returns the output path.
func ImportK9sSkin(yamlPath string) (string, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return "", fmt.Errorf("read k9s skin: %w", err)
	}
	skin, err := FromK9sYAML(data)
	if err != nil {
		return "", err
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("get home dir: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	skinsDir := filepath.Join(configDir, "c9s", "skins")
	if err := os.MkdirAll(skinsDir, 0o755); err != nil {
		return "", fmt.Errorf("create skins dir: %w", err)
	}

	basename := filepath.Base(yamlPath)
	basename = strings.TrimSuffix(basename, filepath.Ext(basename))
	outPath := filepath.Join(skinsDir, basename+".toml")

	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	fmt.Fprintf(f, "# Imported from k9s skin: %s\n", filepath.Base(yamlPath))
	if err := toml.NewEncoder(f).Encode(&skin); err != nil {
		return "", fmt.Errorf("encode TOML: %w", err)
	}

	return outPath, nil
}
