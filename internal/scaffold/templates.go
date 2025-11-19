package scaffold

// registerBuiltInTemplates registers all built-in templates
func (g *Generator) registerBuiltInTemplates() {
	// Go templates
	g.registerTemplate(goServiceTemplate())
	g.registerTemplate(goRepositoryTemplate())
	g.registerTemplate(goHandlerTemplate())
	g.registerTemplate(goModelTemplate())

	// TypeScript templates
	g.registerTemplate(tsServiceTemplate())
	g.registerTemplate(tsControllerTemplate())
	g.registerTemplate(tsModelTemplate())

	// Python templates
	g.registerTemplate(pyServiceTemplate())
	g.registerTemplate(pyRepositoryTemplate())
	g.registerTemplate(pyModelTemplate())
}

// registerTemplate registers a template
func (g *Generator) registerTemplate(tmpl *Template) {
	key := string(tmpl.Language) + "-" + string(tmpl.Type)
	g.templates[key] = tmpl
}

// GetTemplate gets a template by language and type
func (g *Generator) GetTemplate(lang Language, typ TemplateType) (*Template, error) {
	key := string(lang) + "-" + string(typ)
	tmpl, ok := g.templates[key]
	if !ok {
		return nil, ErrTemplateNotFound(key)
	}
	return tmpl, nil
}

// ErrTemplateNotFound creates a template not found error
func ErrTemplateNotFound(key string) error {
	return &TemplateNotFoundError{Key: key}
}

// TemplateNotFoundError represents a template not found error
type TemplateNotFoundError struct {
	Key string
}

func (e *TemplateNotFoundError) Error() string {
	return "template not found: " + e.Key
}

// Go Templates

func goServiceTemplate() *Template {
	return &Template{
		Type:        TypeService,
		Language:    LangGo,
		Name:        "Go Service",
		Description: "Go service with business logic layer",
		Files: []TemplateFile{
			{
				Path: "internal/services/{{.NameSnake}}.go",
				Content: `package services

import (
	"context"
	"fmt"
)

// {{.Name}} handles {{.Description}}
type {{.Name}} struct {
	// Add dependencies here
	// repo repository.{{.Name}}Repository
}

// New{{.Name}} creates a new {{.Name}}
func New{{.Name}}() *{{.Name}} {
	return &{{.Name}}{
		// Initialize dependencies
	}
}

// Get retrieves a resource by ID
func (s *{{.Name}}) Get(ctx context.Context, id string) error {
	// TODO: Implement business logic
	return fmt.Errorf("not implemented")
}

// Create creates a new resource
func (s *{{.Name}}) Create(ctx context.Context, data interface{}) error {
	// TODO: Implement business logic
	return fmt.Errorf("not implemented")
}

// Update updates an existing resource
func (s *{{.Name}}) Update(ctx context.Context, id string, data interface{}) error {
	// TODO: Implement business logic
	return fmt.Errorf("not implemented")
}

// Delete deletes a resource
func (s *{{.Name}}) Delete(ctx context.Context, id string) error {
	// TODO: Implement business logic
	return fmt.Errorf("not implemented")
}
`,
			},
			{
				Path:   "internal/services/{{.NameSnake}}_test.go",
				IsTest: true,
				Content: `package services

import (
	"context"
	"testing"
)

func Test{{.Name}}_Get(t *testing.T) {
	s := New{{.Name}}()
	ctx := context.Background()

	err := s.Get(ctx, "test-id")
	if err == nil {
		t.Error("expected error for unimplemented method")
	}
}

func Test{{.Name}}_Create(t *testing.T) {
	s := New{{.Name}}()
	ctx := context.Background()

	err := s.Create(ctx, nil)
	if err == nil {
		t.Error("expected error for unimplemented method")
	}
}
`,
			},
		},
	}
}

func goRepositoryTemplate() *Template {
	return &Template{
		Type:        TypeRepository,
		Language:    LangGo,
		Name:        "Go Repository",
		Description: "Go repository for data access layer",
		Files: []TemplateFile{
			{
				Path: "internal/repository/{{.NameSnake}}.go",
				Content: `package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// {{.Name}}Repository handles data access for {{.Description}}
type {{.Name}}Repository interface {
	GetByID(ctx context.Context, id string) (interface{}, error)
	Create(ctx context.Context, data interface{}) error
	Update(ctx context.Context, id string, data interface{}) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]interface{}, error)
}

// {{.NameLower}}Repository implements {{.Name}}Repository
type {{.NameLower}}Repository struct {
	db *sql.DB
}

// New{{.Name}}Repository creates a new repository
func New{{.Name}}Repository(db *sql.DB) {{.Name}}Repository {
	return &{{.NameLower}}Repository{
		db: db,
	}
}

// GetByID retrieves a record by ID
func (r *{{.NameLower}}Repository) GetByID(ctx context.Context, id string) (interface{}, error) {
	// TODO: Implement database query
	return nil, fmt.Errorf("not implemented")
}

// Create inserts a new record
func (r *{{.NameLower}}Repository) Create(ctx context.Context, data interface{}) error {
	// TODO: Implement database insert
	return fmt.Errorf("not implemented")
}

// Update updates an existing record
func (r *{{.NameLower}}Repository) Update(ctx context.Context, id string, data interface{}) error {
	// TODO: Implement database update
	return fmt.Errorf("not implemented")
}

// Delete removes a record
func (r *{{.NameLower}}Repository) Delete(ctx context.Context, id string) error {
	// TODO: Implement database delete
	return fmt.Errorf("not implemented")
}

// List retrieves all records
func (r *{{.NameLower}}Repository) List(ctx context.Context) ([]interface{}, error) {
	// TODO: Implement database query
	return nil, fmt.Errorf("not implemented")
}
`,
			},
		},
	}
}

func goHandlerTemplate() *Template {
	return &Template{
		Type:        TypeHandler,
		Language:    LangGo,
		Name:        "Go HTTP Handler",
		Description: "Go HTTP handler for REST API",
		Files: []TemplateFile{
			{
				Path: "internal/handlers/{{.NameSnake}}.go",
				Content: `package handlers

import (
	"encoding/json"
	"net/http"

	"{{.Package}}/internal/services"
)

// {{.Name}}Handler handles HTTP requests for {{.Description}}
type {{.Name}}Handler struct {
	service *services.{{.Name}}
}

// New{{.Name}}Handler creates a new handler
func New{{.Name}}Handler(service *services.{{.Name}}) *{{.Name}}Handler {
	return &{{.Name}}Handler{
		service: service,
	}
}

// HandleGet handles GET requests
func (h *{{.Name}}Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	err := h.service.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// HandlePost handles POST requests
func (h *{{.Name}}Handler) HandlePost(w http.ResponseWriter, r *http.Request) {
	var data interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.Create(r.Context(), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

// RegisterRoutes registers HTTP routes
func (h *{{.Name}}Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/{{.NameKebab}}", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleGet(w, r)
		case http.MethodPost:
			h.HandlePost(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
`,
			},
		},
	}
}

func goModelTemplate() *Template {
	return &Template{
		Type:        TypeModel,
		Language:    LangGo,
		Name:        "Go Model",
		Description: "Go domain model/entity",
		Files: []TemplateFile{
			{
				Path: "internal/models/{{.NameSnake}}.go",
				Content: `package models

import "time"

// {{.Name}} represents {{.Description}}
type {{.Name}} struct {
	ID        string    ` + "`json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `

	// Add your fields here
}

// Validate validates the {{.Name}}
func (m *{{.Name}}) Validate() error {
	// TODO: Add validation logic
	return nil
}
`,
			},
		},
	}
}

// TypeScript Templates

func tsServiceTemplate() *Template {
	return &Template{
		Type:        TypeService,
		Language:    LangTypeScript,
		Name:        "TypeScript Service",
		Description: "TypeScript service class",
		Files: []TemplateFile{
			{
				Path: "src/services/{{.NameKebab}}.service.ts",
				Content: `/**
 * {{.Name}} - {{.Description}}
 * @author {{.Author}}
 */

export class {{.Name}} {
  constructor() {
    // Initialize dependencies
  }

  /**
   * Get a resource by ID
   */
  async get(id: string): Promise<any> {
    // TODO: Implement business logic
    throw new Error('Not implemented');
  }

  /**
   * Create a new resource
   */
  async create(data: any): Promise<any> {
    // TODO: Implement business logic
    throw new Error('Not implemented');
  }

  /**
   * Update an existing resource
   */
  async update(id: string, data: any): Promise<any> {
    // TODO: Implement business logic
    throw new Error('Not implemented');
  }

  /**
   * Delete a resource
   */
  async delete(id: string): Promise<void> {
    // TODO: Implement business logic
    throw new Error('Not implemented');
  }
}
`,
			},
			{
				Path:   "src/services/{{.NameKebab}}.service.test.ts",
				IsTest: true,
				Content: `import { {{.Name}} } from './{{.NameKebab}}.service';

describe('{{.Name}}', () => {
  let service: {{.Name}};

  beforeEach(() => {
    service = new {{.Name}}();
  });

  it('should be defined', () => {
    expect(service).toBeDefined();
  });

  it('should throw error on get (not implemented)', async () => {
    await expect(service.get('test-id')).rejects.toThrow('Not implemented');
  });

  it('should throw error on create (not implemented)', async () => {
    await expect(service.create({})).rejects.toThrow('Not implemented');
  });
});
`,
			},
		},
	}
}

func tsControllerTemplate() *Template {
	return &Template{
		Type:        TypeController,
		Language:    LangTypeScript,
		Name:        "TypeScript Controller",
		Description: "TypeScript REST controller",
		Files: []TemplateFile{
			{
				Path: "src/controllers/{{.NameKebab}}.controller.ts",
				Content: `import { Request, Response } from 'express';
import { {{.Name}} } from '../services/{{.NameKebab}}.service';

/**
 * {{.Name}}Controller - Handles HTTP requests for {{.Description}}
 */
export class {{.Name}}Controller {
  private service: {{.Name}};

  constructor(service: {{.Name}}) {
    this.service = service;
  }

  /**
   * GET /api/{{.NameKebab}}/:id
   */
  async get(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const result = await this.service.get(id);
      res.json(result);
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * POST /api/{{.NameKebab}}
   */
  async create(req: Request, res: Response): Promise<void> {
    try {
      const result = await this.service.create(req.body);
      res.status(201).json(result);
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * PUT /api/{{.NameKebab}}/:id
   */
  async update(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      const result = await this.service.update(id, req.body);
      res.json(result);
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  }

  /**
   * DELETE /api/{{.NameKebab}}/:id
   */
  async delete(req: Request, res: Response): Promise<void> {
    try {
      const { id } = req.params;
      await this.service.delete(id);
      res.status(204).send();
    } catch (error) {
      res.status(500).json({ error: error.message });
    }
  }
}
`,
			},
		},
	}
}

func tsModelTemplate() *Template {
	return &Template{
		Type:        TypeModel,
		Language:    LangTypeScript,
		Name:        "TypeScript Model",
		Description: "TypeScript domain model/interface",
		Files: []TemplateFile{
			{
				Path: "src/models/{{.NameKebab}}.model.ts",
				Content: `/**
 * {{.Name}} - {{.Description}}
 */
export interface {{.Name}} {
  id: string;
  createdAt: Date;
  updatedAt: Date;

  // Add your fields here
}

/**
 * Create{{.Name}}Input - Input for creating a {{.Name}}
 */
export interface Create{{.Name}}Input {
  // Add required fields here
}

/**
 * Update{{.Name}}Input - Input for updating a {{.Name}}
 */
export interface Update{{.Name}}Input {
  // Add updatable fields here
}
`,
			},
		},
	}
}

// Python Templates

func pyServiceTemplate() *Template {
	return &Template{
		Type:        TypeService,
		Language:    LangPython,
		Name:        "Python Service",
		Description: "Python service class",
		Files: []TemplateFile{
			{
				Path: "services/{{.NameSnake}}.py",
				Content: `"""
{{.Name}} - {{.Description}}
Author: {{.Author}}
"""

from typing import Any, Optional


class {{.Name}}:
    """{{.Name}} handles {{.Description}}"""

    def __init__(self):
        """Initialize the service"""
        # Initialize dependencies
        pass

    def get(self, id: str) -> Optional[Any]:
        """Get a resource by ID"""
        # TODO: Implement business logic
        raise NotImplementedError("Method not implemented")

    def create(self, data: Any) -> Any:
        """Create a new resource"""
        # TODO: Implement business logic
        raise NotImplementedError("Method not implemented")

    def update(self, id: str, data: Any) -> Any:
        """Update an existing resource"""
        # TODO: Implement business logic
        raise NotImplementedError("Method not implemented")

    def delete(self, id: str) -> None:
        """Delete a resource"""
        # TODO: Implement business logic
        raise NotImplementedError("Method not implemented")
`,
			},
			{
				Path:   "tests/test_{{.NameSnake}}.py",
				IsTest: true,
				Content: `"""
Tests for {{.Name}}
"""

import pytest
from services.{{.NameSnake}} import {{.Name}}


@pytest.fixture
def service():
    """Create a {{.Name}} instance"""
    return {{.Name}}()


def test_service_creation(service):
    """Test that service can be created"""
    assert service is not None


def test_get_not_implemented(service):
    """Test that get raises NotImplementedError"""
    with pytest.raises(NotImplementedError):
        service.get("test-id")


def test_create_not_implemented(service):
    """Test that create raises NotImplementedError"""
    with pytest.raises(NotImplementedError):
        service.create({})
`,
			},
		},
	}
}

func pyRepositoryTemplate() *Template {
	return &Template{
		Type:        TypeRepository,
		Language:    LangPython,
		Name:        "Python Repository",
		Description: "Python repository for data access",
		Files: []TemplateFile{
			{
				Path: "repositories/{{.NameSnake}}.py",
				Content: `"""
{{.Name}}Repository - Data access layer for {{.Description}}
"""

from abc import ABC, abstractmethod
from typing import Any, List, Optional


class {{.Name}}Repository(ABC):
    """Abstract repository interface"""

    @abstractmethod
    def get_by_id(self, id: str) -> Optional[Any]:
        """Get a record by ID"""
        pass

    @abstractmethod
    def create(self, data: Any) -> Any:
        """Create a new record"""
        pass

    @abstractmethod
    def update(self, id: str, data: Any) -> Any:
        """Update an existing record"""
        pass

    @abstractmethod
    def delete(self, id: str) -> None:
        """Delete a record"""
        pass

    @abstractmethod
    def list(self) -> List[Any]:
        """List all records"""
        pass


class {{.Name}}RepositoryImpl({{.Name}}Repository):
    """Concrete implementation of {{.Name}}Repository"""

    def __init__(self, db):
        """Initialize repository with database connection"""
        self.db = db

    def get_by_id(self, id: str) -> Optional[Any]:
        """Get a record by ID"""
        # TODO: Implement database query
        raise NotImplementedError("Method not implemented")

    def create(self, data: Any) -> Any:
        """Create a new record"""
        # TODO: Implement database insert
        raise NotImplementedError("Method not implemented")

    def update(self, id: str, data: Any) -> Any:
        """Update an existing record"""
        # TODO: Implement database update
        raise NotImplementedError("Method not implemented")

    def delete(self, id: str) -> None:
        """Delete a record"""
        # TODO: Implement database delete
        raise NotImplementedError("Method not implemented")

    def list(self) -> List[Any]:
        """List all records"""
        # TODO: Implement database query
        raise NotImplementedError("Method not implemented")
`,
			},
		},
	}
}

func pyModelTemplate() *Template {
	return &Template{
		Type:        TypeModel,
		Language:    LangPython,
		Name:        "Python Model",
		Description: "Python domain model/dataclass",
		Files: []TemplateFile{
			{
				Path: "models/{{.NameSnake}}.py",
				Content: `"""
{{.Name}} - {{.Description}}
"""

from dataclasses import dataclass
from datetime import datetime
from typing import Optional


@dataclass
class {{.Name}}:
    """{{.Name}} domain model"""

    id: str
    created_at: datetime
    updated_at: datetime

    # Add your fields here

    def validate(self) -> bool:
        """Validate the model"""
        # TODO: Add validation logic
        return True

    def to_dict(self) -> dict:
        """Convert to dictionary"""
        return {
            'id': self.id,
            'created_at': self.created_at.isoformat(),
            'updated_at': self.updated_at.isoformat(),
            # Add your fields here
        }

    @classmethod
    def from_dict(cls, data: dict) -> '{{.Name}}':
        """Create from dictionary"""
        return cls(
            id=data['id'],
            created_at=datetime.fromisoformat(data['created_at']),
            updated_at=datetime.fromisoformat(data['updated_at']),
            # Add your fields here
        )
`,
			},
		},
	}
}
