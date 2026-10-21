# Integration Tests

## Overview

This directory contains integration tests that verify the security modules work together correctly.

## Test Files

### `security_integration_test.go`
Tests the integration between:
- `exec.SafeExecutor` - Safe command execution with whitelisting
- `path.IsPathSafe` - Path traversal prevention
- Command injection prevention

Test scenarios:
1. **SafeCommandExecutionWithinAllowedPaths** - Verifies whitelisted commands work
2. **PathTraversalPrevention** - Blocks `../../../etc/passwd` style attacks
3. **CommandInjectionPrevention** - Blocks shell metacharacters (`;`, `|`, `&`, `` ` ``, etc.)
4. **CombinedSafetyChecks** - End-to-end security validation

### `validation_integration_test.go`
Tests the integration of all validation functions:
- `validation.ValidateVersion` - Semantic versioning format
- `validation.ValidatePath` - Path safety and traversal prevention
- `validation.ValidateURL` - HTTPS enforcement and SSRF protection

Test scenarios:
1. **VersionValidationIntegration** - Valid/invalid version formats
2. **PathValidationIntegration** - Safe paths, URL encoding, null bytes
3. **URLValidationIntegration** - HTTPS only, no localhost/private IPs
4. **CombinedValidationScenarios** - Real-world attack scenarios

## Running Tests

```bash
# Run all integration tests
go test ./test/integration/... -v

# Run integration tests with race detection
go test ./test/integration/... -v -race

# Run all tests (unit + integration)
go test ./... -v -race

# Skip integration tests in short mode
go test ./... -v -short
```

## TDD Methodology

These tests were written following Test-Driven Development:
1. ✅ **Red Phase** - Tests written first (would fail on initial run)
2. ✅ **Green Phase** - Implementation exists (tests should pass)
3. ✅ Tests verify security modules work together correctly
4. ✅ All tests skip in short mode for CI/CD efficiency

## Coverage

The integration tests provide coverage for:
- Command injection prevention
- Path traversal attacks
- SSRF (Server-Side Request Forgery) prevention
- Input validation across multiple layers
- Combined security scenarios

## Notes

- Tests create temporary directories for isolation
- All temporary directories are cleaned up after tests
- Tests use `github.com/stretchr/testify/assert` for clear assertions
- Integration tests take longer than unit tests and can be skipped with `-short`
