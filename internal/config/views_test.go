package config

import "testing"

func TestApplyViewConfig_NoConfig(t *testing.T) {
	columns := []string{"id", "name", "status"}
	viewCfg := ViewConfig{}

	result := ApplyViewConfig(columns, viewCfg)

	// Should return all columns unchanged when no config
	if len(result) != len(columns) {
		t.Errorf("expected %d columns, got %d", len(columns), len(result))
	}
	for i, col := range columns {
		if result[i] != col {
			t.Errorf("expected column %d to be %q, got %q", i, col, result[i])
		}
	}
}

func TestApplyViewConfig_HideColumn(t *testing.T) {
	columns := []string{"id", "name", "status"}
	viewCfg := ViewConfig{
		Columns: map[string]ColumnConfig{
			"status": {Visible: false},
		},
	}

	result := ApplyViewConfig(columns, viewCfg)

	// status should be filtered out
	expected := []string{"id", "name"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(result))
	}
	for i, want := range expected {
		if result[i] != want {
			t.Errorf("expected column %d to be %q, got %q", i, want, result[i])
		}
	}
}

func TestApplyViewConfig_ReorderColumns(t *testing.T) {
	columns := []string{"id", "name", "status"}
	viewCfg := ViewConfig{
		Columns: map[string]ColumnConfig{
			"id":     {Visible: true, Order: 2},
			"name":   {Visible: true, Order: 0},
			"status": {Visible: true, Order: 1},
		},
	}

	result := ApplyViewConfig(columns, viewCfg)

	// Should be reordered: name (0), status (1), id (2)
	expected := []string{"name", "status", "id"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(result))
	}
	for i, want := range expected {
		if result[i] != want {
			t.Errorf("expected column %d to be %q, got %q", i, want, result[i])
		}
	}
}

func TestApplyViewConfig_PartialReorder(t *testing.T) {
	columns := []string{"id", "name", "status", "image"}
	viewCfg := ViewConfig{
		Columns: map[string]ColumnConfig{
			"image": {Visible: true, Order: 0},
			// Other columns have no explicit order, should maintain relative position
		},
	}

	result := ApplyViewConfig(columns, viewCfg)

	// image should be first, others maintain order
	if result[0] != "image" {
		t.Errorf("expected first column to be 'image', got %q", result[0])
	}
	// Remaining columns should be id, name, status in order
	remaining := result[1:]
	expected := []string{"id", "name", "status"}
	for i, want := range expected {
		if remaining[i] != want {
			t.Errorf("expected column %d to be %q, got %q", i+1, want, remaining[i])
		}
	}
}

func TestApplyViewConfig_HideAndReorder(t *testing.T) {
	columns := []string{"id", "name", "status", "image"}
	viewCfg := ViewConfig{
		Columns: map[string]ColumnConfig{
			"image":  {Visible: true, Order: 0},
			"status": {Visible: false},
			"id":     {Visible: true, Order: 2},
			"name":   {Visible: true, Order: 1},
		},
	}

	result := ApplyViewConfig(columns, viewCfg)

	// Should be: image (0), name (1), id (2); status hidden
	expected := []string{"image", "name", "id"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(result))
	}
	for i, want := range expected {
		if result[i] != want {
			t.Errorf("expected column %d to be %q, got %q", i, want, result[i])
		}
	}
}

func TestGetColumnWidth_Default(t *testing.T) {
	viewCfg := ViewConfig{}

	// When no config, should return 0 (use default)
	width := GetColumnWidth("name", viewCfg)
	if width != 0 {
		t.Errorf("expected width 0 for unconfigured column, got %d", width)
	}
}

func TestGetColumnWidth_Configured(t *testing.T) {
	viewCfg := ViewConfig{
		Columns: map[string]ColumnConfig{
			"name": {Visible: true, Width: 30},
		},
	}

	width := GetColumnWidth("name", viewCfg)
	if width != 30 {
		t.Errorf("expected width 30, got %d", width)
	}
}

func TestGetColumnWidth_NotConfigured(t *testing.T) {
	viewCfg := ViewConfig{
		Columns: map[string]ColumnConfig{
			"name": {Visible: true, Width: 30},
		},
	}

	width := GetColumnWidth("other", viewCfg)
	if width != 0 {
		t.Errorf("expected width 0 for unconfigured column, got %d", width)
	}
}
