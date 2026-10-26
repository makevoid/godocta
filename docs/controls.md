# Controls

Everything the studio responds to, by mouse, toolbar and keyboard. Press **?** or
**H** in the app for the same list next to a diagram of the unit cell.

| Control | Action |
| --- | --- |
| Enter (welcome screen) / OCTA (toolbar) | Enter the studio / return to the welcome screen, keeping the current arrangement |
| Explore builds | Search by name; select a card to load it; **Esc** closes the collection |
| Builds ▾, or **5–9** | Open a `.octa` file from `structures/`; shortcuts select the first five alphabetically |
| Primitives ▾, or **1–4** | Assemble the cube, sphere, square pyramid or cylinder |
| Replay, **R**, or **Space** | Restart the current arrangement |
| Drag | Orbit |
| Wheel / pinch | Zoom |
| Shift + drag / middle drag | Pan |
| Frame / **F** | Restore framing |
| **?** / **H** | The help card: every control plus the unit diagram |
| Settings ▾ | X-ray view (**X**), auto orbit, and enabling the scoop tool |
| Save / Load | Write or open a compact `.octa` arrangement |
| Orbit / **V** or **Esc** | Navigation only: left drag orbits |
| Brush in / **B** | Paint cells in: new cells fly in and snap to the cells under the brush (or rest on the ground) and take the brush color. A stroke keeps the brush at the depth of its first hit, so it fills that globe and stops; drag to paint sideways, and when the globe is full the gizmo greys out and a toast asks for another click |
| Black hole / **E** | Swallow cells within a globe fixed at the first hit's depth for the stroke; hold or drag to clear that foreground area, then release and click again to target deeper cells |
| Undo / **Cmd/Ctrl+Z** | Undo the last brush, black-hole or scoop stroke; keeps up to **10 strokes**, including the latest pending edit |
| Redo / **Cmd/Ctrl+Shift+Z** or **Ctrl+Y** | Restore an undone stroke. A new edit clears redo; switching builds clears all history |
| Scoop / **S** | A half globe that shaves the surface about one layer deep; its cut plane is fixed per stroke so dragging planes the build. Enable it in Settings first |
| **←/→**, **↑/↓** | Tilt the scoop 45° around the screen X / Y axes (flat by default) |
| **[** / **]** or Brush size | Shrink / grow the brush (0.6–12 cells). Brush in and the black hole start at 3 cells, the scoop at 2, and each tool keeps the last size it was used at |
| Brush color | Auto (new cells copy the cells they grow from) or one of the build's palette colors: eight swatches, with More for all 16. Uncolored presets stay on auto |
| Right drag, or Alt + drag | Orbit while a tool is active |

## Brushing tips

- A stroke locks the brush globe at the depth of its first hit. Hold still to fill
  one round spot; drag to paint or carve sideways at that depth.
- Release and click again to reach deeper cells, or orbit (right drag) to get a
  new angle without leaving the tool.
- Edits settle about a second after a stroke, or right away before the next stroke,
  a tool change, Save or Replay. The cell count in the top bar updates then.
- Undo and Redo restore exact coordinates, order and colors without moving the view.
  Switching builds clears the history.
- Uncolored presets always brush in on **Auto**. Builds with a palette offer eight
  swatches, and **More** shows all 16.

See [How Octa works](concepts.md#editing) for what happens to the data during a stroke.
