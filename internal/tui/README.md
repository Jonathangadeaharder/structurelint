# Interactive TUI Mode

## Overview

This package provides a rich terminal user interface for navigating and fixing violations.

## Components

- **model.go**: Bubbletea model implementing the TUI with 4 view modes

## View Modes

1. **List View**: Navigate through all violations
2. **Detail View**: See detailed information about a specific violation
3. **Fix Preview**: Preview auto-fix changes before applying
4. **Graph View**: Visual dependency graph (placeholder)

## Keyboard Navigation

- **j/k or ↑/↓**: Navigate up/down
- **Enter**: View details
- **f**: Preview fix
- **a**: Apply fix
- **g**: View graph
- **q**: Quit/back

## Technology

Built using:
- [Bubbletea](https://github.com/charmbracelet/bubbletea): The Elm Architecture for Go
- [Lipgloss](https://github.com/charmbracelet/lipgloss): Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles): TUI components

## Usage

```bash
structurelint tui
structurelint tui --fixable-only
```
