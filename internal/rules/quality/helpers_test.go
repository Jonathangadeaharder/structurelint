package quality

import "os"

// writeFile writes content to a file, for use in tests.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
