package github

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type fileContentRes struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type Client struct {
	Token      string
	Owner      string
	Repo       string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(token, owner, repo string) *Client {
	return &Client{
		Token:   token,
		Owner:   owner,
		Repo:    repo,
		BaseURL: "https://api.github.com",
		HTTPClient: &http.Client{
			Timeout: time.Second * 15,
		},
	}
}

func (c *Client) ListFleetings(directoryPath string) ([]string, error) {
	escapedPath := escapeGitPath(directoryPath)
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, escapedPath)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creando request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error ejecutando request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error inesperado en la respuesta del servidor: %s", res.Status)
	}

	// Mapear respuesta de GitHub
	var items []ContentItem
	if err := json.NewDecoder(res.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("Error decodificando respuesta: %w", err)
	}

	var mds []string
	for _, item := range items {
		if item.Type == TypeFile && strings.HasSuffix(item.Name, ".md") {
			mds = append(mds, item.Path)
		}
	}
	return mds, nil
}

func (c *Client) GetFileContent(path string) (string, error) {
	escapedPath := escapeGitPath(path)
	endpoint := fmt.Sprintf("%s/repos/%s/%s/contents/%s", c.BaseURL, c.Owner, c.Repo, escapedPath)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("Error creando request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", c.Token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error ejecutando request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Error inesperado en la respuesta del servidor: %s", res.Status)
	}

	var contentRes fileContentRes
	if err := json.NewDecoder(res.Body).Decode(&contentRes); err != nil {
		return "", fmt.Errorf("Error decodificando respuesta: %w", err)
	}

	// Github devuelve el contenido codificado en base64, hay que decodificar y formatear
	cleanBase64 := strings.ReplaceAll(contentRes.Content, "/n", "")
	decodedBytes, err := base64.StdEncoding.DecodeString(cleanBase64)
	if err != nil {
		return "", fmt.Errorf("Error decodificando contenido: %w", err)
	}

	return string(decodedBytes), nil
}

func escapeGitPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}
