# Given-When-Then (GWT) Test Naming Pattern

## Overview

The **Given-When-Then (GWT)** naming pattern creates self-documenting test names that clearly express:
- **Given**: The preconditions and context (setup)
- **When**: The action being tested (behavior)
- **Then**: The expected outcome (assertion)

When combined with the AAA pattern for test structure, GWT naming creates highly readable and maintainable tests where **the test name tells you what** and **the test body shows you how**.

## Why GWT Naming?

### Benefits

✅ **Self-Documenting**: Test name describes exactly what is being tested
✅ **Searchable**: Easy to find tests for specific scenarios
✅ **Review-Friendly**: Reviewers understand intent from the name alone
✅ **Living Documentation**: Test names serve as behavioral specifications
✅ **Debugging**: Failure messages immediately show what scenario failed

### The Problem It Solves

**Bad test names:**
```go
func TestAdd(t *testing.T) { }                    // What about add?
func TestUserCreation(t *testing.T) { }           // What scenario?
func TestValidation(t *testing.T) { }             // Too vague
```

**Good GWT names:**
```go
func TestCalculator_GivenTwoPositiveNumbers_WhenAdding_ThenReturnsSum(t *testing.T) { }
func TestUser_GivenValidEmail_WhenCreating_ThenUserIsSaved(t *testing.T) { }
func TestValidator_GivenEmptyString_WhenValidating_ThenReturnsError(t *testing.T) { }
```

## Language-Specific Patterns

### Go

**Naming Convention**: `TestFunction_GivenContext_WhenAction_ThenOutcome`

```go
func TestCalculator_GivenTwoPositiveNumbers_WhenAdding_ThenReturnsSum(t *testing.T) {
    // Arrange
    calc := NewCalculator()
    a, b := 2, 3

    // Act
    result := calc.Add(a, b)

    // Assert
    assert.Equal(t, 5, result)
}
```

**Pattern Rules**:
- Start with `Test` + component/function name
- Use `_Given` to describe preconditions
- Use `_When` to describe action
- Use `_Then` to describe outcome
- Use PascalCase for each section

**More Examples**:
```go
func TestUserService_GivenInvalidEmail_WhenCreating_ThenReturnsError(t *testing.T)
func TestCache_GivenExpiredEntry_WhenRetrieving_ThenReturnsNil(t *testing.T)
func TestValidator_GivenNilInput_WhenValidating_ThenReturnsFalse(t *testing.T)
```

### TypeScript/JavaScript

**Naming Convention**: `"given X, when Y, then Z"`

```typescript
describe('Calculator', () => {
  it('given two positive numbers, when adding, then returns sum', () => {
    // Arrange
    const calc = new Calculator();
    const a = 2;
    const b = 3;

    // Act
    const result = calc.add(a, b);

    // Assert
    expect(result).toBe(5);
  });
});
```

**Pattern Rules**:
- Use lowercase sentence structure
- Separate clauses with commas
- Keep it concise and readable
- Natural language flow

**More Examples**:
```typescript
it('given an invalid email, when creating user, then throws validation error', () => {})
it('given expired token, when authenticating, then returns 401', () => {})
it('given empty cart, when checking out, then redirects to cart', () => {})
```

**Alternative Style** (more formal):
```typescript
it('Given two positive numbers, When adding, Then returns sum', () => {})
```

### Python

**Naming Convention**: `test_function_given_context_when_action_then_outcome`

```python
def test_calculator_given_two_positive_numbers_when_adding_then_returns_sum():
    # Arrange
    calc = Calculator()
    a = 2
    b = 3

    # Act
    result = calc.add(a, b)

    # Assert
    assert result == 5
```

**Pattern Rules**:
- Start with `test_`
- Use snake_case throughout
- Separate clauses with underscores
- Keep reasonably short (under 80 chars if possible)

**More Examples**:
```python
def test_user_service_given_invalid_email_when_creating_then_raises_error():
def test_cache_given_expired_entry_when_retrieving_then_returns_none():
def test_validator_given_nil_input_when_validating_then_returns_false():
```

## Configuration

### Enable GWT Templates

Add to your `.structurelint.yml`:

```yaml
rules:
  file-content:
    template-dir: ".structurelint/templates"
    templates:
      # Go tests with GWT naming
      "**/*_test.go": "test-gwt-go"

      # TypeScript/JavaScript tests with GWT naming
      "**/*.test.ts": "test-gwt-typescript"
      "**/*.test.js": "test-gwt-typescript"
      "**/*.spec.ts": "test-gwt-typescript"

      # Python tests with GWT naming
      "**/test_*.py": "test-gwt-python"
      "**/*_test.py": "test-gwt-python"
```

### Strict Enforcement

For maximum consistency, use the strict GWT template:

```yaml
rules:
  file-content:
    templates:
      "**/*_test.go": "test-gwt-strict"
      "**/*.test.ts": "test-gwt-strict"
      "**/test_*.py": "test-gwt-strict"
```

This enforces:
- ✅ GWT naming pattern
- ✅ AAA structure comments
- ✅ No debugging statements
- ✅ No TODO/FIXME markers
- ✅ No test skipping

## Best Practices

### 1. Be Specific in "Given"

**Good:**
```go
TestCalculator_GivenTwoNegativeNumbers_WhenAdding_ThenReturnsNegativeSum
TestUser_GivenExistingEmail_WhenRegistering_ThenReturnsConflictError
```

**Too Vague:**
```go
TestCalculator_GivenNumbers_WhenAdding_ThenReturnsResult
TestUser_GivenEmail_WhenRegistering_ThenReturnsError
```

### 2. Use Action Verbs in "When"

**Good:**
```typescript
it('given valid credentials, when authenticating, then returns access token', ...)
it('given empty cart, when calculating total, then returns zero', ...)
```

**Avoid Nouns:**
```typescript
it('given valid credentials, when authentication, then ...', ...)  // ❌
it('given empty cart, when total calculation, then ...', ...)       // ❌
```

### 3. State Observable Outcome in "Then"

**Good:**
```python
def test_validator_given_invalid_data_when_validating_then_returns_false():
def test_service_given_timeout_when_calling_then_raises_timeout_error():
```

**Too Generic:**
```python
def test_validator_given_invalid_data_when_validating_then_works():     # ❌
def test_service_given_timeout_when_calling_then_handles_error():       # ❌
```

### 4. Keep Names Reasonably Short

Aim for clarity over brevity, but avoid excessive length:

**Good Balance:**
```go
TestCache_GivenExpiredEntry_WhenRetrieving_ThenReturnsNil
```

**Too Long:**
```go
TestCache_GivenAnEntryThatWasAddedMoreThan24HoursAgo_WhenAttemptingToRetrieveIt_ThenReturnsNilBecauseItExpired
```

**Too Short:**
```go
TestCache_GivenExpired_WhenGet_ThenNil
```

### 5. Handle Edge Cases Clearly

```go
TestDivider_GivenZeroDivisor_WhenDividing_ThenReturnsDivideByZeroError
TestParser_GivenEmptyString_WhenParsing_ThenReturnsEmptyResult
TestFormatter_GivenNilInput_WhenFormatting_ThenReturnsEmptyString
```

## Table-Driven Tests with GWT

### Go with Subtests

```go
func TestCalculator_WhenAdding(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {
            name: "GivenTwoPositiveNumbers_ThenReturnsSum",
            a:    2,
            b:    3,
            want: 5,
        },
        {
            name: "GivenTwoNegativeNumbers_ThenReturnsNegativeSum",
            a:    -2,
            b:    -3,
            want: -5,
        },
        {
            name: "GivenPositiveAndNegative_ThenReturnsDifference",
            a:    5,
            b:    -3,
            want: 2,
        },
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

### TypeScript with Parametrized Tests

```typescript
describe('Calculator', () => {
  describe('when adding', () => {
    const testCases = [
      {
        desc: 'given two positive numbers, then returns sum',
        a: 2,
        b: 3,
        expected: 5,
      },
      {
        desc: 'given two negative numbers, then returns negative sum',
        a: -2,
        b: -3,
        expected: -5,
      },
    ];

    testCases.forEach(({ desc, a, b, expected }) => {
      it(desc, () => {
        // Arrange
        const calc = new Calculator();

        // Act
        const result = calc.add(a, b);

        // Assert
        expect(result).toBe(expected);
      });
    });
  });
});
```

### Python with pytest.parametrize

```python
@pytest.mark.parametrize("a,b,expected,description", [
    (2, 3, 5, "given_two_positive_numbers_then_returns_sum"),
    (-2, -3, -5, "given_two_negative_numbers_then_returns_negative_sum"),
    (5, -3, 2, "given_positive_and_negative_then_returns_difference"),
])
def test_calculator_when_adding(a, b, expected, description):
    # Arrange
    calc = Calculator()

    # Act
    result = calc.add(a, b)

    # Assert
    assert result == expected, description
```

## Migration Strategy

### Phase 1: Start with New Tests

Only enforce GWT for new test files:

```yaml
rules:
  file-content:
    templates:
      "**/new_*_test.go": "test-gwt-go"
```

### Phase 2: Enable for All Tests (Lenient)

Use lenient templates that encourage but don't strictly require GWT:

```yaml
rules:
  file-content:
    templates:
      "**/*_test.go": "test-go"  # Standard AAA, no GWT requirement
```

### Phase 3: Strict GWT Enforcement

Move to strict GWT enforcement:

```yaml
rules:
  file-content:
    templates:
      "**/*_test.go": "test-gwt-go"  # Requires GWT naming
```

### Phase 4: Ultra-Strict for Critical Code

```yaml
rules:
  file-content:
    templates:
      "**/*_test.go": "test-gwt-go"

overrides:
  # Strict enforcement for business logic
  - files: ['internal/core/**/*_test.go']
    rules:
      file-content:
        templates:
          "**/*_test.go": "test-gwt-strict"
```

## Common Violations

### Missing GWT Pattern

```
internal/calculator_test.go: missing required pattern "TestFunction_GivenContext_WhenAction_ThenOutcome" (template: test-gwt-go)
```

**Fix**: Rename test to follow GWT convention:
```go
// Before
func TestAdd(t *testing.T) { }

// After
func TestCalculator_GivenTwoNumbers_WhenAdding_ThenReturnsSum(t *testing.T) { }
```

### Missing AAA Comments

```
src/calculator.test.ts: missing required pattern "// Arrange" (template: test-gwt-typescript)
```

**Fix**: Add AAA structure comments:
```typescript
it('given two numbers, when adding, then returns sum', () => {
  // Arrange
  const calc = new Calculator();

  // Act
  const result = calc.add(2, 3);

  // Assert
  expect(result).toBe(5);
});
```

## Exemptions

### Integration Tests

Integration tests often have complex setup and may not fit GWT naming:

```yaml
overrides:
  - files: ['tests/integration/**']
    rules:
      file-content: 0
```

### Test Utilities

```yaml
overrides:
  - files: ['**/__fixtures__/**', '**/__mocks__/**', '**/testdata/**']
    rules:
      file-content: 0
```

## Comparison: AAA vs GWT Templates

| Template | AAA Structure | GWT Naming | Use Case |
|----------|---------------|------------|----------|
| `test-go.yml` | ✅ Encouraged | ❌ Not required | General Go tests |
| `test-gwt-go.yml` | ✅ Required | ✅ Required | Descriptive Go tests |
| `test-strict-aaa.yml` | ✅ Required | ❌ Not required | Strict AAA only |
| `test-gwt-strict.yml` | ✅ Required | ✅ Required | Maximum quality |

## Integration with Phase 3

Combine GWT naming with test existence validation:

```yaml
rules:
  # Phase 3: Ensure tests exist
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "**/*.go"
    exemptions:
      - "cmd/**/*.go"

  # Phase 4: Ensure tests follow GWT naming
  file-content:
    templates:
      "**/*_test.go": "test-gwt-go"
```

This ensures:
1. Every source file has tests (Phase 3)
2. Every test follows GWT naming + AAA structure (Phase 4)

## Real-World Examples

### Authentication Service

```go
func TestAuthService_GivenValidCredentials_WhenAuthenticating_ThenReturnsAccessToken(t *testing.T)
func TestAuthService_GivenInvalidPassword_WhenAuthenticating_ThenReturnsUnauthorizedError(t *testing.T)
func TestAuthService_GivenExpiredToken_WhenRefreshing_ThenReturnsNewToken(t *testing.T)
func TestAuthService_GivenRevokedToken_WhenValidating_ThenReturnsFalse(t *testing.T)
```

### Shopping Cart

```typescript
it('given empty cart, when calculating total, then returns zero', ...)
it('given items in cart, when applying discount, then reduces total', ...)
it('given out of stock item, when adding to cart, then shows error', ...)
it('given guest user, when checking out, then prompts for login', ...)
```

### Data Validator

```python
def test_validator_given_valid_email_when_validating_then_returns_true():
def test_validator_given_missing_at_symbol_when_validating_then_returns_false():
def test_validator_given_empty_string_when_validating_then_raises_value_error():
def test_validator_given_none_input_when_validating_then_raises_type_error():
```

## FAQ

### Q: Is GWT naming redundant with AAA comments?

A: No, they serve different purposes:
- **GWT naming**: Describes WHAT is being tested (the scenario)
- **AAA comments**: Shows HOW the test is structured (the implementation)

Together they provide complete understanding at both high and low levels.

### Q: What if test name becomes too long?

A: Prioritize clarity. If the name exceeds ~80 characters, consider:
1. Abbreviating common terms (HTTP, URL, ID)
2. Removing redundant words
3. Splitting into multiple focused tests

### Q: Can I use "Should" instead of "Then"?

A: The templates require GWT, but you can create custom templates:

```yaml
# .structurelint/templates/test-should.yml
required-patterns:
  - "Test\\w+_Given\\w+_When\\w+_Should\\w+"
```

### Q: What about BDD frameworks like Cucumber?

A: GWT is the foundation of BDD! These templates apply the same principles to unit tests. For Cucumber/Gherkin, you already have GWT built into the framework.

## Related Documentation

- [AAA Pattern](TEST_AAA_PATTERN.md) - Test structure documentation
- [Phase 3: Test Validation](TEST_VALIDATION.md) - Ensure tests exist
- [Phase 4: File Content Templates](FILE_CONTENT_TEMPLATES.md) - Template system overview

## Summary

GWT naming + AAA structure = **Highly Readable Tests**

- ✅ **Test names** describe the scenario (Given-When-Then)
- ✅ **Test structure** is consistent (Arrange-Act-Assert)
- ✅ **Teams** write tests the same way
- ✅ **Reviews** are faster and more thorough
- ✅ **Documentation** is built into test names

Start with lenient templates and gradually adopt strict GWT naming as your team builds the habit. The investment in descriptive test names pays dividends in code comprehension and maintenance.
