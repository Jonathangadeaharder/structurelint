package metrics

import "testing"

// This is a temporary test to demonstrate CI artifact upload
// It will intentionally fail to trigger artifact creation
func TestArtifactDemo_IntentionalFailure(t *testing.T) {
	// This intentional failure will trigger artifact upload in CI
	t.Error("🎯 DEMO: This intentional test failure will trigger automatic log artifact upload!")
	t.Error("Check the 'Artifacts' section at the bottom of the workflow run to download test.log")
	t.Error("The artifact will contain this complete output plus all other test results")
}
