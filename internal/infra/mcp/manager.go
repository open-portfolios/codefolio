package mcp

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/open-portfolios/codefolio/internal/conf"
	"github.com/open-portfolios/codefolio/internal/domain"
)

const defaultTimeout = 30 * time.Second

type Manager struct {
	configs []conf.MCPServer
	factory ClientFactory

	mu       sync.Mutex
	clients  map[string]Client
	loaded   map[string]bool
	failed   map[string]string
	toolSize map[string]int
	closed   bool
}

type Summary struct {
	Configured  int
	Ready       int
	Unavailable int
	Tools       int
}

func NewManager(cfg *conf.Struct) (*Manager, func(), error) {
	manager := NewManagerWithFactory(cfg.MCPServers, NewSDKClientFactory())
	return manager, func() { _ = manager.Close() }, nil
}

func NewManagerWithFactory(configs []conf.MCPServer, factory ClientFactory) *Manager {
	enabled := make([]conf.MCPServer, 0, len(configs))
	for _, config := range configs {
		if config.IsEnabled() {
			enabled = append(enabled, config)
		}
	}
	sort.Slice(enabled, func(i, j int) bool { return enabled[i].Name < enabled[j].Name })
	return &Manager{
		configs:  enabled,
		factory:  factory,
		clients:  make(map[string]Client),
		loaded:   make(map[string]bool),
		failed:   make(map[string]string),
		toolSize: make(map[string]int),
	}
}

// Discover starts each configured server at most once, catalogs its tools, and
// returns only tools first discovered by this call.
func (m *Manager) Discover(ctx context.Context) ([]domain.Tool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, errors.New("MCP manager is closed")
	}

	tools := make([]domain.Tool, 0)
	errs := make([]error, 0)
	catalogued := false
	for _, config := range m.configs {
		if m.loaded[config.Name] {
			continue
		}
		serverCtx, cancel := context.WithTimeout(ctx, timeoutFor(config))
		client, err := m.factory.Start(serverCtx, config)
		cancel()
		if err != nil {
			wrapped := fmt.Errorf("MCP server %q: %w", config.Name, err)
			m.failed[config.Name] = wrapped.Error()
			errs = append(errs, wrapped)
			continue
		}
		m.clients[config.Name] = client

		serverCtx, cancel = context.WithTimeout(ctx, timeoutFor(config))
		remoteTools, err := client.ListTools(serverCtx)
		cancel()
		if err != nil {
			_ = client.Close()
			delete(m.clients, config.Name)
			wrapped := fmt.Errorf("MCP server %q list tools: %w", config.Name, err)
			m.failed[config.Name] = wrapped.Error()
			errs = append(errs, wrapped)
			continue
		}
		m.loaded[config.Name] = true
		delete(m.failed, config.Name)
		m.toolSize[config.Name] = len(remoteTools)
		catalogued = true
		for _, remote := range remoteTools {
			tools = append(tools, NewTool(config.Name, remote, client))
		}
	}
	if len(errs) > 0 && !catalogued {
		return nil, errors.Join(errs...)
	}
	return tools, nil
}

func (m *Manager) Summary() Summary {
	m.mu.Lock()
	defer m.mu.Unlock()
	summary := Summary{Configured: len(m.configs)}
	for _, config := range m.configs {
		if m.loaded[config.Name] {
			summary.Ready++
			summary.Tools += m.toolSize[config.Name]
			continue
		}
		if m.failed[config.Name] != "" {
			summary.Unavailable++
		}
	}
	return summary
}

func timeoutFor(config conf.MCPServer) time.Duration {
	if config.TimeoutMS > 0 {
		return time.Duration(config.TimeoutMS) * time.Millisecond
	}
	return defaultTimeout
}

func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	var errs []error
	for name, client := range m.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close MCP server %q: %w", name, err))
		}
	}
	m.clients = make(map[string]Client)
	return errors.Join(errs...)
}
