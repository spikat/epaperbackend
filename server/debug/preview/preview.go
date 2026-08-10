package preview

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/osteele/liquid"
)

type Bindings struct {
	Data   map[string]any
	Size   string
	Config map[string]any
	Width  int
	Height int
}

func ViewportSize(size string, width, height int) (int, int) {
	switch size {
	case "half_vertical", "half_horizontal":
		return width / 2, height
	case "quadrant":
		return width / 2, height / 2
	default:
		return width, height
	}
}

func RenderPlugin(pluginDir, size string, bindings Bindings) (string, error) {
	vw, vh := ViewportSize(size, bindings.Width, bindings.Height)
	bindings.Width = vw
	bindings.Height = vh

	shared, _ := readTemplate(pluginDir, "shared.liquid")
	body, err := readTemplate(pluginDir, templateFileForSize(size))
	if err != nil {
		return "", err
	}

	markup := shared + "\n" + wrapView(body)
	engine := liquid.NewEngine()
	vars := map[string]interface{}{
		"data":   bindings.Data,
		"size":   size,
		"config": bindings.Config,
		"trmnl": map[string]interface{}{
			"width":  bindings.Width,
			"height": bindings.Height,
		},
	}

	rendered, err := engine.ParseAndRenderString(markup, vars)
	if err != nil {
		return "", fmt.Errorf("liquid render: %w", err)
	}

	page := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8"/>
<style>
html, body { margin:0; padding:0; background:#ddd; }
body { width:%dpx; height:%dpx; overflow:hidden; background:#fff; color:#111; font-family: sans-serif; }
.view { width:100%%; height:100%%; box-sizing:border-box; padding:6px; display:flex; flex-direction:column; overflow:hidden; }
img, svg { max-width:100%%; }
.weather-icon { width:56px; height:56px; }
.chart { display:block; width:100%%; flex:0 0 auto; margin:2px 0 4px; overflow:hidden; line-height:0; }
.chart svg { display:block; width:100%%; height:auto; }
.daily-row { display:flex; gap:4px; margin-top:auto; flex-shrink:0; }
.daily-item { flex:1; min-width:0; text-align:center; border:1px solid #ccc; padding:4px 2px; box-sizing:border-box; font-size:12px; }
.title { font-size:18px; font-weight:bold; }
.label { font-size:13px; color:#666; }
.value { font-size:34px; font-weight:bold; }
</style>
</head>
<body><div class="preview-root">%s</div></body>
</html>`, bindings.Width, bindings.Height, rendered)

	return page, nil
}

func templateFileForSize(size string) string {
	switch size {
	case "half_vertical":
		return "half_vertical.liquid"
	case "half_horizontal":
		return "half_horizontal.liquid"
	case "quadrant":
		return "quadrant.liquid"
	default:
		return "full.liquid"
	}
}

func wrapView(body string) string {
	body = strings.TrimSpace(body)
	if strings.Contains(body, "view--{{ size }}") || strings.Contains(body, "view--{{size}}") {
		return body
	}
	return `<div class="view view--{{ size }}">` + "\n" + body + "\n" + `</div>`
}

func readTemplate(pluginDir, filename string) (string, error) {
	path := filepath.Join(pluginDir, filename)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && filename != "full.liquid" {
			return readTemplate(pluginDir, "full.liquid")
		}
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

func ListPluginFiles(pluginDir string) ([]string, error) {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".liquid") || name == "settings.yml" {
			files = append(files, name)
		}
	}
	return files, nil
}

func ReadPluginFile(pluginDir, name string) (string, error) {
	clean := filepath.Clean(name)
	if clean != name || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid path")
	}
	path := filepath.Join(pluginDir, clean)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func WritePluginFile(pluginDir, name, content string) error {
	clean := filepath.Clean(name)
	if clean != name || strings.Contains(clean, "..") {
		return fmt.Errorf("invalid path")
	}
	path := filepath.Join(pluginDir, clean)
	return os.WriteFile(path, []byte(content), 0o644)
}
