// Package rules provides API specification enforcement rules.
package rules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"gopkg.in/yaml.v3"
)

// OpenAPIAsyncAPIRule enforces the presence of API specifications for REST and event-driven systems.
// It checks for:
// - OpenAPI/Swagger specifications for REST endpoints
// - AsyncAPI specifications for event-driven/asynchronous systems
type OpenAPIAsyncAPIRule struct {
	RequireOpenAPI  bool     `yaml:"require-openapi"`
	RequireAsyncAPI bool     `yaml:"require-asyncapi"`
	CustomSpecs     []string `yaml:"custom-specs"`
}

// Name returns the rule name
func (r *OpenAPIAsyncAPIRule) Name() string {
	return "api-spec"
}

// Check validates API specification requirements
func (r *OpenAPIAsyncAPIRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Detect if the project uses REST endpoints
	hasRESTEndpoints := r.detectRESTEndpoints(files)

	// Detect if the project uses event-driven/async patterns
	hasEventDriven := r.detectEventDrivenPatterns(files)

	// If OpenAPI is required and REST endpoints are detected, check for OpenAPI specs
	if r.RequireOpenAPI && hasRESTEndpoints {
		if !r.hasOpenAPISpec(files) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".",
				Message: r.formatMissingOpenAPIMessage(),
			})
		}
	}

	// If AsyncAPI is required and event-driven patterns are detected, check for AsyncAPI specs
	if r.RequireAsyncAPI && hasEventDriven {
		if !r.hasAsyncAPISpec(files) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".",
				Message: r.formatMissingAsyncAPIMessage(),
			})
		}
	}

	return violations
}

// detectRESTEndpoints detects if the project uses REST endpoints
func (r *OpenAPIAsyncAPIRule) detectRESTEndpoints(files []walker.FileInfo) bool {
	// REST endpoint indicators by language
	restIndicators := map[string][]string{
		// Go
		".go": {
			"net/http",
			"github.com/gin-gonic/gin",
			"github.com/gorilla/mux",
			"github.com/labstack/echo",
			"github.com/gofiber/fiber",
			"chi.Router",
			"http.HandleFunc",
			"http.Handle",
			"@app.route",       // Flask-style decorators in comments
			"@router.get",
			"@router.post",
			"@router.put",
			"@router.delete",
		},
		// Python
		".py": {
			"from flask import",
			"from fastapi import",
			"from django.urls import",
			"@app.route",
			"@router.get",
			"@router.post",
			"@router.put",
			"@router.delete",
			"@api_view",
			"class.*APIView",
			"class.*ViewSet",
		},
		// TypeScript/JavaScript
		".ts": {
			"express()",
			"fastify()",
			"@nestjs/common",
			"@Get(",
			"@Post(",
			"@Put(",
			"@Delete(",
			"router.get",
			"router.post",
			"app.get",
			"app.post",
		},
		".js": {
			"express()",
			"fastify()",
			"router.get",
			"router.post",
			"app.get",
			"app.post",
		},
		// Rust
		".rs": {
			"actix_web::",
			"rocket::",
			"warp::",
			"axum::",
			"#[get(",
			"#[post(",
			"Router::new",
		},
		// Java
		".java": {
			"@RestController",
			"@RequestMapping",
			"@GetMapping",
			"@PostMapping",
			"@PutMapping",
			"@DeleteMapping",
			"@Path(",
			"javax.ws.rs",
			"jakarta.ws.rs",
		},
	}

	return r.detectPatterns(files, restIndicators)
}

// detectEventDrivenPatterns detects if the project uses event-driven patterns
func (r *OpenAPIAsyncAPIRule) detectEventDrivenPatterns(files []walker.FileInfo) bool {
	// Event-driven indicators by language
	eventIndicators := map[string][]string{
		// Go
		".go": {
			"kafka",
			"rabbitmq",
			"nats",
			"pubsub",
			"EventBus",
			"MessageBroker",
			"kafka.NewProducer",
			"kafka.NewConsumer",
			"amqp.Dial",
		},
		// Python
		".py": {
			"kafka",
			"rabbitmq",
			"celery",
			"redis.pubsub",
			"from confluent_kafka import",
			"from pika import",
			"from kombu import",
			"@celery.task",
		},
		// TypeScript/JavaScript
		".ts": {
			"kafkajs",
			"amqplib",
			"redis.publish",
			"redis.subscribe",
			"EventEmitter",
			"@nestjs/microservices",
			"@MessagePattern",
			"@EventPattern",
		},
		".js": {
			"kafkajs",
			"amqplib",
			"redis.publish",
			"redis.subscribe",
			"EventEmitter",
		},
		// Rust
		".rs": {
			"rdkafka::",
			"lapin::",
			"tokio::sync::mpsc",
			"async_channel",
		},
		// Java
		".java": {
			"org.apache.kafka",
			"spring.kafka",
			"spring.amqp",
			"@KafkaListener",
			"@RabbitListener",
			"@EventListener",
		},
	}

	return r.detectPatterns(files, eventIndicators)
}

// detectPatterns checks if any files contain the specified patterns
func (r *OpenAPIAsyncAPIRule) detectPatterns(files []walker.FileInfo, patterns map[string][]string) bool {
	for _, file := range files {
		if file.IsDir {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Path))
		indicators, ok := patterns[ext]
		if !ok {
			continue
		}

		// Read file content
		content, err := os.ReadFile(file.Path)
		if err != nil {
			continue
		}

		contentStr := string(content)
		for _, indicator := range indicators {
			if strings.Contains(contentStr, indicator) {
				return true
			}
		}
	}

	return false
}

// hasOpenAPISpec checks if OpenAPI specification files exist
func (r *OpenAPIAsyncAPIRule) hasOpenAPISpec(files []walker.FileInfo) bool {
	openAPIFiles := []string{
		"openapi.yaml",
		"openapi.yml",
		"openapi.json",
		"swagger.yaml",
		"swagger.yml",
		"swagger.json",
		"api.yaml",
		"api.yml",
		"api.json",
	}

	for _, file := range files {
		if file.IsDir {
			continue
		}

		filename := strings.ToLower(filepath.Base(file.Path))
		for _, specFile := range openAPIFiles {
			if filename == specFile {
				// Verify it's actually an OpenAPI spec
				if r.isOpenAPISpec(file.Path) {
					return true
				}
			}
		}
	}

	return false
}

// hasAsyncAPISpec checks if AsyncAPI specification files exist
func (r *OpenAPIAsyncAPIRule) hasAsyncAPISpec(files []walker.FileInfo) bool {
	asyncAPIFiles := []string{
		"asyncapi.yaml",
		"asyncapi.yml",
		"asyncapi.json",
	}

	for _, file := range files {
		if file.IsDir {
			continue
		}

		filename := strings.ToLower(filepath.Base(file.Path))
		for _, specFile := range asyncAPIFiles {
			if filename == specFile {
				// Verify it's actually an AsyncAPI spec
				if r.isAsyncAPISpec(file.Path) {
					return true
				}
			}
		}
	}

	return false
}

// isOpenAPISpec verifies if a file is a valid OpenAPI specification
func (r *OpenAPIAsyncAPIRule) isOpenAPISpec(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	// Parse as YAML/JSON and check for OpenAPI markers
	var spec map[string]interface{}

	// Try YAML first
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return false
	}

	// Check for OpenAPI 3.x
	if openapi, ok := spec["openapi"].(string); ok && strings.HasPrefix(openapi, "3.") {
		return true
	}

	// Check for Swagger 2.0
	if swagger, ok := spec["swagger"].(string); ok && swagger == "2.0" {
		return true
	}

	return false
}

// isAsyncAPISpec verifies if a file is a valid AsyncAPI specification
func (r *OpenAPIAsyncAPIRule) isAsyncAPISpec(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	// Parse as YAML/JSON and check for AsyncAPI markers
	var spec map[string]interface{}

	if err := yaml.Unmarshal(data, &spec); err != nil {
		return false
	}

	// Check for AsyncAPI version
	if asyncapi, ok := spec["asyncapi"].(string); ok && asyncapi != "" {
		return true
	}

	return false
}

// formatMissingOpenAPIMessage creates a detailed error message for missing OpenAPI specification
func (r *OpenAPIAsyncAPIRule) formatMissingOpenAPIMessage() string {
	return "REST endpoints detected but no OpenAPI/Swagger specification found. " +
		"Expected one of: openapi.yaml, openapi.yml, openapi.json, swagger.yaml, swagger.yml, swagger.json. " +
		"OpenAPI specifications enable design-by-contract for REST APIs and improve API documentation."
}

// formatMissingAsyncAPIMessage creates a detailed error message for missing AsyncAPI specification
func (r *OpenAPIAsyncAPIRule) formatMissingAsyncAPIMessage() string {
	return "Event-driven/asynchronous patterns detected but no AsyncAPI specification found. " +
		"Expected one of: asyncapi.yaml, asyncapi.yml, asyncapi.json. " +
		"AsyncAPI is designed for message brokers and event streams, enabling design-by-contract for async systems."
}

// NewOpenAPIAsyncAPIRule creates a new OpenAPIAsyncAPIRule
func NewOpenAPIAsyncAPIRule(config OpenAPIAsyncAPIRule) *OpenAPIAsyncAPIRule {
	return &OpenAPIAsyncAPIRule{
		RequireOpenAPI:  config.RequireOpenAPI,
		RequireAsyncAPI: config.RequireAsyncAPI,
		CustomSpecs:     config.CustomSpecs,
	}
}
