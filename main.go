package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

//go:embed shell.html
var shellHTML embed.FS

var (
	port      = flag.Int("port", 8081, "Port to listen on")
	assetsDir = flag.String("assets", "", "Path to assets/ folder (default: ./assets next to binary)")
	noOpen    = flag.Bool("no-open", false, "Don't open browser automatically")
)

func main() {
	flag.Parse()

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	if *assetsDir == "" {
		*assetsDir = filepath.Join(exeDir, "assets")
	}

	// Verify WASM and JS files exist
	for _, f := range []string{"ClassiCube.wasm", "ClassiCube.js"} {
		if _, err := os.Stat(filepath.Join(*assetsDir, f)); err != nil {
			log.Fatalf("Missing %s in %s\nRun build.sh first.", f, *assetsDir)
		}
	}

	mux := http.NewServeMux()

	// Serve index (our custom shell) at /
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			// Serve embedded shell.html
			data, _ := shellHTML.ReadFile("shell.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// Required headers for SharedArrayBuffer (WASM threads)
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			w.Write(data)
			return
		}

		// Try assets dir first
		localPath := filepath.Join(*assetsDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(localPath); err == nil && !info.IsDir() {
			// Set COOP/COEP for all responses (needed for WASM)
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
			// Correct MIME for WASM
			if filepath.Ext(localPath) == ".wasm" {
				w.Header().Set("Content-Type", "application/wasm")
			}
			http.ServeFile(w, r, localPath)
			return
		}

		http.NotFound(w, r)
	})

	// Health check
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%d", *port)
	url := fmt.Sprintf("http://localhost:%d", *port)

	log.Printf("=== ClassiCube Launcher ===")
	log.Printf("Assets: %s", *assetsDir)
	log.Printf("URL:    %s", url)

	srv := &http.Server{Addr: addr, Handler: mux}

	if !*noOpen {
		go func() {
			waitForPort(*port, 5*time.Second)
			openBrowser(url)
		}()
	}

	log.Printf("Serving at %s  (Ctrl+C to quit)", url)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func waitForPort(port int, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
		if err == nil {
			c.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

// Implement fs.FS for the assets directory at runtime (for future use)
type dirFS struct{ root string }
func (d dirFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.Join(d.root, name))
}
