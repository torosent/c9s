package cli

// Resource enumerates the top-level resource categories c9s manages.
// Plan 2 will add per-resource structs (Container, Image, etc.).
type Resource string

const (
	ResourceContainers Resource = "containers"
	ResourceImages     Resource = "images"
	ResourceVolumes    Resource = "volumes"
	ResourceNetworks   Resource = "networks"
	ResourceBuilder    Resource = "builder"
	ResourceRegistry   Resource = "registry"
	ResourceSystem     Resource = "system"
)
