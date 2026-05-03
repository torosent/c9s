package config

// ResolveAlias resolves a command through the alias table.
// If the command is found in aliases, returns the aliased value.
// Otherwise, returns the original command unchanged.
// Only resolves the first word (verb) of the command.
func ResolveAlias(cmd string, aliases map[string]string) string {
	if cmd == "" {
		return cmd
	}

	if resolved, ok := aliases[cmd]; ok {
		return resolved
	}

	return cmd
}
