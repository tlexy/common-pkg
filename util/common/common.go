package common

import (
	"fmt"
	"os"
)

func EnsureOutputDirectory(outputPath string) error {
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}
	return nil
}
