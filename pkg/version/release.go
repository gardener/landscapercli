package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func GetRelease() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/gardener/landscaper/releases/latest")
	if err != nil {
		return "", fmt.Errorf("failed to get latest landscaper release info: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to get latest landscaper release info: read failed: %w", err)
	}

	var release Release
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", fmt.Errorf("failed to get latest landscaper release info: unmarshal failed: %w", err)
	}

	if release.TagName == "" {
		return "", fmt.Errorf("failed to get latest landscaper release info: no tag")
	}

	return release.TagName, err
}
