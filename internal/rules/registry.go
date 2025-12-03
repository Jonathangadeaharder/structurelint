package rules

import (
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
)

// RuleContext contains context available to rules during instantiation
type RuleContext struct {
	RootDir     string
	ImportGraph *graph.ImportGraph
	Config      map[string]interface{} // The raw config map for this rule
}

// RuleFactory is a function that creates a rule instance
type RuleFactory func(ctx *RuleContext) (Rule, error)

var (
	registry = make(map[string]RuleFactory)
)

// Register registers a rule factory
func Register(name string, factory RuleFactory) {
	registry[name] = factory
}

// GetFactory returns a rule factory by name
func GetFactory(name string) (RuleFactory, bool) {
	factory, ok := registry[name]
	return factory, ok
}

// Helper to get int from config
func (c *RuleContext) GetInt(key string) (int, bool) {
	if val, ok := c.Config[key].(int); ok {
		return val, true
	}
	if val, ok := c.Config[key].(float64); ok {
		return int(val), true
	}
	return 0, false
}

// Helper to get string map from config
func (c *RuleContext) GetStringMap(key string) (map[string]string, bool) {
	// If the config itself is the map (common for some rules)
	if key == "" {
		result := make(map[string]string)
		for k, v := range c.Config {
			if strVal, ok := v.(string); ok {
				result[k] = strVal
			}
		}
		return result, len(result) > 0
	}
	
	// Otherwise look for key
	if val, ok := c.Config[key].(map[string]interface{}); ok {
		result := make(map[string]string)
		for k, v := range val {
			if strVal, ok := v.(string); ok {
				result[k] = strVal
			}
		}
		return result, true
	}
	return nil, false
}

// Helper to get string slice from config
func (c *RuleContext) GetStringSlice(key string) ([]string, bool) {
	// If the config itself is the slice
	if key == "" {
		// This case is tricky because Config is map[string]interface{}
		// Usually rules that are just a list are parsed differently before here
		return nil, false
	}

	if val, ok := c.Config[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if strVal, ok := v.(string); ok {
				result = append(result, strVal)
			}
		}
		return result, true
	}
	return nil, false
}
