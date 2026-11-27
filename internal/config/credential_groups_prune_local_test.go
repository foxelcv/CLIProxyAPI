package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Removing a downstream key (or its mapping) must remove it from the saved
// file; otherwise the next reload resurrects the stale restriction.
func TestV8AccessAPIKeyCredentialGroupsSavePrunesRemovedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	src := "config-version: 8\naccess:\n    api-keys:\n        - \"k1\"\n        - \"k2\"\n    api-key-credential-groups:\n        k1:\n            - mine\n        k2:\n            - solo\n"
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.APIKeys = []string{"k1"}
	delete(cfg.APIKeyCredentialGroups, "k2")
	if err = SaveConfigPreserveComments(path, cfg); err != nil {
		t.Fatal(err)
	}
	again, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string][]string{"k1": {"mine"}}; !reflect.DeepEqual(again.APIKeyCredentialGroups, want) {
		raw, _ := os.ReadFile(path)
		t.Fatalf("reloaded groups = %#v, want %#v\n%s", again.APIKeyCredentialGroups, want, raw)
	}

	again.APIKeyCredentialGroups = nil
	if err = SaveConfigPreserveComments(path, again); err != nil {
		t.Fatal(err)
	}
	cleared, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.APIKeyCredentialGroups) != 0 {
		raw, _ := os.ReadFile(path)
		t.Fatalf("cleared groups resurrected: %#v\n%s", cleared.APIKeyCredentialGroups, raw)
	}
}

// An explicitly empty group list is a fail-closed restriction and must survive a save.
func TestV8AccessAPIKeyCredentialGroupsSaveKeepsEmptyRestriction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	src := "config-version: 8\naccess:\n    api-keys:\n        - \"k1\"\n    api-key-credential-groups:\n        k1:\n            - mine\n"
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg.APIKeyCredentialGroups["k1"] = []string{}
	if err = SaveConfigPreserveComments(path, cfg); err != nil {
		t.Fatal(err)
	}
	again, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	groups, ok := again.APIKeyCredentialGroups["k1"]
	if !ok || len(groups) != 0 {
		raw, _ := os.ReadFile(path)
		t.Fatalf("empty restriction not preserved: present=%v groups=%#v\n%s", ok, groups, raw)
	}
}
