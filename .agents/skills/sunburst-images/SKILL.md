---
name: sunburst-images
description: Generate production image assets with OpenAI GPT Image 2.5 "Sunburst" on fal.ai, using a library of eight prompts tuned and measured against the live endpoint — character reference sheets and expression atlases, software UI button kits and icon sets, game HUDs and inventory screens, and light-novel sci-fi background plates. Use this skill whenever the user wants to generate, render, or mock up any image asset — concept art, a character sheet, UI mockups, icons, buttons, a game HUD, key art, a background plate, a texture, a splash screen, placeholder or filler art for an app — or whenever they mention Sunburst, GPT Image, gpt-image-2.5, fal.ai image generation, or ask for "an image of" something. Also use it when they want help writing an image prompt, want an existing prompt improved, or ask why a generated image came out wrong (blurry edges, wrong text, low resolution, a background that would not key out) — the bundled findings explain the failure modes with measurements. Prefer this over writing raw fal.ai API calls by hand.
---

# Sunburst image generation

Sunburst is the precision variant of OpenAI GPT Image 2.5: it spends longer per
request in exchange for holding fine structure together. That is what makes
small labels, 1px HUD strokes and 16-glyph icon grids survive instead of
smearing. Use it when a frame will be examined closely or enlarged; for rough
drafts, drop `--quality low` and iterate cheaply first.

Everything here comes from generating and revising these prompts against the
live endpoint, so the defaults are measured rather than guessed.

## Setup

The CLI is `scripts/sunburst`, next to this file. It lives wherever this skill
is installed, which is not the user's project — so invoke it by its **absolute
path**, and set that path once at the start of a session:

```bash
SB=<absolute path to this skill>/scripts/sunburst
"$SB" doctor      # checks the key and the excon gem, spends nothing
```

The key is read from `FAL_KEY`, `FAL_AI_API_KEY`, `FAL_AI_KEY` or
`~/.fal_ai_api_key`, in that order. If `doctor` reports a missing credential,
ask the user for their key rather than guessing a path. If excon is missing:
`gem install excon`.

## Generating

Run it from the user's project directory — `--out` is resolved relative to
their cwd, while the CLI itself is self-contained and never touches their
Gemfile.

```bash
# What is available
"$SB" list
"$SB" show game_hud          # rendered prompt + why it is written that way

# Render library prompts (several at once is barely slower than one --
# everything is queued before anything is collected)
"$SB" generate game_hud character_sheet --out assets/concept

# Everything matching a tag
"$SB" generate --tag light_novel --out assets/backgrounds

# Your own prompt
"$SB" prompt --file prompt.md --name key-art --size 2048x1152 --quality xhigh
cat prompt.md | "$SB" prompt --name key-art
```

It prints the path of each image it writes, and writes a `<name>.json` sidecar
beside each one recording the exact prompt, settings and request id. Keep the
sidecars — a good frame whose prompt you cannot reconstruct is worth very little.

Flags: `--out` `--name` `--quality` `--size` `--format` `--background` `--num`
`--tag` `--quiet`. Anything you do not pass keeps the entry's tuned value, which
is the point of the library — override one thing at a time so you know what
changed the result.

Expect 30–60s per image at `xhigh`. If a run is interrupted, the request keeps
going: `"$SB" status REQUEST_ID`.

## Writing a prompt

Write prompts as **labeled blocks**, one per concern, not as a paragraph:

```
PURPOSE:      what the image is for, and its shape
SUBJECT:      what is in it
LAYOUT:       where things sit relative to each other
STYLE:        medium and finish
MATERIALS:    what things are made of
PALETTE:      named colours, with hex where it matters
TEXT:         every string, quoted, with its repeat count
LIGHT:        direction, quality, shadows
CAMERA/FRAME: lens, height, projection
TECHNICAL:    margins, background, edge quality
EXCLUDE:      what must not appear
```

Use the blocks the image actually needs; omit the rest. This format earns its
keep three ways: a revision becomes a one-block edit so you are comparing one
variable rather than two, the model stops averaging instructions that would
otherwise share a sentence, and the result is diffable.

The habits that most change output quality, roughly in order:

**Quote every string and say how many times it appears.** This is the single
highest-value line in any prompt that renders type. Left to itself the model
writes plausible-looking gibberish in small numerals. Write
`"HULL 87%", appearing exactly once` or `the label appears three times, once per
state column` — the latter is what stops it labelling one state and leaving the
others blank.

**Name every item; never say "various".** Nine named props beat "assorted loot"
every time, and the same goes for icons, expressions and detail callouts.

**Say how the drawing should look, not how the world works.** "The floor curves
upward, revealing the station's spin" is a fact about the setting and gets
silently dropped. "Draw the floor curving visibly upward at both edges of the
frame" is an instruction about the image and renders.

**Rule out the dominant prior explicitly.** If the model has far more training
data for a neighbouring style, naming your style is not enough — you have to
name the thing you do *not* want. "Hand-painted 2D illustration, not a
photograph and not a photorealistic 3D render" flipped an interior that kept
coming back as CG.

**Always write an EXCLUDE block.** Background plates need "no people, no
characters" in both `PURPOSE` and `EXCLUDE` or they get populated. Watermarks
and signatures need excluding by name.

**Carve exceptions out of your own exclusions.** "No filled shapes" turned a
record dot into an outline; it needed "the only filled shape in the whole set"
and a narrowed exclusion.

Layout zones are a strong suggestion, not a guarantee — expect the occasional
regrouping and check the result rather than assuming.

## Two things that will bite you

**`--background transparent` does not give a crisp cut-out.** It returns
soft-matted, bloomed artwork, and prompting against it barely helps. Render on a
flat key colour with `--background opaque` and key it out locally instead — that
came back fully opaque and pin sharp at a *lower* quality tier and half the
resolution. Pick a key far in hue from the asset's own palette. The three
cut-out entries in the library already do this and declare their `key_color`.

**Size presets are small.** `landscape_16_9` returns 1088×608. Pass explicit
dimensions — `--size 2048x1152` costs the same request. Edges must divide by 16,
neither over 3840, aspect no wider than 3:1.

`references/findings.md` has the measurements behind both, plus the quality-tier
and cost guidance. Read it when a result disappoints and you want to know
whether the prompt or the parameters are at fault.

## The prompt library

`references/prompts.md` holds all eight prompts in full, with the reasoning
behind each. Read it when the user wants a prompt to start from, wants to see
how a particular effect is specified, or asks for something close to one of
these:

| key | makes |
| --- | --- |
| `character_sheet` | 3-view turnaround, head studies, labelled detail callouts |
| `expression_sheet` | 4×3 expression atlas, one face, keyable |
| `ui_buttons` | 4 controls × 3 states, software button kit, keyable |
| `ui_icon_set` | 16 monoline toolbar glyphs on one grid, keyable |
| `game_hud` | full 16:9 space-sim HUD with live readouts |
| `game_inventory` | diegetic inventory terminal, item grid + detail pane |
| `light_novel_rooftop` | unpopulated sci-fi exterior background plate |
| `light_novel_interior` | unpopulated sci-fi interior background plate |

These are starting points, not a catalogue. The usual move is to take the
closest entry, swap `SUBJECT` and `PALETTE` for the user's actual project, and
leave the structural blocks — `LAYOUT`, `TECHNICAL`, `EXCLUDE` — alone, since
those are what was tuned.

## Ruby API

For anything beyond one-off renders, the bundled library is importable:

```ruby
$LOAD_PATH.unshift "<this skill>/scripts/lib"
require "sunburst"

prompt = Sunburst::PromptLibrary[:game_hud].merge(
  subject: "A fantasy airship cockpit overlay above a cloud sea"
)
Sunburst.generate(prompt: prompt, name: "airship-hud",
                  image_size: { width: 2048, height: 1152 }, quality: "xhigh")
```

`Prompt#merge` returns a copy, so library prompts stay pristine while you
iterate. `Sunburst.generator.generate_batch(specs)` queues a whole set before
collecting any of it.
