package version

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func extractImagesFromChart(chartPath, chartName string) ([]ImageConfig, error) {
	data, err := os.ReadFile(chartPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Chart.yaml: %w", err)
	}

	var chart map[string]any
	if err := yaml.Unmarshal(data, &chart); err != nil {
		return nil, fmt.Errorf("failed to parse Chart.yaml: %w", err)
	}

	annotations, ok := chart["annotations"].(map[string]any)
	if !ok {
		return []ImageConfig{}, nil
	}

	annotationKey := fmt.Sprintf("%s/images", chartName)
	imagesYAML, ok := annotations[annotationKey].(string)
	if !ok || imagesYAML == "" {
		return []ImageConfig{}, nil
	}

	var images []ImageConfig
	if err := yaml.Unmarshal([]byte(imagesYAML), &images); err != nil {
		return nil, fmt.Errorf("failed to parse images annotation: %w", err)
	}

	return images, nil
}
