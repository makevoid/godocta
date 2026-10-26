<div align="center">

<img alt="The Octa welcome screen: Small blocks. Infinite worlds." src="https://github.com/user-attachments/assets/4bbbf557-c19a-4cba-822e-2ee44bce2dc2" width="100%" />

# Octa · Structure Studio

**Small blocks. Infinite worlds.**<br>
Build, paint and explore structures made of truncated octahedrons, the 14-faced cells that tile space.

[![MIT license](https://img.shields.io/badge/license-MIT-2bd4c4?style=flat-square)](LICENSE)
[![Godot 4.4+](https://img.shields.io/badge/Godot-4.4%2B-478cbf?style=flat-square&logo=godotengine&logoColor=white)](https://godotengine.org)
[![Go 1.22+](https://img.shields.io/badge/Go-1.22%2B-00add8?style=flat-square&logo=go&logoColor=white)](gen/README.md)
[![85 builds](https://img.shields.io/badge/builds-85-e0a040?style=flat-square)](docs/builds.md)

[Quick start](#quick-start) · [Using the studio](#using-the-studio) · [Gallery](#gallery) · [Docs](#documentation)

</div>

## Take the tour

https://github.com/user-attachments/assets/954155b8-4db1-4b37-abe7-11697d8fbf5e

## Highlights

- 🏰 **85 builds**: landmarks, old towns, ships, machines and fantasy worlds, mostly 80,000 to 130,000 cells each. Every build assembles from the bottom up in its own sky, ground or water.
- 🖌️ **Brush in** new cells in any palette color, or swallow them with a **black hole**.
- ↩️ **Undo and redo** the last ten strokes.
- 🔍 **X-ray** view to see inside any build.
- 💾 A compact **`.octa`** file format: 12 bytes per cell, plus one byte for its color.
- 📦 **Offline and dependency-free**: no account, no API key, no addons. A standalone Go toolkit generates the builds.

## Quick start

Install [Godot 4.4 or newer](https://godotengine.org/download) (tested with 4.7.2), then:

```sh
git clone https://github.com/makevoid/godocta.git
cd godocta
bin/run            # or: godot --path .
```

You can also open `project.godot` in Godot and press **F5**. On macOS, double-click **Launch Octa.command**. On Windows, run `godot --path .` from a terminal.

The builds are already in the repository. You only need Go 1.22+ to [regenerate them](docs/builds.md#regenerating-builds).

## Using the studio

Press **Enter** on the welcome screen to start. The top bar holds the menus and file actions. The toolkit is on the left, and the selected tool's settings appear at the bottom right.

### 1. Start from a primitive

Open **Primitives** (or press **1–4**) for a cube, sphere, pyramid or cylinder. Every arrangement assembles cell by cell. Press **R** to replay it.

<img alt="The Primitives menu over a cube" src="https://github.com/user-attachments/assets/b732f233-436d-46c0-8d64-945ab27acec5" width="100%" />

### 2. Brush in and black hole

| **Brush in** · `B` | **Black hole** · `E` |
| :---: | :---: |
| <img alt="Brush in adding cells to a sphere" src="https://github.com/user-attachments/assets/3b94af25-6eac-43e2-b8bc-0848df9b194c" /> | <img alt="A black hole carving a sphere" src="https://github.com/user-attachments/assets/6b9fc59b-1c27-46d6-812b-38b8c8bb8593" /> |
| New cells fly in and snap to the surface under the brush. | Every cell inside the globe is swallowed. |

- **Drag** to paint across the surface. **Hold still** to fill one round spot.
- Change the size with the **Radius** slider or **[** and **]**, from 0.6 to 12 cells.
- **Undo** and **Redo** (**Cmd/Ctrl+Z**, **Cmd/Ctrl+Shift+Z**) step through the last ten strokes.
- **Right drag** orbits without leaving the tool. **Orbit** (**V**) switches back to plain navigation.

### 3. Load a build

Click **Explore builds** and search the collection, or use **Load** to open any `.octa` file. **Save** writes your edited arrangement back out.

| Explore builds | Castle, ready to edit |
| :---: | :---: |
| <img alt="The build library" src="https://github.com/user-attachments/assets/16ba1b67-ea83-469a-9a65-9112d500e6db" /> | <img alt="The castle build in its world" src="https://github.com/user-attachments/assets/d0baa2da-e984-4454-942b-89fd8bf7b996" /> |

### 4. Make it yours

| Carve with a big black hole | Paint with a palette color |
| :---: | :---: |
| <img alt="A large black hole biting into a castle tower" src="https://github.com/user-attachments/assets/decf9ea6-a5ea-4e1e-96e8-be5b39cc1a0f" /> | <img alt="Brushing a red dot onto the Taj Mahal dome" src="https://github.com/user-attachments/assets/25c64140-c006-422a-a4fe-8c7a1184c2a9" /> |

Brush in uses **Auto** by default, so new cells copy the color of the cells they grow from. On a colored build, you can pick one of eight swatches, or click **More** to see all 16.

### Navigate

| Input | Action |
| --- | --- |
| Drag · right drag with a tool | Orbit |
| Wheel / pinch | Zoom |
| Shift + drag / middle drag | Pan |
| **F** | Reframe the build |
| **X** | X-ray view |
| **?** / **H** | Help card with every control |

[All controls →](docs/controls.md)

## Gallery

<img alt="Nine Octa builds: Eiffel Tower, Neuschwanstein, Golden Gate Bridge, Pirate Galleon, Floating Islands, Sydney Opera House, Saturn V Launch, Venice Grand Canal and Japanese Castle" src="https://github.com/user-attachments/assets/336b0e28-f52c-4106-8579-efc44a1a6c5f" width="100%" />

<sub>Eiffel Tower · Neuschwanstein · Golden Gate Bridge · Pirate Galleon · Floating Islands · Sydney Opera House · Saturn V Launch · Venice Grand Canal · Japanese Castle. [See all 85 →](docs/builds.md)</sub>

## Documentation

| Guide | Contents |
| --- | --- |
| [Controls](docs/controls.md) | Every mouse, toolbar and keyboard control, plus brushing tips |
| [Builds](docs/builds.md) | The full collection with cell counts, and regenerating builds in Go |
| [How Octa works](docs/concepts.md) | The cell lattice, presets, scripting your own arrangements, and rendering |
| [Architecture](docs/architecture.md) | Every GDScript class, data flow, shaders, workers and worlds |
| [Development](docs/development.md) | Tests, screenshots, the walkthrough video, web export and change recipes |
| [Go generator](gen/README.md) | CLI, density, geometry and grid APIs, components, and adding scenes |
| [OCTA format](docs/octa-format.md) | Binary layouts and cross-language compatibility |
| [Credits](docs/credits.md) | Artwork provenance and distribution notices |

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md), and run `python3 tools/check.py` before opening a pull request. Report bugs in [issues](https://github.com/makevoid/godocta/issues), and report vulnerabilities as described in [SECURITY.md](SECURITY.md).

The desktop app has been tested on macOS. Windows and Linux are untested, and the web export is an experimental preview.

## License

[MIT](LICENSE) © makevoid and Octa contributors.
