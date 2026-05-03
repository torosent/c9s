package cli

import (
	"context"
	"fmt"
	"sync"
)

// Fake is a stub implementation of Client used by tests. It records
// every method called on it for later assertions. All methods are
// safe to call from multiple goroutines.
type Fake struct {
	VersionResp string
	VersionErr  error
	CapsResp    Capabilities
	CapsErr     error

	// Plan 2: container lifecycle responses
	ListContainersResp   []Container
	ListContainersErr    error
	InspectContainerResp []byte
	InspectContainerErr  error
	StopContainerErr     error
	KillContainerErr     error
	RestartContainerErr  error
	DeleteContainerErr   error
	PauseContainerErr    error
	UnpauseContainerErr  error
	PruneContainersResp  int
	PruneContainersErr   error

	// Plan 3: streaming responses
	logStreamEvents   []StreamEvent
	logStreamExitCode int

	buildStreamEvents   []StreamEvent
	buildStreamExitCode int

	layerStreamEvents   []StreamEvent
	layerStreamExitCode int

	// Plan 4: image responses
	ListImagesResp   []Image
	ListImagesErr    error
	InspectImageResp []byte
	InspectImageErr  error
	TagImageErr      error
	DeleteImageErr   error
	PruneImagesResp  int
	PruneImagesErr   error
	LoadImageErr     error
	SaveImageErr     error

	// Plan 4: volume responses
	ListVolumesResp  []Volume
	ListVolumesErr   error
	CreateVolumeErr  error
	DeleteVolumeErr  error
	PruneVolumesResp int
	PruneVolumesErr  error

	// Plan 4: network responses
	ListNetworksResp  []Network
	ListNetworksErr   error
	CreateNetworkErr  error
	DeleteNetworkErr  error
	PruneNetworksResp int
	PruneNetworksErr  error

	// Plan 4: builder responses
	BuilderStatusResp BuilderStatus
	BuilderStatusErr  error
	BuilderStartErr   error
	BuilderStopErr    error
	BuilderDeleteErr  error

	// Plan 4: registry responses
	ListRegistriesResp    []RegistryEntry
	ListRegistriesErr     error
	RegistryLoginErr      error
	RegistryLogoutErr     error
	RegistrySetDefaultErr error
	RegistryLoginLastUser string
	RegistryLoginLastPass string
	RegistryLoginLastHost string

	// Plan 4: system responses
	ListSystemServicesResp   []SystemService
	ListSystemServicesErr    error
	SystemStartAllErr        error
	SystemStopAllErr         error
	SystemDFResp             SystemDF
	SystemDFErr              error
	ListDNSDomainsResp       []DNSDomain
	ListDNSDomainsErr        error
	CreateDNSDomainErr       error
	DeleteDNSDomainErr       error
	SetDefaultDNSDomainErr   error
	ListSystemPropertiesResp []SystemProperty
	ListSystemPropertiesErr  error
	SetSystemPropertyErr     error
	ResetSystemPropertyErr   error

	// Plan 4: system log stream
	systemLogStreamEvents   []StreamEvent
	systemLogStreamExitCode int

	// Plan 4: run stream
	runStreamEvents   []StreamEvent
	runStreamExitCode int

	mu    sync.Mutex
	Calls []string
}

// NewFake returns a new Fake client for testing.
func NewFake() *Fake {
	return &Fake{}
}

// Version implements Client.
func (f *Fake) Version(_ context.Context) (string, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "Version")
	f.mu.Unlock()
	return f.VersionResp, f.VersionErr
}

// Capabilities implements Client.
func (f *Fake) Capabilities(_ context.Context) (Capabilities, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "Capabilities")
	f.mu.Unlock()
	return f.CapsResp, f.CapsErr
}

// ListContainers implements Client.
func (f *Fake) ListContainers(_ context.Context, _ bool) ([]Container, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListContainers")
	f.mu.Unlock()
	return f.ListContainersResp, f.ListContainersErr
}

// InspectContainer implements Client.
func (f *Fake) InspectContainer(_ context.Context, _ string) ([]byte, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "InspectContainer")
	f.mu.Unlock()
	return f.InspectContainerResp, f.InspectContainerErr
}

// StopContainer implements Client.
func (f *Fake) StopContainer(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "StopContainer")
	f.mu.Unlock()
	return f.StopContainerErr
}

// KillContainer implements Client.
func (f *Fake) KillContainer(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "KillContainer")
	f.mu.Unlock()
	return f.KillContainerErr
}

// RestartContainer implements Client.
func (f *Fake) RestartContainer(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "RestartContainer")
	f.mu.Unlock()
	return f.RestartContainerErr
}

// DeleteContainer implements Client.
func (f *Fake) DeleteContainer(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "DeleteContainer")
	f.mu.Unlock()
	return f.DeleteContainerErr
}

// PauseContainer implements Client.
func (f *Fake) PauseContainer(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "PauseContainer")
	f.mu.Unlock()
	return f.PauseContainerErr
}

// UnpauseContainer implements Client.
func (f *Fake) UnpauseContainer(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "UnpauseContainer")
	f.mu.Unlock()
	return f.UnpauseContainerErr
}

// PruneContainers implements Client.
func (f *Fake) PruneContainers(_ context.Context) (int, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "PruneContainers")
	f.mu.Unlock()
	return f.PruneContainersResp, f.PruneContainersErr
}

// ListImages implements Client.
func (f *Fake) ListImages(_ context.Context) ([]Image, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListImages")
	f.mu.Unlock()
	return f.ListImagesResp, f.ListImagesErr
}

// InspectImage implements Client.
func (f *Fake) InspectImage(_ context.Context, _ string) ([]byte, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "InspectImage")
	f.mu.Unlock()
	return f.InspectImageResp, f.InspectImageErr
}

// TagImage implements Client.
func (f *Fake) TagImage(_ context.Context, _, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "TagImage")
	f.mu.Unlock()
	return f.TagImageErr
}

// DeleteImage implements Client.
func (f *Fake) DeleteImage(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "DeleteImage")
	f.mu.Unlock()
	return f.DeleteImageErr
}

// PruneImages implements Client.
func (f *Fake) PruneImages(_ context.Context, _ bool) (int, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "PruneImages")
	f.mu.Unlock()
	return f.PruneImagesResp, f.PruneImagesErr
}

// LoadImage implements Client.
func (f *Fake) LoadImage(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "LoadImage")
	f.mu.Unlock()
	return f.LoadImageErr
}

// SaveImage implements Client.
func (f *Fake) SaveImage(_ context.Context, _, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "SaveImage")
	f.mu.Unlock()
	return f.SaveImageErr
}

// ListVolumes implements Client.
func (f *Fake) ListVolumes(_ context.Context) ([]Volume, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListVolumes")
	f.mu.Unlock()
	return f.ListVolumesResp, f.ListVolumesErr
}

// CreateVolume implements Client.
func (f *Fake) CreateVolume(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "CreateVolume")
	f.mu.Unlock()
	return f.CreateVolumeErr
}

// DeleteVolume implements Client.
func (f *Fake) DeleteVolume(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "DeleteVolume")
	f.mu.Unlock()
	return f.DeleteVolumeErr
}

// PruneVolumes implements Client.
func (f *Fake) PruneVolumes(_ context.Context) (int, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "PruneVolumes")
	f.mu.Unlock()
	return f.PruneVolumesResp, f.PruneVolumesErr
}

// ListNetworks implements Client.
func (f *Fake) ListNetworks(_ context.Context) ([]Network, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListNetworks")
	f.mu.Unlock()
	return f.ListNetworksResp, f.ListNetworksErr
}

// CreateNetwork implements Client.
func (f *Fake) CreateNetwork(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "CreateNetwork")
	f.mu.Unlock()
	return f.CreateNetworkErr
}

// DeleteNetwork implements Client.
func (f *Fake) DeleteNetwork(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "DeleteNetwork")
	f.mu.Unlock()
	return f.DeleteNetworkErr
}

// PruneNetworks implements Client.
func (f *Fake) PruneNetworks(_ context.Context) (int, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "PruneNetworks")
	f.mu.Unlock()
	return f.PruneNetworksResp, f.PruneNetworksErr
}

// BuilderStatus implements Client.
func (f *Fake) BuilderStatus(_ context.Context) (BuilderStatus, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "BuilderStatus")
	f.mu.Unlock()
	return f.BuilderStatusResp, f.BuilderStatusErr
}

// BuilderStart implements Client.
func (f *Fake) BuilderStart(_ context.Context) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "BuilderStart")
	f.mu.Unlock()
	return f.BuilderStartErr
}

// BuilderStop implements Client.
func (f *Fake) BuilderStop(_ context.Context) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "BuilderStop")
	f.mu.Unlock()
	return f.BuilderStopErr
}

// BuilderDelete implements Client.
func (f *Fake) BuilderDelete(_ context.Context) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "BuilderDelete")
	f.mu.Unlock()
	return f.BuilderDeleteErr
}

// ListRegistries implements Client.
func (f *Fake) ListRegistries(_ context.Context) ([]RegistryEntry, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListRegistries")
	f.mu.Unlock()
	return f.ListRegistriesResp, f.ListRegistriesErr
}

// RegistryLogin implements Client.
func (f *Fake) RegistryLogin(_ context.Context, host, user, pass string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "RegistryLogin")
	f.RegistryLoginLastHost = host
	f.RegistryLoginLastUser = user
	f.RegistryLoginLastPass = pass
	f.mu.Unlock()
	return f.RegistryLoginErr
}

// RegistryLogout implements Client.
func (f *Fake) RegistryLogout(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "RegistryLogout")
	f.mu.Unlock()
	return f.RegistryLogoutErr
}

// RegistrySetDefault implements Client.
func (f *Fake) RegistrySetDefault(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "RegistrySetDefault")
	f.mu.Unlock()
	return f.RegistrySetDefaultErr
}

// ListSystemServices implements Client.
func (f *Fake) ListSystemServices(_ context.Context) ([]SystemService, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListSystemServices")
	f.mu.Unlock()
	return f.ListSystemServicesResp, f.ListSystemServicesErr
}

// SystemStartAll implements Client.
func (f *Fake) SystemStartAll(_ context.Context) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "SystemStartAll")
	f.mu.Unlock()
	return f.SystemStartAllErr
}

// SystemStopAll implements Client.
func (f *Fake) SystemStopAll(_ context.Context) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "SystemStopAll")
	f.mu.Unlock()
	return f.SystemStopAllErr
}

// SystemDF implements Client.
func (f *Fake) SystemDF(_ context.Context) (SystemDF, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "SystemDF")
	f.mu.Unlock()
	return f.SystemDFResp, f.SystemDFErr
}

// ListDNSDomains implements Client.
func (f *Fake) ListDNSDomains(_ context.Context) ([]DNSDomain, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListDNSDomains")
	f.mu.Unlock()
	return f.ListDNSDomainsResp, f.ListDNSDomainsErr
}

// CreateDNSDomain implements Client.
func (f *Fake) CreateDNSDomain(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "CreateDNSDomain")
	f.mu.Unlock()
	return f.CreateDNSDomainErr
}

// DeleteDNSDomain implements Client.
func (f *Fake) DeleteDNSDomain(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "DeleteDNSDomain")
	f.mu.Unlock()
	return f.DeleteDNSDomainErr
}

// SetDefaultDNSDomain implements Client.
func (f *Fake) SetDefaultDNSDomain(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "SetDefaultDNSDomain")
	f.mu.Unlock()
	return f.SetDefaultDNSDomainErr
}

// ListSystemProperties implements Client.
func (f *Fake) ListSystemProperties(_ context.Context) ([]SystemProperty, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ListSystemProperties")
	f.mu.Unlock()
	return f.ListSystemPropertiesResp, f.ListSystemPropertiesErr
}

// SetSystemProperty implements Client.
func (f *Fake) SetSystemProperty(_ context.Context, _, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "SetSystemProperty")
	f.mu.Unlock()
	return f.SetSystemPropertyErr
}

// ResetSystemProperty implements Client.
func (f *Fake) ResetSystemProperty(_ context.Context, _ string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, "ResetSystemProperty")
	f.mu.Unlock()
	return f.ResetSystemPropertyErr
}

// StreamSystemLogs implements Client.
func (f *Fake) StreamSystemLogs(_ context.Context, follow bool) (Stream, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, fmt.Sprintf("StreamSystemLogs(follow=%t)", follow))
	events := f.systemLogStreamEvents
	exitCode := f.systemLogStreamExitCode
	f.mu.Unlock()

	eventsCh := make(chan StreamEvent, len(events)+1)
	doneCh := make(chan StreamResult, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(eventsCh)
		defer func() {
			doneCh <- StreamResult{ExitCode: exitCode}
			close(doneCh)
		}()
		for _, ev := range events {
			select {
			case eventsCh <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()
	return Stream{Events: eventsCh, Done: doneCh, Cancel: cancel}, nil
}

// ReplaySystemLogStream configures the Fake's StreamSystemLogs payload.
func (f *Fake) ReplaySystemLogStream(events []StreamEvent, exitCode int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.systemLogStreamEvents = events
	f.systemLogStreamExitCode = exitCode
}

// RunContainer implements Client.
func (f *Fake) RunContainer(_ context.Context, opts RunOpts) (Stream, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, fmt.Sprintf("RunContainer(image=%s)", opts.Image))
	events := f.runStreamEvents
	exitCode := f.runStreamExitCode
	f.mu.Unlock()

	eventsCh := make(chan StreamEvent, len(events)+1)
	doneCh := make(chan StreamResult, 1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(eventsCh)
		defer func() {
			doneCh <- StreamResult{ExitCode: exitCode}
			close(doneCh)
		}()
		for _, ev := range events {
			select {
			case eventsCh <- ev:
			case <-ctx.Done():
				return
			}
		}
	}()
	return Stream{Events: eventsCh, Done: doneCh, Cancel: cancel}, nil
}

// ReplayRunStream configures the Fake's RunContainer payload.
func (f *Fake) ReplayRunStream(events []StreamEvent, exitCode int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.runStreamEvents = events
	f.runStreamExitCode = exitCode
}
