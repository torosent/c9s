package config

import "sort"

// ApplyViewConfig filters and reorders columns based on ViewConfig.
// Returns a new slice with visible columns in the specified order.
// Columns without explicit config are kept in their original relative order.
func ApplyViewConfig(columns []string, viewCfg ViewConfig) []string {
	if len(viewCfg.Columns) == 0 {
		return columns
	}

	// Build list of visible columns with ordering info
	type orderedCol struct {
		name             string
		order            int
		originalIdx      int
		hasExplicitOrder bool
	}

	var result []orderedCol
	for idx, col := range columns {
		cfg, hasCfg := viewCfg.Columns[col]
		if hasCfg && !cfg.Visible {
			// Column is explicitly hidden
			continue
		}

		// Column is visible (either explicitly or by default)
		// Order 0 could be explicit (first position) or unset
		// We treat any configured column with Order >= 0 as having explicit order
		hasOrder := hasCfg && cfg.Order >= 0
		order := cfg.Order

		result = append(result, orderedCol{
			name:             col,
			order:            order,
			originalIdx:      idx,
			hasExplicitOrder: hasOrder,
		})
	}

	// Sort by order, with fallback to original index
	sort.SliceStable(result, func(i, j int) bool {
		// Both have explicit order: sort by order value
		if result[i].hasExplicitOrder && result[j].hasExplicitOrder {
			return result[i].order < result[j].order
		}
		// One has explicit order: it comes first
		if result[i].hasExplicitOrder != result[j].hasExplicitOrder {
			return result[i].hasExplicitOrder
		}
		// Neither has explicit order: maintain original order
		return result[i].originalIdx < result[j].originalIdx
	})

	// Extract names
	names := make([]string, len(result))
	for i, oc := range result {
		names[i] = oc.name
	}

	return names
}

// GetColumnWidth returns the configured width for a column, or 0 if not configured.
// A width of 0 means "use default width".
func GetColumnWidth(columnName string, viewCfg ViewConfig) int {
	if cfg, ok := viewCfg.Columns[columnName]; ok {
		return cfg.Width
	}
	return 0
}
