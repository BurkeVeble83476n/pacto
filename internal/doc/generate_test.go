package doc

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/trianalab/pacto/internal/graph"
	"github.com/trianalab/pacto/pkg/contract"
)

func intPtr(v int) *int { return &v }

func fullContract() *contract.Contract {
	return &contract.Contract{
		PactoVersion: "1.0",
		Service: contract.ServiceIdentity{
			Name:    "payments-api",
			Version: "2.1.0",
			Owner:   "team/payments",
			Image:   &contract.Image{Ref: "ghcr.io/acme/payments-api:2.1.0", Private: true},
		},
		Interfaces: []contract.Interface{
			{
				Name:       "rest-api",
				Type:       "http",
				Port:       intPtr(8080),
				Visibility: "public",
				Contract:   "interfaces/openapi.yaml",
			},
			{
				Name:       "grpc-api",
				Type:       "grpc",
				Port:       intPtr(9090),
				Visibility: "internal",
				Contract:   "interfaces/service.proto",
			},
			{
				Name:       "order-events",
				Type:       "event",
				Visibility: "internal",
				Contract:   "interfaces/events.yaml",
			},
		},
		Configuration: &contract.Configuration{
			Schema: "configuration/schema.json",
		},
		Dependencies: []contract.Dependency{
			{
				Ref:           "ghcr.io/acme/auth-service-pacto@sha256:abc123",
				Required:      true,
				Compatibility: "^2.0.0",
			},
			{
				Ref:           "ghcr.io/acme/notification-service-pacto:1.0.0",
				Required:      false,
				Compatibility: "~1.0.0",
			},
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State: contract.State{
				Type:            "stateful",
				DataCriticality: "high",
				Persistence: contract.Persistence{
					Scope:      "shared",
					Durability: "persistent",
				},
			},
			Lifecycle: &contract.Lifecycle{
				UpgradeStrategy:         "rolling",
				GracefulShutdownSeconds: intPtr(30),
			},
			Health: contract.Health{
				Interface:           "rest-api",
				Path:                "/health",
				InitialDelaySeconds: intPtr(15),
			},
		},
		Scaling: &contract.Scaling{Min: 2, Max: 10},
		Metadata: map[string]interface{}{
			"team": "payments",
			"tier": "critical",
		},
	}
}

func fullFS() fstest.MapFS {
	return fstest.MapFS{
		"interfaces/openapi.yaml": &fstest.MapFile{Data: []byte(`
openapi: "3.0.0"
paths:
  /health:
    get:
      summary: Health check
  /payments:
    post:
      summary: Create a payment
`)},
		"interfaces/events.yaml": &fstest.MapFile{Data: []byte(`
description: Order placement events
`)},
		"configuration/schema.json": &fstest.MapFile{Data: []byte(`{
  "type": "object",
  "properties": {
    "PORT": {
      "type": "integer",
      "description": "HTTP server port",
      "default": 8080
    },
    "REDIS_URL": {
      "type": "string",
      "description": "Redis connection string"
    }
  },
  "required": ["PORT", "REDIS_URL"]
}`)},
	}
}

func TestGenerate_Full(t *testing.T) {
	c := fullContract()
	fsys := fullFS()

	md, err := Generate(c, fsys, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain := []struct {
		name, substr string
	}{
		{"service heading", "# payments-api"},
		{"description paragraph", "**payments-api** `v2.1.0` is a `stateful` `service` workload exposing 3 interfaces with 2 dependencies."},
		{"owner in description", "Owned by `team/payments`"},
		{"scaling in description", "scales from `2` to `10` replicas"},
		{"concept table header", "| Concept | Value | Description |"},
		{"workload row", "| **Workload** | `service` |"},
		{"state row", "| **State** | `stateful` |"},
		{"stateful explanation", "Retains data between requests"},
		{"persistence scope row", "| **Persistence scope** | `shared` |"},
		{"persistence durability row", "| **Persistence durability** | `persistent` |"},
		{"persistent explanation", "must survive restarts"},
		{"data criticality row", "| **Data criticality** | `high` |"},
		{"high criticality explanation", "severe business impact"},
		{"upgrade strategy row", "| **Upgrade strategy** | `rolling` |"},
		{"contract reference link", "[Contract Reference](https://trianalab.github.io/pacto/contract-reference/)"},
		{"TOC heading", "## Table of Contents"},
		{"TOC Architecture link", "- [Architecture](#architecture)"},
		{"TOC Interfaces link", "- [Interfaces](#interfaces)"},
		{"TOC Configuration link", "- [Configuration](#configuration)"},
		{"TOC Dependencies link", "- [Dependencies](#dependencies)"},
		{"TOC HTTP sub-link", "  - [HTTP Interface: rest-api](#http-interface-rest-api)"},
		{"TOC gRPC sub-link", "  - [gRPC Interface: grpc-api](#grpc-interface-grpc-api)"},
		{"TOC Event sub-link", "  - [Event Interface: order-events](#event-interface-order-events)"},
		{"architecture section", "## Architecture"},
		{"mermaid block", "```mermaid"},
		{"mermaid graph direction", "graph TD"},
		{"auth dep in mermaid", "payments-api --> auth-service-pacto"},
		{"notification dep in mermaid", "payments-api --> notification-service-pacto"},
		{"interfaces section", "## Interfaces"},
		{"rest-api in interfaces table", "| `rest-api` | `http` | `8080` | `public` |"},
		{"configuration section", "## Configuration"},
		{"PORT property in configuration", "| `PORT` | `integer` | HTTP server port | `8080` | Yes |"},
		{"HTTP interface subsection", "### HTTP Interface: rest-api"},
		{"endpoints heading", "#### Endpoints"},
		{"GET /health endpoint", "| `GET` | `/health` | Health check |"},
		{"POST /payments endpoint", "| `POST` | `/payments` | Create a payment |"},
		{"gRPC interface subsection", "### gRPC Interface: grpc-api"},
		{"gRPC contract reference", "Its contract is defined in `interfaces/service.proto`"},
		{"Event interface subsection", "### Event Interface: order-events"},
		{"dependencies section", "## Dependencies"},
		{"auth dependency", "| `ghcr.io/acme/auth-service-pacto@sha256:abc123` | `^2.0.0` | Yes |"},
		{"notification dependency", "| `ghcr.io/acme/notification-service-pacto:1.0.0` | `~1.0.0` | No |"},
		{"image ref in description", "packaged as `ghcr.io/acme/payments-api:2.1.0`"},
		{"container image section", "## Container Image"},
		{"image ref", "**Ref:** `ghcr.io/acme/payments-api:2.1.0`"},
		{"private flag", "**Private:** Yes"},
		{"container image in TOC", "- [Container Image](#container-image)"},
		{"health path in interface", "owns the health path under `/health`"},
		{"initial delay in interface", "requires an initial delay of `15s`"},
		{"verbal description rest-api", "The `rest-api` interface is `public` and exposes port `8080`."},
		{"verbal description grpc-api", "The `grpc-api` interface is `internal` and exposes port `9090`."},
		{"verbal description order-events", "The `order-events` interface is `internal`."},
		{"graceful shutdown in concepts", "| **Graceful shutdown** | `30s` |"},
		{"team metadata tag", "`team: payments`"},
		{"tier metadata tag", "`tier: critical`"},
		{"Pacto footer", "Generated by [Pacto](https://trianalab.github.io/pacto)"},
	}

	for _, tc := range mustContain {
		t.Run("contains/"+tc.name, func(t *testing.T) {
			if !strings.Contains(md, tc.substr) {
				t.Errorf("expected %q in output", tc.substr)
			}
		})
	}

	mustNotContain := []struct {
		name, substr string
	}{
		{"no Overview link in TOC", "- [Overview]"},
		{"no overview section", "## Overview"},
		{"no metadata heading", "## Metadata"},
	}

	for _, tc := range mustNotContain {
		t.Run("excludes/"+tc.name, func(t *testing.T) {
			if strings.Contains(md, tc.substr) {
				t.Errorf("unexpected %q in output", tc.substr)
			}
		})
	}
}

func TestGenerate_Minimal(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service: contract.ServiceIdentity{
			Name:    "simple-svc",
			Version: "1.0.0",
		},
		Interfaces: []contract.Interface{
			{Name: "api", Type: "http", Port: intPtr(8080)},
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State: contract.State{
				Type:            "stateless",
				DataCriticality: "low",
			},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain := []struct {
		name, substr string
	}{
		{"service heading", "# simple-svc"},
		{"description paragraph", "**simple-svc** `v1.0.0` is a `stateless` `service` workload exposing 1 interface."},
		{"workload row", "| **Workload** | `service` |"},
		{"state row", "| **State** | `stateless` |"},
		{"stateless explanation", "Does not retain data"},
		{"TOC heading", "## Table of Contents"},
		{"mermaid block", "```mermaid"},
		{"standalone node in mermaid", "  simple-svc\n"},
	}

	for _, tc := range mustContain {
		t.Run("contains/"+tc.name, func(t *testing.T) {
			if !strings.Contains(md, tc.substr) {
				t.Errorf("expected %q in output", tc.substr)
			}
		})
	}

	mustNotContain := []struct {
		name, substr string
	}{
		{"no persistence durability", "| **Persistence durability**"},
		{"no persistence scope", "| **Persistence scope**"},
		{"no Overview link", "- [Overview]"},
		{"no Dependencies link", "- [Dependencies]"},
		{"no Configuration link", "- [Configuration]"},
		{"no scaling replicas", "replicas"},
		{"no Configuration section", "## Configuration"},
		{"no Dependencies section", "## Dependencies"},
	}

	for _, tc := range mustNotContain {
		t.Run("excludes/"+tc.name, func(t *testing.T) {
			if strings.Contains(md, tc.substr) {
				t.Errorf("unexpected %q in output", tc.substr)
			}
		})
	}
}

func TestGenerate_MissingSpecFiles(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service: contract.ServiceIdentity{
			Name:    "svc",
			Version: "1.0.0",
		},
		Interfaces: []contract.Interface{
			{
				Name:     "api",
				Type:     "http",
				Port:     intPtr(8080),
				Contract: "interfaces/openapi.yaml",
			},
		},
		Configuration: &contract.Configuration{
			Schema: "configuration/schema.json",
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	// Empty FS — spec files don't exist.
	md, err := Generate(c, fstest.MapFS{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce inline notes, not fatal errors.
	if !strings.Contains(md, "_Could not read") {
		t.Error("expected inline error note for missing spec files")
	}
}

func TestGenerate_NoInterfaces(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(md, "## Interfaces") {
		t.Error("should not contain Interfaces section when there are none")
	}
}

func TestGenerate_InterfaceWithoutPort(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces: []contract.Interface{
			{Name: "events", Type: "event", Visibility: "internal"},
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Port should show as em-dash in interfaces table
	if !strings.Contains(md, "| `events` | `event` | \u2014 | `internal` |") {
		t.Errorf("expected em-dash for missing port, got:\n%s", md)
	}
}

func TestLoadSchemaDescriptions_InvalidJSON(t *testing.T) {
	dst := loadSchemaDescriptions([]byte("{invalid"))
	if len(dst) != 0 {
		t.Errorf("expected empty map for invalid JSON, got %v", dst)
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "Hello"},
		{"", ""},
		{"Hello", "Hello"},
	}
	for _, tt := range tests {
		got := capitalizeFirst(tt.input)
		if got != tt.want {
			t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInterfaceHeading_UnknownType(t *testing.T) {
	iface := contract.Interface{Name: "custom", Type: "websocket"}
	heading := interfaceHeading(iface)
	expected := "Websocket Interface: custom"
	if heading != expected {
		t.Errorf("expected %q, got %q", expected, heading)
	}
}

func TestGenerate_LifecycleWithEmptyUpgradeStrategy(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces:   []contract.Interface{{Name: "api", Type: "http", Port: intPtr(8080)}},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
			Lifecycle: &contract.Lifecycle{
				GracefulShutdownSeconds: intPtr(30),
			},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not contain upgrade strategy row when empty
	if strings.Contains(md, "**Upgrade strategy**") {
		t.Error("should not contain upgrade strategy row when empty")
	}
}

func TestGenerate_HTTPInterfaceWithoutContract(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces: []contract.Interface{
			{Name: "api", Type: "http", Port: intPtr(8080), Visibility: "public"},
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
			Health:   contract.Health{Interface: "api"},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have the interface heading but no endpoints section
	if !strings.Contains(md, "### HTTP Interface: api") {
		t.Error("expected HTTP interface subsection")
	}
	if strings.Contains(md, "#### Endpoints") {
		t.Error("should not contain endpoints section when no contract")
	}
}

func TestGenerate_HTTPInterfaceWithEmptySpec(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces: []contract.Interface{
			{Name: "api", Type: "http", Port: intPtr(8080), Contract: "interfaces/openapi.yaml"},
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	// Spec exists but has no paths
	fsys := fstest.MapFS{
		"interfaces/openapi.yaml": &fstest.MapFile{Data: []byte(`openapi: "3.0.0"
info:
  title: Empty API
`)},
	}

	md, err := Generate(c, fsys, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(md, "### HTTP Interface: api") {
		t.Error("expected HTTP interface subsection")
	}
	if strings.Contains(md, "#### Endpoints") {
		t.Error("should not contain endpoints section for empty spec")
	}
}

func TestExtractEnumDescriptions_NonObjectValue(t *testing.T) {
	props := map[string]interface{}{
		"name":    "not an object",
		"version": 42,
	}
	dst := make(map[string]string)
	extractEnumDescriptions(props, "", dst)
	if len(dst) != 0 {
		t.Errorf("expected empty map for non-object values, got %v", dst)
	}
}

func TestGenerate_ConfigurationSchemaError(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces:   []contract.Interface{{Name: "api", Type: "http", Port: intPtr(8080)}},
		Configuration: &contract.Configuration{
			Schema: "configuration/schema.json",
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	// FS with an empty schema that has properties: {}
	fsys := fstest.MapFS{
		"configuration/schema.json": &fstest.MapFile{Data: []byte(`{"type":"object","properties":{}}`)},
	}

	md, err := Generate(c, fsys, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty properties should not produce a Configuration section
	if strings.Contains(md, "## Configuration") {
		t.Error("should not contain Configuration section when properties are empty")
	}
}

func TestGenerate_ConfigPropertyWithoutDescription(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces:   []contract.Interface{{Name: "api", Type: "http", Port: intPtr(8080)}},
		Configuration: &contract.Configuration{
			Schema: "configuration/schema.json",
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	fsys := fstest.MapFS{
		"configuration/schema.json": &fstest.MapFile{Data: []byte(`{
  "type": "object",
  "properties": {
    "DEBUG": {"type": "boolean"}
  }
}`)},
	}

	md, err := Generate(c, fsys, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(md, "## Configuration") {
		t.Error("expected configuration section")
	}
	// Property without description should show em-dash
	if !strings.Contains(md, "| `DEBUG` | `boolean` | \u2014 |") {
		t.Errorf("expected DEBUG with em-dash description, got:\n%s", md)
	}
}

func TestGenerate_EndpointWithoutSummary(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces: []contract.Interface{
			{Name: "api", Type: "http", Port: intPtr(8080), Contract: "interfaces/openapi.yaml"},
		},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	fsys := fstest.MapFS{
		"interfaces/openapi.yaml": &fstest.MapFile{Data: []byte(`
openapi: "3.0.0"
paths:
  /items:
    get: {}
`)},
	}

	md, err := Generate(c, fsys, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Endpoint without summary should show em-dash
	if !strings.Contains(md, "| `GET` | `/items` | \u2014 |") {
		t.Errorf("expected em-dash for missing summary, got:\n%s", md)
	}
}

func TestDepName(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"ghcr.io/acme/auth-service-pacto@sha256:abc123", "auth-service-pacto"},
		{"ghcr.io/acme/notification-service-pacto:1.0.0", "notification-service-pacto"},
		{"simple-ref", "simple-ref"},
	}
	for _, tt := range tests {
		got := depName(tt.ref)
		if got != tt.want {
			t.Errorf("depName(%q) = %q, want %q", tt.ref, got, tt.want)
		}
	}
}

func TestWriteMermaidDiagram_WithGraphResult(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "frontend", Version: "1.0.0"},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}
	gr := &graph.Result{
		Root: &graph.Node{
			Name:    "frontend",
			Version: "1.0.0",
			Dependencies: []graph.Edge{
				{
					Ref: "reg/backend:1.0.0",
					Node: &graph.Node{
						Name:    "backend",
						Version: "1.0.0",
						Dependencies: []graph.Edge{
							{Ref: "reg/postgres:16.4.0", Node: &graph.Node{Name: "postgres", Version: "16.4.0"}},
							{Ref: "reg/keycloak:26.0.0", Shared: true, Node: &graph.Node{Name: "keycloak", Version: "26.0.0"}},
						},
					},
				},
				{
					Ref: "reg/keycloak:26.0.0",
					Node: &graph.Node{
						Name:    "keycloak",
						Version: "26.0.0",
						Dependencies: []graph.Edge{
							{Ref: "reg/postgres:16.4.0", Node: &graph.Node{Name: "postgres", Version: "16.4.0"}},
						},
					},
				},
			},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, gr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain := []string{
		"```mermaid",
		"graph TD",
		"frontend --> backend",
		"frontend --> keycloak",
		"backend --> postgres",
		"backend --> keycloak",
		"keycloak --> postgres",
	}
	for _, s := range mustContain {
		if !strings.Contains(md, s) {
			t.Errorf("expected %q in output:\n%s", s, md)
		}
	}
}

func TestWriteMermaidDiagram_StandaloneNode(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "standalone", Version: "1.0.0"},
		Interfaces:   []contract.Interface{{Name: "api", Type: "http", Port: intPtr(8080)}},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(md, "  standalone\n") {
		t.Errorf("expected standalone node in Mermaid, got:\n%s", md)
	}
}

func TestWriteMermaidDiagram_DuplicateEdges(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces:   []contract.Interface{{Name: "api", Type: "http", Port: intPtr(8080)}},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}
	depNode := &graph.Node{Name: "dep", Version: "1.0.0"}
	gr := &graph.Result{
		Root: &graph.Node{
			Name:    "svc",
			Version: "1.0.0",
			Dependencies: []graph.Edge{
				{Ref: "reg/dep:1.0.0", Node: depNode},
				{Ref: "reg/dep:1.0.0", Node: depNode, Shared: true},
			},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, gr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The edge should only appear once
	count := strings.Count(md, "svc --> dep")
	if count != 1 {
		t.Errorf("expected 1 occurrence of 'svc --> dep', got %d", count)
	}
}

func TestWriteMermaidDiagram_NilEdgeNode(t *testing.T) {
	c := &contract.Contract{
		PactoVersion: "1.0",
		Service:      contract.ServiceIdentity{Name: "svc", Version: "1.0.0"},
		Interfaces:   []contract.Interface{{Name: "api", Type: "http", Port: intPtr(8080)}},
		Runtime: contract.Runtime{
			Workload: "service",
			State:    contract.State{Type: "stateless", DataCriticality: "low"},
		},
	}
	gr := &graph.Result{
		Root: &graph.Node{
			Name:    "svc",
			Version: "1.0.0",
			Dependencies: []graph.Edge{
				{Ref: "reg/missing:1.0.0", Node: nil, Error: "not found"},
			},
		},
	}

	md, err := Generate(c, fstest.MapFS{}, gr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(md, "```mermaid") {
		t.Error("expected mermaid block")
	}
}

func TestWalkMermaidEdges_NilNode(t *testing.T) {
	var b strings.Builder
	walkMermaidEdges(&b, nil, map[string]bool{})
	if b.Len() != 0 {
		t.Errorf("expected empty output for nil node, got %q", b.String())
	}
}
