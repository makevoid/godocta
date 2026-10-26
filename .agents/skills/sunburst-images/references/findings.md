# Sunburst: measured findings

Everything here came out of generating and revising the bundled prompts against
the live endpoint. Read it when a result disappoints and you need to know
whether the prompt or the parameters are at fault.

## Finding 1: transparent backgrounds come back soft-matted

`background: "transparent"` on this endpoint does **not** produce a crisp
cut-out. Across six renders it returned soft-matted artwork with heavy bloom:
alpha that never reaches 255, and 10–13% of the canvas sitting in partial
transparency.

Prompting against it barely helped. An explicit alpha contract — *"hard-edged
alpha, every pixel fully opaque or fully transparent, no glow, no bloom, no
halo, no feathered edges"* — moved the soft-alpha share on a button kit from
10.7% to 3.9% and did nothing at all to an icon set. Both still looked blurred.
Switching icon strokes from near-white to near-black changed nothing either,
which ruled out the colour and pointed at the alpha matting itself.

| render | mode | quality | size | soft alpha | verdict |
| --- | --- | --- | --- | --- | --- |
| icons v1 | transparent | max | 1536 | 13.0% | blurred, heavy bloom |
| icons v2 (anti-glow prompt) | transparent | max | 1536 | 13.3% | unchanged |
| icons probe (dark strokes) | transparent | high | 1024 | 8.6% | still blurred |
| icons probe (flat backdrop) | **opaque** | high | 1024 | **0%** | pin sharp |

The opaque version was sharper at `high`/1024px than the transparent version at
`max`/1536px — cheaper *and* better.

**Render cut-out assets on a flat key colour and key them out locally.** Two
details matter:

- Ask for *"one single solid colour at an identical RGB value in every
  background pixel"* and explicitly forbid gradient, vignette, texture and
  lighting falloff. Without that you get a subtle gradient that keys badly.
- **Pick a key far in hue from the asset's own palette.** A button kit first
  used green and its electric-lime primary pill came back sitting right on top
  of the key — unkeyable. On magenta `#FF00D4` it separated cleanly.

Measured flatness of the final renders, as the share of canvas within tolerance
18 of the requested key: icon set 91%, button kit 72%, expression sheet 41% (the
heads cover most of that canvas).

A footnote worth knowing, because it will mislead you: **this model's alpha tops
out at 253/254 and never reaches exactly 255.** "Zero fully-opaque pixels" is an
artifact of that encoding, not evidence of a problem. Measure the soft-falloff
share (alpha 1–249) instead.

## Finding 2: size presets are small

`landscape_16_9` returned **1088×608**. The same prompt with explicit
`2048×1152` costs the same request and returns something printable. Use presets
only when you genuinely do not care about resolution.

Custom sizes need both edges divisible by 16, neither over 3840, and an aspect
ratio no wider than 3:1.

Preset names, for reference: `square_hd`, `square`, `portrait_4_3`,
`portrait_16_9`, `landscape_4_3`, `landscape_16_9`, `auto`.

## Quality tiers

`auto` `low` `medium` `high` `xhigh` `max`, defaulting to `high`.

Nothing in the bundled library sits below `high` — small type and thin strokes
fall apart beneath it. `xhigh` is the workhorse. `max` earned its keep nowhere
once the cut-out entries moved to an opaque backdrop; the icon set dropped from
`max` to `high` and got *better*, because the backdrop change mattered far more
than the tier.

The sensible workflow: draft wide and cheap at `low`, promote the brief that
works to `high`/`xhigh`, and spend the top tier only on frames that will be
enlarged or examined closely. Per-image cost ranges from roughly $0.004 to
$0.40 depending on size and tier, so a careless `max` sweep is expensive.

Observed latency at `xhigh`, 1536–2048px: 30–60s per image. Batching queues
everything before collecting any of it, so eight assets take about as long as
the slowest one rather than the sum.

## API shape

Endpoint: `openai/gpt-image-2.5/sunburst/text-to-image` on `https://queue.fal.run`.
Auth header is `Authorization: Key <token>` (not `Bearer`).

**You submit to the endpoint path but poll the app path.** Submitting goes to
`/openai/gpt-image-2.5/sunburst/text-to-image`; status and result live under
`/openai/gpt-image-2.5/requests/{id}` — the variant suffix is dropped. Polling
the submit path returns **405**. The submit response hands back the correct
`status_url`, `response_url` and `cancel_url`; use those rather than building
your own.

Inputs: `prompt` (required), `image_size`, `quality`, `background`
(`auto`/`transparent`/`opaque`), `output_format` (`png`/`jpeg`/`webp`),
`output_compression` (0–100, jpeg/webp only), `num_images`, `sync_mode`.
Output: `{ "images": [{ "url", "width", "height", "content_type", "file_name",
"file_size" }] }`.

Note the endpoint does not always return `width`/`height`, so do not rely on
them; read the dimensions off the downloaded file if you need them.

## Prompting failures worth recognising

| symptom | cause | fix |
| --- | --- | --- |
| gibberish in small type | copy not quoted | quote every string verbatim |
| label on one state only | no repeat count | "appears three times, once per column" |
| generic or wrong props | "various", "assorted" | name every item |
| a stated world-fact ignored | described the setting, not the drawing | rephrase as an instruction about the image |
| photoreal when you wanted illustration | neighbouring style has a larger prior | name what you do *not* want |
| figures in a background plate | no exclusion | "no people, no characters" in PURPOSE *and* EXCLUDE |
| a deliberate exception normalised away | your own EXCLUDE caught it | carve it out by name |
| blurred, bloomed cut-out | `background: transparent` | flat key colour + `opaque` |
| low resolution | size preset | explicit `WIDTHxHEIGHT` |
