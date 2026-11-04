package handler

import "testing"

// Memastikan SEMUA template valid (sintaks + define "content"). Dijalankan di CI.
func TestTemplatesParse(t *testing.T) {
	if _, err := NewRenderer("../../templates", false); err != nil {
		t.Fatalf("template tidak valid: %v", err)
	}
}
