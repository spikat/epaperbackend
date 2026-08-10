package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jonathanribas/epaperbackend/pkg/config"
	"github.com/jonathanribas/epaperbackend/pkg/httpx"
	"github.com/jonathanribas/epaperbackend/pkg/registry"
	"github.com/jonathanribas/epaperbackend/server/debug/preview"
)

type Server struct {
	cfg    config.Config
	client *http.Client
}

func New(cfg config.Config) *Server {
	return &Server{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /static/debug.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write([]byte(debugCSS))
	})
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /debug/{service}", s.handleServicePage)
	mux.HandleFunc("GET /api/services", s.handleServices)
	mux.HandleFunc("POST /api/proxy", s.handleProxy)
	mux.HandleFunc("GET /api/plugin/{service}/files", s.handlePluginFiles)
	mux.HandleFunc("GET /api/plugin/{service}/file", s.handlePluginFileGet)
	mux.HandleFunc("PUT /api/plugin/{service}/file", s.handlePluginFilePut)
	mux.HandleFunc("POST /api/preview/{service}", s.handlePreview)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	services := registry.Infos()
	health := registry.Health(r.Context())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html><html><head><meta charset="utf-8"><title>epaperbackend debug</title>
<link rel="stylesheet" href="/static/debug.css"></head><body>
<h1>epaperbackend debug</h1>
<table><thead><tr><th>Service</th><th>Route</th><th>Status</th><th></th></tr></thead><tbody>`)
	for _, svc := range services {
		st := health[svc.Name]
		status := "ok"
		if !st.OK {
			status = "error"
		}
		_, _ = fmt.Fprintf(w, `<tr><td>%s</td><td><code>%s</code></td><td class="status-%s">%s</td>
<td><a href="/debug/%s">open</a></td></tr>`, svc.Name, svc.RoutePrefix, status, status, svc.Name)
	}
	_, _ = fmt.Fprint(w, `</tbody></table></body></html>`)
}

func (s *Server) handleServicePage(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("service")
	svc, ok := registry.Get(name)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	page := fmt.Sprintf(servicePageHTML,
		name,
		name,
		svc.RoutePrefix(),
		s.cfg.DebugPreviewWidth,
		s.cfg.DebugPreviewHeight,
		name,
		svc.RoutePrefix(),
		s.cfg.MainAPIBaseURL,
	)
	_, _ = w.Write([]byte(page))
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"services": registry.Infos(),
		"health":   registry.Health(r.Context()),
	})
}

type proxyRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type proxyResponse struct {
	Request  proxyRequest `json:"request"`
	Status   int          `json:"status"`
	Headers  http.Header  `json:"headers"`
	Body     string       `json:"body"`
	Duration string       `json:"duration"`
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	var req proxyRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Method == "" {
		req.Method = http.MethodGet
	}
	if req.URL == "" {
		httpx.WriteError(w, http.StatusBadRequest, "url is required")
		return
	}

	target, err := url.Parse(req.URL)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var body io.Reader
	if req.Body != "" {
		body = strings.NewReader(req.Body)
	}

	start := time.Now()
	upReq, err := http.NewRequestWithContext(r.Context(), req.Method, target.String(), body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	for k, v := range req.Headers {
		upReq.Header.Set(k, v)
	}

	resp, err := s.client.Do(upReq)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, proxyResponse{
		Request:  req,
		Status:   resp.StatusCode,
		Headers:  resp.Header,
		Body:     string(respBody),
		Duration: time.Since(start).String(),
	})
}

func (s *Server) handlePluginFiles(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("service")
	svc, ok := registry.Get(name)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "service not found")
		return
	}
	files, err := preview.ListPluginFiles(svc.PluginDir())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"files": files})
}

func (s *Server) handlePluginFileGet(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("service")
	path := r.URL.Query().Get("path")
	svc, ok := registry.Get(name)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "service not found")
		return
	}
	content, err := preview.ReadPluginFile(svc.PluginDir(), path)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"path": path, "content": content})
}

func (s *Server) handlePluginFilePut(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("service")
	svc, ok := registry.Get(name)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "service not found")
		return
	}
	var body struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := preview.WritePluginFile(svc.PluginDir(), body.Path, body.Content); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

type previewRequest struct {
	Data   json.RawMessage   `json:"data"`
	Size   string            `json:"size"`
	Width  int               `json:"width"`
	Height int               `json:"height"`
	Config map[string]string `json:"config"`
}

func (s *Server) handlePreview(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("service")
	svc, ok := registry.Get(name)
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "service not found")
		return
	}

	var req previewRequest
	if err := httpx.DecodeJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Size == "" {
		req.Size = "full"
	}
	if req.Width <= 0 {
		req.Width = s.cfg.DebugPreviewWidth
	}
	if req.Height <= 0 {
		req.Height = s.cfg.DebugPreviewHeight
	}

	var data map[string]any
	if len(req.Data) > 0 {
		if err := json.Unmarshal(req.Data, &data); err != nil {
			httpx.WriteError(w, http.StatusBadRequest, "invalid data json")
			return
		}
	}

	cfgAny := map[string]any{
		"backend_url": s.cfg.MainAPIBaseURL,
	}
	for k, v := range req.Config {
		cfgAny[k] = v
	}

	html, err := preview.RenderPlugin(svc.PluginDir(), req.Size, preview.Bindings{
		Data:   data,
		Size:   req.Size,
		Config: cfgAny,
		Width:  req.Width,
		Height: req.Height,
	})
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	vw, vh := preview.ViewportSize(req.Size, req.Width, req.Height)
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"html":   html,
		"width":  vw,
		"height": vh,
	})
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.DebugPort),
		Handler: s.Handler(),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}

var debugCSS = `
body { font-family: system-ui, sans-serif; margin: 24px; color: #111; max-width: 960px; }
table { border-collapse: collapse; width: 100%; }
th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
.status-ok { color: green; }
.status-error { color: red; }
.layout { display: flex; flex-direction: column; gap: 16px; }
label { display:block; margin-bottom:8px; }
textarea, input, select { width: 100%; box-sizing: border-box; }
textarea { min-height: 180px; font-family: monospace; }
.panel { border: 1px solid #ddd; padding: 12px; border-radius: 8px; max-width: 100%; overflow: hidden; }
.preview-frame { border: 1px solid #999; background: #eee; overflow: auto; margin-top:8px; max-width: 100%; }
button { margin-right: 8px; margin-top: 8px; }
pre.output-box {
  margin: 0;
  padding: 10px;
  background: #f6f6f6;
  border: 1px solid #ddd;
  border-radius: 4px;
  max-width: 100%;
  max-height: 360px;
  overflow-x: hidden;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
  overflow-wrap: anywhere;
  font-family: ui-monospace, monospace;
  font-size: 12px;
  line-height: 1.4;
}
`

const servicePageHTML = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8"/>
<title>Debug %s</title>
<link rel="stylesheet" href="/static/debug.css"/>
</head>
<body>
<h1>Debug: %s</h1>
<p>Route: <code>%s</code></p>
<div class="layout">
  <div class="panel">
    <h2>API call</h2>
    <label>Country <input id="country" value="FR"/></label>
    <label>City / postal <input id="city" value="Marseille"/></label>
    <label>Width <input id="width" type="number" value="%d"/></label>
    <label>Height <input id="height" type="number" value="%d"/></label>
    <button id="callBtn">Call service</button>
    <h3>Request</h3>
    <pre id="requestOut" class="output-box"></pre>
    <h3>Response</h3>
    <pre id="responseOut" class="output-box"></pre>
  </div>
  <div class="panel">
    <h2>Plugin</h2>
    <label>File <select id="pluginFile"></select></label>
    <button id="reloadPluginBtn">Reload plugin</button>
    <textarea id="pluginContent"></textarea>
    <label>Size
      <select id="previewSize">
        <option value="full">full</option>
        <option value="half_vertical">half_vertical</option>
        <option value="quadrant">quadrant</option>
      </select>
    </label>
    <button id="previewBtn">Generate preview</button>
    <div class="preview-frame"><iframe id="previewFrame" title="preview"></iframe></div>
  </div>
</div>
<script>
const serviceName = %q;
const routePrefix = %q;
const mainAPIBase = %q;
const storageKey = 'epaperbackend.debug.resolution';

function loadResolution() {
  try {
    const saved = JSON.parse(localStorage.getItem(storageKey) || '{}');
    if (saved.width) document.getElementById('width').value = saved.width;
    if (saved.height) document.getElementById('height').value = saved.height;
  } catch (_) {}
}
function saveResolution() {
  localStorage.setItem(storageKey, JSON.stringify({
    width: Number(document.getElementById('width').value),
    height: Number(document.getElementById('height').value)
  }));
}

async function loadPluginFiles() {
  const res = await fetch('/api/plugin/' + serviceName + '/files');
  const data = await res.json();
  const sel = document.getElementById('pluginFile');
  sel.innerHTML = '';
  (data.files || []).forEach(f => {
    const opt = document.createElement('option');
    opt.value = f; opt.textContent = f; sel.appendChild(opt);
  });
  if (sel.value) await loadPluginFile();
}

async function loadPluginFile() {
  const path = document.getElementById('pluginFile').value;
  const res = await fetch('/api/plugin/' + serviceName + '/file?path=' + encodeURIComponent(path));
  const data = await res.json();
  document.getElementById('pluginContent').value = data.content || '';
}

document.getElementById('reloadPluginBtn').onclick = loadPluginFile;
document.getElementById('pluginFile').onchange = loadPluginFile;

document.getElementById('callBtn').onclick = async () => {
  saveResolution();
  const country = document.getElementById('country').value;
  const city = document.getElementById('city').value;
  const url = mainAPIBase + routePrefix + '?country=' + encodeURIComponent(country) + '&city=' + encodeURIComponent(city);
  document.getElementById('requestOut').textContent = 'GET ' + url;
  const res = await fetch('/api/proxy', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({method:'GET', url})
  });
  const out = await res.json();
  document.getElementById('responseOut').textContent = JSON.stringify(out, null, 2);
  window.lastResponseBody = out.body;
};

document.getElementById('previewBtn').onclick = async () => {
  saveResolution();
  let data = {};
  try { data = JSON.parse(window.lastResponseBody || '{}'); } catch (_) {}
  const width = Number(document.getElementById('width').value) || 800;
  const height = Number(document.getElementById('height').value) || 480;
  const size = document.getElementById('previewSize').value;
  const res = await fetch('/api/preview/' + serviceName, {
    method: 'POST',
    headers: {'Content-Type':'application/json'},
    body: JSON.stringify({data, size, width, height, config: {backend_url: mainAPIBase, country: document.getElementById('country').value, city: document.getElementById('city').value}})
  });
  const out = await res.json();
  const frame = document.getElementById('previewFrame');
  frame.width = out.width;
  frame.height = out.height;
  frame.srcdoc = out.html;
};

loadResolution();
loadPluginFiles();
</script>
</body>
</html>`
