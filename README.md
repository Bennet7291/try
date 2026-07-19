# ClassiCube Launcher (Go Binary)

Serves the ClassiCube WASM client on localhost. No installation required beyond the binary.

## Structure

```
classicube-launcher      ← the Go binary (run this)
assets/
  ClassiCube.wasm        ← compiled game (2.6 MB)
  ClassiCube.js          ← emscripten glue code
```

The HTML shell is embedded directly inside the binary — no extra HTML file needed.

## Usage

```bash
./classicube-launcher                  # serves on http://localhost:8081
./classicube-launcher --port 9000      # custom port
./classicube-launcher --no-open        # don't auto-open browser
./classicube-launcher --assets ./path  # custom assets dir
```

## Game Notes

- **Player data** (settings, maps, skins) is stored in the browser's IndexedDB — persists between sessions
- **Multiplayer**: connect to any classic Minecraft server using the in-game menu
  - Works with public ClassiCube servers (search in-game → "Play online")
  - Compatible with original Minecraft Classic servers
- **Singleplayer**: available from the main menu

## Build from source

Requirements: `clang-18`, `lld-18`, `llvm-18`, `binaryen`, `emscripten 3.1.x`

```bash
# Set up emscripten (from source, no emsdk needed)
export PATH="/path/to/emscripten-3.1.74:$PATH"
cd ClassiCube-master
make -f misc/makefiles/web.mk RELEASE=1 -j$(nproc)
# Output: ClassiCube.html, ClassiCube.js, ClassiCube.wasm
cp ClassiCube.js ClassiCube.wasm /path/to/classicube-launcher/assets/
```

## Cross-compile the launcher

```bash
GOOS=windows GOARCH=amd64  go build -o classicube-launcher.exe .
GOOS=darwin  GOARCH=arm64  go build -o classicube-launcher-mac .
GOOS=linux   GOARCH=amd64  go build -o classicube-launcher-linux .
```
