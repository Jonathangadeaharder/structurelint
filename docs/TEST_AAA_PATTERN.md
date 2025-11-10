# Arrange-Act-Assert (AAA) Pattern Enforcement

## Overview

The Arrange-Act-Assert (AAA) pattern is a widely-used testing pattern that improves test readability and maintainability by clearly separating test phases:

- **Arrange**: Set up test data, mocks, and preconditions
- **Act**: Execute the code being tested
- **Assert**: Verify the results match expectations

structurelint can enforce this pattern through file content templates, ensuring all tests follow consistent, readable structure.

## Benefits

✅ **Improved Readability**: Tests become self-documenting with clear phases
✅ **Easier Debugging**: Quickly identify where test failures occur
✅ **Consistent Style**: All team members write tests the same way
✅ **Better Reviews**: Reviewers can quickly understand test intent
✅ **Teaching Tool**: New developers learn testing best practices

## Quick Start

### 1. Enable AAA Templates

Add to your `.structurelint.yml`:

```yaml
rules:
  file-content:
    templates:
      # Go tests
      "**/*_test.go": "test-go"

      # TypeScript/JavaScript tests
      "**/*.test.ts": "test-typescript"
      "**/*.test.js": "test-typescript"
      "**/*.spec.ts": "test-typescript"

      # Python tests
      "**/test_*.py": "test-python"
      "**/*_test.py": "test-python"
```

### 2. Write Compliant Tests

**Go Example:**

```go
func TestCalculator_Add_ReturnsCorrectSum(t *testing.T) {
    // Arrange
    calc := NewCalculator()
    a, b := 2, 3
    expected := 5

    // Act
    result := calc.Add(a, b)

    // Assert
    if result != expected {
        t.Errorf("Expected %d, got %d", expected, result)
    }
}
```

**TypeScript Example:**

```typescript
describe('Calculator', () => {
  it('should add two numbers correctly', () => {
    // Arrange
    const calc = new Calculator();
    const a = 2;
    const b = 3;
    const expected = 5;

    // Act
    const result = calc.add(a, b);

    // Assert
    expect(result).toBe(expected);
  });
});
```

**Python Example:**

```python
def test_calculator_add_returns_correct_sum():
    # Arrange
    calc = Calculator()
    a = 2
    b = 3
    expected = 5

    # Act
    result = calc.add(a, b)

    # Assert
    assert result == expected
```

## Available Templates

### Standard Templates (Recommended)

These templates encourage but don't strictly require AAA comments:

#### `test-go.yml`
- Validates Go test function exists
- Encourages AAA comments
- Discourages overly complex tests

#### `test-typescript.yml`
- Validates test blocks (it/test/describe)
- Encourages AAA comments
- Warns about missing expectations

#### `test-python.yml`
- Validates test functions exist
- Encourages AAA comments
- Discourages tests without assertions

### Strict Template

#### `test-strict-aaa.yml`
Enforces AAA pattern strictly:
- **Requires** all three AAA comments in every test
- Forbids TODO/FIXME/XXX markers
- Forbids debugging statements (console.log, print)
- Best for teams wanting maximum consistency

**Usage:**

```yaml
rules:
  file-content:
    templates:
      "**/*_test.go": "test-strict-aaa"
      "**/*.test.ts": "test-strict-aaa"
      "**/test_*.py": "test-strict-aaa"
```

## Template Configuration Options

### Choosing the Right Template

| Template | Use Case | Enforcement Level |
|----------|----------|-------------------|
| `test-go` | General Go projects | Lenient - encourages AAA |
| `test-typescript` | General JS/TS projects | Lenient - encourages AAA |
| `test-python` | General Python projects | Lenient - encourages AAA |
| `test-strict-aaa` | Teams wanting consistency | Strict - requires AAA |

### Custom Templates

Create your own template in `.structurelint/templates/`:

```yaml
# .structurelint/templates/test-custom.yml
required-patterns:
  - "def test_"
  - "# Arrange|# Act|# Assert"

forbidden-patterns:
  - "time\\.sleep"  # No sleep in tests
  - "random\\."     # No randomness in tests
```

## Best Practices

### 1. One Assertion Concept Per Test

**Good:**
```python
def test_calculator_add_returns_sum():
    # Arrange
    calc = Calculator()

    # Act
    result = calc.add(2, 3)

    # Assert
    assert result == 5
```

**Avoid:**
```python
def test_calculator_operations():  # Testing too much
    calc = Calculator()
    assert calc.add(2, 3) == 5
    assert calc.subtract(5, 2) == 3
    assert calc.multiply(2, 3) == 6
    # Split into separate tests!
```

### 2. Keep Arrange Section Focused

**Good:**
```typescript
it('should calculate total with tax', () => {
  // Arrange
  const calculator = new TaxCalculator();
  const subtotal = 100;
  const taxRate = 0.08;

  // Act
  const total = calculator.calculateTotal(subtotal, taxRate);

  // Assert
  expect(total).toBe(108);
});
```

**Avoid:**
```typescript
it('should calculate total with tax', () => {
  // Arrange - too complex
  const config = loadConfig();
  const db = connectDatabase();
  const user = createUser();
  const cart = createCart(user);
  addItemsToCart(cart, 10);
  // ... use fixtures or setup helpers instead
});
```

### 3. Single Act Statement

The Act phase should ideally be a single line or call:

```go
// Arrange
input := "test data"
processor := NewProcessor()

// Act
result := processor.Process(input)  // Single action

// Assert
assert.NotNil(t, result)
```

### 4. Use Comments Even for Simple Tests

Even simple tests benefit from AAA structure:

```python
def test_user_get_name_returns_name():
    # Arrange
    user = User(name="Alice")

    # Act
    result = user.get_name()

    # Assert
    assert result == "Alice"
```

## Integration with Test Validation

Combine AAA templates with Phase 3 test validation:

```yaml
rules:
  # Phase 3: Ensure tests exist
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "**/*.go"
    exemptions:
      - "cmd/**/*.go"

  # Phase 4: Ensure tests follow AAA pattern
  file-content:
    templates:
      "**/*_test.go": "test-go"
```

This ensures:
1. Every source file has tests (Phase 3)
2. Every test follows AAA pattern (Phase 4)

## Example Violations

### Missing AAA Comments

```
internal/calculator/calculator_test.go: missing required pattern "// Arrange|// Act|// Assert" (template: test-go)
```

**Fix:** Add AAA comments to your tests.

### Debugging Statement Left In

```
src/utils/helper.test.ts: contains forbidden pattern "console.log" (template: test-strict-aaa)
```

**Fix:** Remove console.log statements before committing.

### Missing Test Function

```
tests/test_processor.py: missing required pattern "def test_" (template: test-python)
```

**Fix:** Ensure file contains at least one test function.

## Exemptions

Some test files may not need AAA enforcement:

```yaml
rules:
  file-content:
    templates:
      "**/*.test.ts": "test-typescript"

overrides:
  # Exclude integration tests from AAA enforcement
  - files: ['tests/integration/**']
    rules:
      file-content: 0

  # Exclude test fixtures
  - files: ['**/__fixtures__/**']
    rules:
      file-content: 0
```

## Migration Strategy

### For Existing Codebases

1. **Start with lenient templates** (`test-go`, `test-typescript`, `test-python`)
2. **Enable for new tests only** initially
3. **Gradually refactor** existing tests
4. **Move to strict template** when ready

```yaml
# Phase 1: Only check new test files
rules:
  file-content:
    templates:
      "**/new_*_test.go": "test-go"

# Phase 2: Enable for all tests
rules:
  file-content:
    templates:
      "**/*_test.go": "test-go"

# Phase 3: Enforce strictly
rules:
  file-content:
    templates:
      "**/*_test.go": "test-strict-aaa"
```

## Framework-Specific Patterns

### Jest (JavaScript/TypeScript)

```typescript
describe('UserService', () => {
  describe('createUser', () => {
    it('should create user with valid data', () => {
      // Arrange
      const userData = { name: 'Alice', email: 'alice@example.com' };
      const service = new UserService();

      // Act
      const user = service.createUser(userData);

      // Assert
      expect(user.name).toBe('Alice');
      expect(user.email).toBe('alice@example.com');
    });
  });
});
```

### Pytest (Python)

```python
class TestUserService:
    def test_create_user_with_valid_data(self):
        # Arrange
        user_data = {'name': 'Alice', 'email': 'alice@example.com'}
        service = UserService()

        # Act
        user = service.create_user(user_data)

        # Assert
        assert user.name == 'Alice'
        assert user.email == 'alice@example.com'
```

### Go Testing with Testify

```go
func TestUserService_CreateUser_ValidData(t *testing.T) {
    // Arrange
    userData := UserData{Name: "Alice", Email: "alice@example.com"}
    service := NewUserService()

    // Act
    user, err := service.CreateUser(userData)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
    assert.Equal(t, "alice@example.com", user.Email)
}
```

## Related Documentation

- [Phase 3: Test Validation](TEST_VALIDATION.md) - Ensure tests exist
- [Phase 4: File Content Templates](FILE_CONTENT_TEMPLATES.md) - Template system overview
- [Examples](../examples/) - Complete configuration examples

## FAQ

### Q: Do integration tests need AAA comments?

A: It depends on your team. Integration tests often have more complex setup, so AAA comments can be especially helpful. However, you may choose to exempt them:

```yaml
overrides:
  - files: ['tests/integration/**']
    rules:
      file-content: 0
```

### Q: What about Given-When-Then pattern?

A: Given-When-Then (GWT) is equivalent to AAA:
- Given = Arrange
- When = Act
- Then = Assert

You can create a custom template for GWT if your team prefers that terminology.

### Q: Should setup/teardown functions follow AAA?

A: No. AAA is specifically for test functions. Setup/teardown functions are for common initialization and cleanup.

### Q: Can I use AAA with table-driven tests?

A: Yes! Table-driven tests can still use AAA:

```go
func TestCalculator_Add(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -2, -3, -5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            calc := NewCalculator()

            // Act
            got := calc.Add(tt.a, tt.b)

            // Assert
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Summary

AAA pattern enforcement through structurelint templates:

- ✅ Improves test readability and maintainability
- ✅ Enforces team consistency
- ✅ Integrates with test existence validation (Phase 3)
- ✅ Supports multiple languages and frameworks
- ✅ Flexible: lenient or strict enforcement
- ✅ Helps new developers learn best practices

Start with lenient templates and gradually move to stricter enforcement as your team adopts the pattern.
