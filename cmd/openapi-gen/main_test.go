package main

import "testing"

func TestBuildSpecOperationIDsAreUnique(t *testing.T) {
	spec := buildSpec("https://example.com")
	seen := map[string]string{}
	for path, methods := range spec.Paths {
		for method, operation := range methods {
			if operation.OperationID == "" {
				t.Fatalf("%s %s has empty operationId", method, path)
			}
			if previous, ok := seen[operation.OperationID]; ok {
				t.Fatalf("operationId %q used by both %s and %s %s", operation.OperationID, previous, method, path)
			}
			seen[operation.OperationID] = method + " " + path
		}
	}
}
