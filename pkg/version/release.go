package version

import (
	"encoding/json"
	"io"
	"net/http"
)

type Release struct {
	TagName string `json:"tag_name"`
}

func GetRelease() (string, error) {
	resp, err := http.Get("https://api.github.com/repos/gardener/landscaper/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var release Release
	err = json.Unmarshal(body, &release)
	if err != nil {
		return "", err
	}

	return release.TagName, err
}
