package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

// TestOpenAPIAsyncAPIRule_RESTEndpoints tests OpenAPI enforcement for REST endpoints
func TestOpenAPIAsyncAPIRule_RESTEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(dir string) ([]walker.FileInfo, error)
		requireOpenAPI bool
		wantViolations bool
		description    string
	}{
		{
			name:           "Go REST endpoint with OpenAPI spec",
			requireOpenAPI: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create a Go file with REST endpoint
				mainGoPath := filepath.Join(dir, "main.go")
				if err := os.WriteFile(mainGoPath, []byte(`
package main

import "net/http"

func main() {
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("users"))
	})
}
`), 0644); err != nil {
					return nil, err
				}

				// Create OpenAPI spec
				openapiPath := filepath.Join(dir, "openapi.yaml")
				if err := os.WriteFile(openapiPath, []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /api/users:
    get:
      summary: Get users
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: mainGoPath, ParentPath: dir, IsDir: false},
					{Path: openapiPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: false,
			description:    "Should pass when REST endpoint has OpenAPI spec",
		},
		{
			name:           "Go REST endpoint without OpenAPI spec",
			requireOpenAPI: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create a Go file with REST endpoint but no spec
				mainGoPath := filepath.Join(dir, "main.go")
				if err := os.WriteFile(mainGoPath, []byte(`
package main

import "net/http"

func main() {
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("users"))
	})
}
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: mainGoPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: true,
			description:    "Should fail when REST endpoint lacks OpenAPI spec",
		},
		{
			name:           "Python Flask endpoint with OpenAPI spec",
			requireOpenAPI: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create a Python file with Flask endpoint
				appPyPath := filepath.Join(dir, "app.py")
				if err := os.WriteFile(appPyPath, []byte(`
from flask import Flask

app = Flask(__name__)

@app.route("/api/users")
def get_users():
    return {"users": []}
`), 0644); err != nil {
					return nil, err
				}

				// Create OpenAPI spec
				openapiPath := filepath.Join(dir, "openapi.yaml")
				if err := os.WriteFile(openapiPath, []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: appPyPath, ParentPath: dir, IsDir: false},
					{Path: openapiPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: false,
			description:    "Should pass when Flask app has OpenAPI spec",
		},
		{
			name:           "No REST endpoints",
			requireOpenAPI: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create a non-REST file
				utilsPyPath := filepath.Join(dir, "utils.py")
				if err := os.WriteFile(utilsPyPath, []byte(`
def calculate(a, b):
    return a + b
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: utilsPyPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: false,
			description:    "Should pass when no REST endpoints are detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			files, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			rule := NewOpenAPIAsyncAPIRule(OpenAPIAsyncAPIRule{
				RequireOpenAPI: tt.requireOpenAPI,
			})

			// Act
			violations := rule.Check(files, make(map[string]*walker.DirInfo))

			// Assert
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestOpenAPIAsyncAPIRule_EventDrivenPatterns tests AsyncAPI enforcement for event-driven systems
func TestOpenAPIAsyncAPIRule_EventDrivenPatterns(t *testing.T) {
	tests := []struct {
		name            string
		setupFiles      func(dir string) ([]walker.FileInfo, error)
		requireAsyncAPI bool
		wantViolations  bool
		description     string
	}{
		{
			name:            "Go Kafka consumer with AsyncAPI spec",
			requireAsyncAPI: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create a Go file with Kafka consumer
				consumerGoPath := filepath.Join(dir, "consumer.go")
				if err := os.WriteFile(consumerGoPath, []byte(`
package main

import "github.com/confluentinc/confluent-kafka-go/kafka"

func main() {
	consumer, _ := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	consumer.Subscribe("orders", nil)
}
`), 0644); err != nil {
					return nil, err
				}

				// Create AsyncAPI spec
				asyncapiPath := filepath.Join(dir, "asyncapi.yaml")
				if err := os.WriteFile(asyncapiPath, []byte(`
asyncapi: 2.6.0
info:
  title: Order Service
  version: 1.0.0
channels:
  orders:
    subscribe:
      message:
        payload:
          type: object
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: consumerGoPath, ParentPath: dir, IsDir: false},
					{Path: asyncapiPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: false,
			description:    "Should pass when Kafka consumer has AsyncAPI spec",
		},
		{
			name:            "Python Celery task without AsyncAPI spec",
			requireAsyncAPI: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create a Python file with Celery task but no spec
				tasksPyPath := filepath.Join(dir, "tasks.py")
				if err := os.WriteFile(tasksPyPath, []byte(`
from celery import Celery

app = Celery('tasks')

@app.task
def process_order(order_id):
    pass
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: tasksPyPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: true,
			description:    "Should fail when Celery task lacks AsyncAPI spec",
		},
		{
			name:            "No event-driven patterns",
			requireAsyncAPI: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create a non-event-driven file
				utilsGoPath := filepath.Join(dir, "utils.go")
				if err := os.WriteFile(utilsGoPath, []byte(`
package utils

func Add(a, b int) int {
    return a + b
}
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: utilsGoPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: false,
			description:    "Should pass when no event-driven patterns are detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			files, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			rule := NewOpenAPIAsyncAPIRule(OpenAPIAsyncAPIRule{
				RequireAsyncAPI: tt.requireAsyncAPI,
			})

			// Act
			violations := rule.Check(files, make(map[string]*walker.DirInfo))

			// Assert
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestOpenAPIAsyncAPIRule_Name tests the rule name
func TestOpenAPIAsyncAPIRule_Name(t *testing.T) {
	// Arrange
	rule := NewOpenAPIAsyncAPIRule(OpenAPIAsyncAPIRule{})

	// Act
	name := rule.Name()

	// Assert
	if name != "api-spec" {
		t.Errorf("Expected rule name to be 'api-spec', got '%s'", name)
	}
}
