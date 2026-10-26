# frozen_string_literal: true

module Sunburst
  # Curated, tested prompts for the asset types Sunburst handles well.
  #
  # Every entry follows the labeled-block structure fal recommends for
  # GPT Image 2.5: one block per concern, exact copy quoted with its repeat
  # count, and an explicit EXCLUDE block. Each entry also pins the settings it
  # was tuned at -- quality and image_size are part of the prompt, not an
  # afterthought, because small type and fine linework fall apart below `high`.
  module PromptLibrary
    # A prompt plus the settings it was validated at.
    Entry = Struct.new(:key, :title, :prompt, :settings, :notes, :tags, :key_color,
                       keyword_init: true) do
      # Keyword arguments ready to splat into Generator#generate.
      def to_spec(name: nil, **overrides)
        { name: name || key.to_s.tr("_", "-"), prompt: prompt }
          .merge(settings)
          .merge(overrides)
      end

      def to_s
        prompt.to_s
      end
    end

    # The measured finding behind every cut-out entry in this library.
    #
    # `background: "transparent"` on this endpoint does NOT give you a crisp
    # cut-out. Across six renders it returned soft-matted artwork: alpha that
    # never reaches 255, 10-13% of the canvas in partial transparency, and a
    # heavy bloom around every shape. Prompting against it does not help --
    # "no glow, no bloom, no halo, hard-edged alpha" moved the soft-alpha share
    # from 10.7% to 3.9% on the button kit and not at all on the icon set, and
    # both still looked blurred. Switching the stroke from near-white to
    # near-black changed nothing, so it is the alpha matting, not the colour.
    #
    # Rendering the same prompt on a flat solid backdrop with
    # `background: "opaque"` came back 100% opaque and pin sharp at `high` and
    # 1024px -- better than the transparent version at `max` and 1536px. So:
    # render on a flat key colour, then key it out locally.
    # Pick a key that is far in hue from everything in the asset's own palette.
    # The button kit first used green and the lime primary pill nearly matched it,
    # which would have made it miserable to key cleanly.
    KEY_GREEN = "#00B140"
    KEY_MAGENTA = "#FF00D4"
    KEY_INK = "#15171A"

    KEYABLE_BACKDROP = lambda do |hex|
      "Place the artwork on a perfectly flat, even #{hex} background that fills " \
      "the canvas edge to edge -- one single solid colour at an identical RGB " \
      "value in every background pixel, no gradient, no vignette, no texture, " \
      "no lighting falloff, no panel or card behind the " \
      "artwork. This background is a key colour that will be removed, so no " \
      "element of the artwork may use #{hex} and nothing may spill onto it: no " \
      "glow, no bloom, no halo, no drop shadow, no blurred or feathered edges."
    end

    NO_ALPHA_SPILL =
      "no glow, no bloom, no halo, no outer or drop shadow, no blurred or " \
      "feathered edges, no gradient or texture in the background"

    module_function

    def entries
      @entries ||= build_entries.freeze
    end

    def keys
      entries.keys
    end

    def [](key)
      fetch(key).prompt
    end

    def fetch(key)
      entries.fetch(key.to_sym) do
        raise InvalidInputError,
              "unknown prompt #{key.inspect}; known: #{keys.join(", ")}"
      end
    end

    def entry(key)
      fetch(key)
    end

    # Entries whose tags include every tag given.
    def tagged(*tags)
      wanted = tags.flatten.map(&:to_sym)
      entries.values.select { |e| (wanted - e.tags).empty? }
    end

    # The showcase set: one of each asset type the library covers.
    def showcase
      %i[character_sheet ui_buttons game_hud light_novel_rooftop].map { |k| fetch(k) }
    end

    def build_entries
      {
        character_sheet: character_sheet,
        expression_sheet: expression_sheet,
        ui_buttons: ui_buttons,
        ui_icon_set: ui_icon_set,
        game_hud: game_hud,
        game_inventory: game_inventory,
        light_novel_rooftop: light_novel_rooftop,
        light_novel_interior: light_novel_interior
      }
    end

    # --- character art ------------------------------------------------------

    def character_sheet
      Entry.new(
        key: :character_sheet,
        title: "Character reference sheet (3-view turnaround + detail callouts)",
        tags: %i[character sheet game scifi],
        settings: { image_size: { width: 2048, height: 1536 }, quality: "xhigh",
                    output_format: "png" },
        notes: <<~NOTES,
          The orthographic + shared-ground-line instruction in CAMERA/FRAME is what
          keeps the three views at one scale; without it the back view drifts taller.
          Naming the callouts and their leader lines is what turns a pin-up into a
          usable reference sheet. Needs xhigh: at `high` the small uppercase labels
          under the detail panels start to smear.
        NOTES
        prompt: Prompt.new(
          purpose: "Production character reference sheet for a 2D action-adventure game, " \
                   "read at 100% zoom by modellers and animators. One flat sheet, horizontal.",
          subject: "\"Vela Ashford\", a 24-year-old deep-space salvage courier: lean athletic " \
                   "build, shoulder-length copper hair tied in a half-knot, pale freckled " \
                   "skin, grey-green eyes, a faded hairline scar through the left eyebrow. " \
                   "Calm, faintly wry expression. Identical face, proportions and costume in " \
                   "every view on the sheet.",
          layout: "Top band: three full-body turnarounds at identical height on one shared " \
                  "ground line -- front, three-quarter, back, evenly spaced. Bottom-left " \
                  "quadrant: a row of four head studies at the same scale -- neutral, half " \
                  "grin, shouting, visor down. Bottom-right quadrant: three square detail " \
                  "panels -- glove seam with wrist terminal, boot magclamp sole, backpack " \
                  "latch -- each joined to the front view by a thin straight leader line.",
          style: "Clean semi-realistic concept art: confident ink linework, flat cel shading " \
                 "in two tonal steps, the look of a printed game art book.",
          materials: "Oil-stained canvas flight jacket over a ribbed thermal layer, cargo " \
                     "trousers with reinforced knee panels, scuffed magnetic-soled boots, a " \
                     "cracked amber visor pushed up on her forehead, a forearm terminal with " \
                     "a hairline-lit readout.",
          palette: "Warm desaturated base -- rust orange, bone, olive drab -- with one " \
                   "saturated cyan accent used only on the terminal glow and the visor edge.",
          text: "Small hand-lettered uppercase sans labels, each appearing exactly once and " \
                "spelled exactly as written: \"FRONT\", \"3/4\" and \"BACK\" centred under " \
                "their turnarounds; \"EXPRESSIONS\" above the head row; \"WRIST TERMINAL\", " \
                "\"MAG SOLE\" and \"PACK LATCH\" under their detail panels. No other text.",
          light: "Flat, even, shadowless reference lighting. No cast shadows anywhere on the " \
                 "sheet; form reads from line and local colour only.",
          camera: "Orthographic projection, eye level, zero perspective distortion. All three " \
                  "turnarounds share one scale and one ground line.",
          technical: "Off-white paper ground with a faint 5% grey grid. Keep a clear margin " \
                     "of roughly 3% of the image width on all four edges.",
          exclude: "No environment or scenery, no props beyond those listed, no colour swatch " \
                   "chips, no duplicate or alternate characters, no watermark, no signature."
        )
      )
    end

    def expression_sheet
      Entry.new(
        key: :expression_sheet,
        title: "Expression sheet, 4x3 grid, keyed out of a flat backdrop",
        tags: %i[character sheet keyable game],
        key_color: KEY_GREEN,
        settings: { image_size: { width: 1536, height: 1536 }, quality: "xhigh",
                    background: "opaque", output_format: "png" },
        notes: <<~NOTES,
          v1 asked for `background: "transparent"` and got a painted brown backdrop
          anyway -- the flag is a hint, not a guarantee. v2 added an explicit alpha
          contract and still got brown. v3 stops fighting it and renders on a flat
          #00B140 key colour instead; see KEYABLE_BACKDROP.

          Consistency across twelve cells was never the problem: one named face
          plus "only the expression changes" carried it on the first try.
        NOTES
        prompt: Prompt.new(
          purpose: "Expression atlas for a visual-novel dialogue system; each cell is cut " \
                   "out and swapped at runtime, so cells must not touch or overlap.",
          subject: "Head and shoulders of \"Vela Ashford\" -- copper half-knot hair, pale " \
                   "freckled skin, grey-green eyes, hairline scar through the left eyebrow, " \
                   "collar of an oil-stained canvas flight jacket. The same face, hairline " \
                   "and collar in all twelve cells; only the expression changes.",
          layout: "A 4-column by 3-row grid of twelve head studies at one identical scale, " \
                  "even gutters, eyes on the same horizontal line within each row. Reading " \
                  "order: neutral, soft smile, broad grin, laughing; surprised, worried, " \
                  "annoyed, furious; exhausted, smug, deadpan, tearful.",
          style: "Anime-adjacent cel illustration, crisp 2px ink outline, two-step shading, " \
                 "no gradients inside the fills.",
          palette: "Rust orange, bone and olive on the costume; cyan only as a thin rim on " \
                   "the jacket collar hardware.",
          light: "Flat even light, no cast shadow under the chin, no shadow on the canvas.",
          camera: "Straight-on bust framing, orthographic, identical head size in every cell.",
          technical: "#{KEYABLE_BACKDROP.call(KEY_GREEN)} No grid lines drawn, no cell " \
                     "borders. Each head is a self-contained cut-out; cells must not touch " \
                     "or overlap.",
          exclude: "No text, no labels, no numbers, no body below the collar, no watermark, " \
                   "#{NO_ALPHA_SPILL}."
        )
      )
    end

    # --- software UI --------------------------------------------------------

    def ui_buttons
      Entry.new(
        key: :ui_buttons,
        title: "Software UI button kit, 4 controls x 3 states, keyable",
        tags: %i[ui software keyable kit],
        key_color: KEY_MAGENTA,
        settings: { image_size: { width: 1536, height: 1536 }, quality: "xhigh",
                    background: "opaque", output_format: "png" },
        notes: <<~NOTES,
          The repeat count on the label copy ("three times, once per state column")
          is the fix for the usual failure here, where the model writes the label on
          the rest state and leaves hover/pressed blank -- v1 got all six labels
          right on the first try. What v1 got wrong was alpha: "2px outer glow" and
          "soft drop shadow" on a transparent canvas turned the whole sheet into a
          lime bloom. Removing the glow language (v2) cut the soft-alpha share from
          10.7% to 3.9% but still left feathered edges, so v3 renders on a flat
          #FF00D4 key colour instead -- magenta, not the green the other cut-out
          entries use, because the lime primary pill would have sat right on top of
          a green key. Keeping every effect inside the button outline is worth
          holding onto either way -- it is what makes the three
          columns read as one control in three states rather than three designs.
        NOTES
        prompt: Prompt.new(
          purpose: "Asset sheet of UI buttons for a desktop audio-mastering application. " \
                   "Each button is cut out and shipped at 1x and 2x, so edges must be clean " \
                   "and nothing may overlap.",
          subject: "Four controls, each shown in three states, from one consistent design " \
                   "system: a wide primary pill button, a wide secondary outlined button, a " \
                   "small square icon button carrying a play triangle, and a round record " \
                   "button with a filled dot.",
          layout: "A 3-column by 4-row grid with even 6% gutters. Columns left to right are " \
                  "the states: rest, hover, pressed. Rows top to bottom are the controls in " \
                  "the order listed. Every control keeps identical size and position across " \
                  "its row; only the state styling changes.",
          style: "Modern dark-mode product UI, flat surfaces with a 1px inner top highlight, " \
                 "8px corner radius on the square control, fully rounded ends on the pills. " \
                 "The primary pill is filled solid electric lime edge to edge with near-black " \
                 "text on it; the other three controls are filled graphite. The three states " \
                 "differ only by fill brightness and border, never by glow: rest is the base " \
                 "fill; hover is one step brighter with a 2px solid hard-edged lime border; " \
                 "pressed is one step darker with a 1px inner top shadow and the content " \
                 "nudged down 1px. The lime fill of the primary pill is present in all three " \
                 "states.",
          palette: "Graphite #23262B surfaces, #E8EAED text and icons. The primary pill's " \
                   "surface is electric lime #B4F03C with #15171A text on it. #FF4D4D only on " \
                   "the record dot.",
          text: "Exactly two label strings, in a medium-weight geometric sans. " \
                "\"RENDER MASTER\" centred on the primary pill, appearing three times, once " \
                "per state column. \"Preview\" centred on the outlined button, appearing " \
                "three times, once per state column. Spelled identically every time, same " \
                "size and letterspacing in all states. The icon button and record button " \
                "carry no text at all.",
          light: "Flat even lighting baked into the fills. All depth comes from inner " \
                 "highlights and inner shadows contained inside each button's own outline; " \
                 "nothing is cast outside a button's edge.",
          camera: "Dead-on flat orthographic view, no perspective, no tilt, no mockup device.",
          technical: "#{KEYABLE_BACKDROP.call(KEY_MAGENTA)} Each button is a self-contained " \
                     "cut-out with crisp pixel-level edges; the gutters between buttons show " \
                     "the flat key colour and nothing else.",
          exclude: "No window chrome, no title bar, no cursor or pointer, no annotation " \
                   "arrows or callouts, no state names written on the sheet, no watermark, " \
                   "no text other than the two labels listed, #{NO_ALPHA_SPILL}."
        )
      )
    end

    def ui_icon_set
      Entry.new(
        key: :ui_icon_set,
        title: "Line icon set, 16 glyphs on one 4x4 grid, keyable",
        tags: %i[ui software keyable kit icons],
        key_color: KEY_INK,
        settings: { image_size: { width: 1536, height: 1536 }, quality: "high",
                    background: "opaque", output_format: "png" },
        notes: <<~NOTES,
          Naming the glyphs one by one beats "16 generic icons" every time -- v1 got
          all sixteen, in order, on the first try, and so did every pass since.

          This entry is where the transparent-background finding was pinned down.
          v1 (`transparent`, `max`, 1536px) came back as soft white glyphs inside a
          huge bloom. v2 added an explicit no-glow alpha contract: no change. A
          probe with near-black strokes: no change, so it was not the colour. The
          same prompt on a flat #15171A backdrop with `background: "opaque"` came
          back 100% opaque and pin sharp -- at `high` and 1024px, which is why this
          entry no longer pays for `max`. Key the backdrop out locally.

          One more pass was needed for the record dot: a single filled glyph in an
          otherwise monoline set gets normalised to an outline unless you call it
          out as "the only filled shape in the whole set" and carve it out of the
          "no filled shapes" exclusion.
        NOTES
        prompt: Prompt.new(
          purpose: "Icon set for the toolbar of a desktop audio-mastering application. Each " \
                   "glyph is exported alone and rendered as small as 20px, so stroke weight " \
                   "must be uniform and shapes must stay open.",
          subject: "Sixteen monoline outline icons, drawn on one shared 24-unit square grid " \
                   "with a 2-unit stroke, rounded caps and rounded joins, and a 2-unit " \
                   "optical margin inside each cell.",
          layout: "A 4-by-4 grid, even gutters, every glyph optically centred and filling " \
                  "the same visual weight. Reading order row by row: waveform, equaliser " \
                  "sliders, play triangle, pause bars; record dot, loop arrows, metronome, " \
                  "headphones; folder, save disk, export arrow leaving a tray, trash can; " \
                  "gear, bell, search magnifier, user circle.",
          style: "Geometric monoline, flat, single colour, no fills, no gradients, no " \
                 "shadows. Consistent 2-unit corner radius throughout.",
          palette: "One flat colour only: solid #E8EAED strokes at full opacity. Not " \
                   "emissive, not glowing, not semi-transparent.",
          technical: "#{KEYABLE_BACKDROP.call(KEY_INK)} Strokes are crisp, solid and of " \
                     "equal width in every glyph. No grid lines drawn, no cell borders.",
          camera: "Dead-on flat orthographic view, no perspective.",
          exclude: "No text, no labels, no numbers, no colour, no filled shapes other than " \
                   "the record dot, no duplicated glyphs, no extra icons beyond the sixteen " \
                   "listed, no watermark, #{NO_ALPHA_SPILL}."
        )
      )
    end

    # --- game UI -----------------------------------------------------------

    def game_hud
      Entry.new(
        key: :game_hud,
        title: "Space-combat sim HUD, full 16:9 screen mockup with live readouts",
        tags: %i[game ui hud scifi],
        settings: { image_size: { width: 2048, height: 1152 }, quality: "xhigh",
                    output_format: "png" },
        notes: <<~NOTES,
          The single highest-value line here is quoting every readout verbatim with
          its value. Left to itself the model writes plausible-looking gibberish in
          the small numerals. Pinning each element to a named screen edge is what
          stops the HUD piling into the centre. This is the frame that most rewards
          Sunburst over Flare -- the tick marks on the compass strip survive.
        NOTES
        prompt: Prompt.new(
          purpose: "Target-look HUD mockup for a space-combat sim, shown to the team as the " \
                   "spec. Full-screen 16:9 frame, readouts legible at 100%.",
          subject: "A thin-stroke holographic HUD composited over a dim view through a " \
                   "cockpit canopy: a drifting asteroid field, one distant ringed planet low " \
                   "on the right, two faint running lights of a wingman at upper left.",
          layout: "Centre: a reticle of four floating brackets with a small lead-indicator " \
                  "dot offset slightly right. Top centre: a narrow compass strip with fine " \
                  "degree tick marks and three triangular contact markers. Bottom left: a " \
                  "curved hull-integrity arc of segmented pips wrapping a small ship " \
                  "silhouette with its port wing segment darkened. Bottom right: a weapons " \
                  "column of four slots with a heat bar beside each. Left edge: a vertical " \
                  "throttle ladder with a filled bar at about two thirds. Right edge: a " \
                  "wireframe radar sphere with three plotted blips.",
          style: "Thin-stroke vector HUD, 1 to 2 pixel lines, additive cyan glow on near " \
                 "black, faint horizontal scanlines across the overlay only, a hair of " \
                 "chromatic fringing at the frame edges.",
          palette: "Cyan #4FE3FF for all nominal HUD strokes, amber #FFB341 for warnings, " \
                   "red #FF4A4A only on the damaged hull segment. The scene behind stays " \
                   "desaturated blue-grey so the overlay dominates.",
          text: "Condensed uppercase technical sans. Each string appears exactly once and is " \
                "spelled and valued exactly as written: \"HULL 87%\" and \"SHIELD 42%\" " \
                "beside the hull arc; \"VELOCITY 310 M/S\" under the throttle ladder; " \
                "\"CONTACTS 3\" under the radar sphere; \"TARGET LOCK\" under the reticle. " \
                "The four weapon slots are labelled, top to bottom, \"PLS\", \"RLY\", " \
                "\"MSL\", \"EMP\", with \"240/600\" beside \"PLS\" only. Every numeral must " \
                "be sharp and readable.",
          light: "The HUD is self-lit and emissive; the scene behind it is lit only by cold " \
                 "starlight and the planet's faint rim.",
          camera: "First-person cockpit view, horizon level, wide field of view. Keep the " \
                  "central third clear of HUD elements apart from the reticle.",
          technical: "Full-bleed 16:9. HUD elements sit at least 4% in from every screen " \
                     "edge. Strokes stay thin and even; no element overlaps another.",
          exclude: "No cockpit frame, canopy struts or dashboard geometry, no pilot hands, " \
                   "no mouse cursor, no subtitles or chat box, no studio logo, no watermark, " \
                   "no text beyond the strings listed."
        )
      )
    end

    def game_inventory
      Entry.new(
        key: :game_inventory,
        title: "RPG inventory screen, diegetic panel with item grid",
        tags: %i[game ui panel scifi],
        settings: { image_size: { width: 2048, height: 1152 }, quality: "xhigh",
                    output_format: "png" },
        notes: <<~NOTES,
          Grid-plus-sidebar screens are where "one block per concern" pays for
          itself: LAYOUT fixes the furniture, TEXT fixes the copy, and a revision
          to either is a one-block edit. Nine named items beat "various loot".
        NOTES
        prompt: Prompt.new(
          purpose: "Inventory screen mockup for a sci-fi salvage RPG, presented as a " \
                   "diegetic terminal the character is physically holding open.",
          subject: "A battered handheld terminal UI: a 3-by-3 item grid on the left, an item " \
                   "detail pane on the right, a thin stat strip along the bottom.",
          layout: "Left two thirds: nine square slots in a 3-by-3 grid, seven filled and the " \
                  "last two empty with a faint dashed outline. Items, reading order: a coiled " \
                  "copper power cell, a cracked amber visor, a folded thermal blanket, a " \
                  "sealed ration tin, a bundle of salvaged fibre optics, a hull patch gun, a " \
                  "data spike. The power cell slot is selected, ringed in lime. Right third: " \
                  "a detail pane showing the power cell rotated three-quarters on a dark " \
                  "plinth above three short stat rows. Bottom: a strip of four small gauges.",
          style: "Diegetic retro-industrial interface: etched metal bezel, slightly convex " \
                 "glass with a faint vignette, phosphor-green readouts with a low bloom, " \
                 "visible dust and a hairline crack across the upper right of the glass.",
          palette: "Gunmetal and oxidised brass housing, phosphor green #7CFFA0 for type and " \
                   "rules, electric lime #B4F03C only on the selection ring, warm rust on " \
                   "the item renders.",
          text: "Condensed uppercase sans, each string exactly once, spelled as written: " \
                "\"CARGO 7/9\" above the grid; \"POWER CELL -- SALVAGED\" as the detail pane " \
                "heading; the three stat rows read \"CHARGE 62%\", \"MASS 1.4 KG\" and " \
                "\"VALUE 310 CR\"; the four bottom gauges are labelled \"O2\", \"PWR\", " \
                "\"TEMP\", \"MASS\".",
          light: "Screen-lit from within, a soft green glow spilling onto the bezel, one cool " \
                 "specular streak across the glass from upper left.",
          camera: "Dead-on flat view of the terminal, filling the frame edge to edge, no tilt.",
          technical: "Crisp small type; grid gutters even; the dashed empty-slot outline is " \
                     "clearly dashed, not solid.",
          exclude: "No hands or fingers, no cursor, no tooltips, no company logo, no " \
                   "watermark, no text beyond the strings listed."
        )
      )
    end

    # --- light novel backgrounds --------------------------------------------

    def light_novel_rooftop
      Entry.new(
        key: :light_novel_rooftop,
        title: "Light novel background: arcology rooftop garden at blue hour",
        tags: %i[background light_novel scifi exterior],
        settings: { image_size: { width: 2048, height: 1152 }, quality: "xhigh",
                    output_format: "png" },
        notes: <<~NOTES,
          Two lines do the heavy lifting for a usable background plate: "no figures,
          characters are composited later" in PURPOSE, repeated in EXCLUDE, and the
          instruction to keep the lower third low-contrast so dialogue type can sit
          over it. Without the first, the model populates the rooftop every time.

          Use an explicit {width:, height:} rather than the `landscape_16_9` preset:
          the preset returned 1088x608 here, while the same prompt at 2048x1152
          costs the same request and gives you something you can actually print.
        NOTES
        prompt: Prompt.new(
          purpose: "Unpopulated background plate for a light-novel colour insert. Characters " \
                   "are drawn and composited separately, so the plate contains no figures.",
          subject: "The rooftop garden of an arcology tower at blue hour, 400 metres up: " \
                   "raised planters of tall silver grass and pale hydrangeas, a weathered " \
                   "steel railing along the near edge, a service stair housing at left with " \
                   "its door propped open, a line of solar petals folded for the night at " \
                   "right.",
          composition: "Looking out and slightly down over the city. Beyond the railing, " \
                       "layered megastructure towers recede into haze, linked by two slender " \
                       "sky bridges; a maglev ribbon curves between them carrying a lit " \
                       "train; distant aircars drift as soft points of light. The horizon " \
                       "sits just above centre.",
          style: "Hand-painted 2D anime background art for a light novel colour insert: " \
                 "crisp ruled architectural linework under soft painterly gradients, visible " \
                 "brush texture in the clouds, the finish of a printed colour insert. " \
                 "Illustration, not a photograph and not a photorealistic 3D render.",
          palette: "Deep indigo and teal sky grading to warm peach at the horizon, cool " \
                   "lavender shadow in the planters, warm sodium-amber window light scattered " \
                   "across the towers, one cyan accent on the maglev ribbon.",
          light: "Last minutes of dusk. Soft ambient skylight from above, warm artificial " \
                 "light rising from the city below and catching the undersides of the " \
                 "railing and the grass. Long soft shadows, nothing harsh.",
          camera: "Cinematic 16:9, eye level at railing height, wide lens, deep focus with a " \
                  "gentle atmospheric falloff into the distance.",
          technical: "Keep the lower third quiet and low in contrast so dialogue type can be " \
                     "set over it. No vignette so strong that the corners go black.",
          exclude: "No people, no characters, no animals, no text, no speech bubbles, no " \
                   "signage with readable lettering, no logos, no UI elements, no watermark, " \
                   "no signature."
        )
      )
    end

    def light_novel_interior
      Entry.new(
        key: :light_novel_interior,
        title: "Light novel background: orbital station dormitory, morning",
        tags: %i[background light_novel scifi interior],
        settings: { image_size: { width: 2048, height: 1152 }, quality: "xhigh",
                    output_format: "png" },
        notes: <<~NOTES,
          Two v1 misses, both fixed in STYLE. "Japanese light-novel illustration"
          alone loses to photorealism on an interior -- the model has far more
          photoreal sci-fi interiors to draw on, so v1 came back as a 3D render.
          Naming the medium as hand-painted 2D and explicitly ruling out photoreal
          CG is what holds the line. The spin curvature also needs to be an
          instruction about the drawing ("draw the floor line curving up"), not a
          fact about the world, or it gets silently dropped.

          Listing lived-in clutter by name is still what separates a set from a
          render. Use explicit dimensions, not the preset -- see
          [[light_novel_rooftop]] notes.
        NOTES
        prompt: Prompt.new(
          purpose: "Unpopulated interior background plate for a light-novel scene; " \
                   "characters composited later.",
          subject: "A two-bunk student dormitory aboard a spinning orbital station, early " \
                   "ship-morning: fitted bunks against the left wall with a ladder to the " \
                   "upper, a shared desk under a wide curved window on the far wall, a " \
                   "pinboard of handwritten notes and photographs, a potted lemon tree " \
                   "strapped into a floor bracket.",
          composition: "Draw the floor and the row of ceiling panels curving visibly upward " \
                       "at both the left and right edges of the frame, so the eye reads the " \
                       "room as the inside of a spinning ring. Through the window, the " \
                       "planet's terminator line " \
                       "and a sliver of the station's outer ring. Desk clutter: a half-drunk " \
                       "mug, a folded jacket over the chair back, an open notebook, a loose " \
                       "cable, a single sock on the floor.",
          style: "Hand-painted 2D anime background art for a light novel colour insert: " \
                 "visible brush strokes, flat blocked colour shapes with painted edges, " \
                 "crisp ruled linework on the hard surfaces, in the tradition of printed " \
                 "Japanese background painting. Illustration, not a photograph and not a " \
                 "photorealistic 3D render.",
          palette: "Warm bone-white panelling and pale birch fittings, muted sage bedding, " \
                   "warm amber morning light, cool blue planet-light as the only cold note.",
          light: "Low warm artificial sunrise from a strip above the window, raking across " \
                 "the desk and throwing soft long shadows to the left; a cool blue fill from " \
                 "the planet below the window line. Visible dust in the warm beam.",
          camera: "Cinematic 16:9, eye level from the doorway, 28mm-equivalent wide angle, " \
                  "deep focus, verticals kept straight.",
          technical: "Keep the lower third quiet and low in contrast for dialogue type. " \
                     "Hold the perspective of the bunk frame and desk edges clean and " \
                     "consistent.",
          exclude: "No people, no characters, no text, no readable lettering on the pinboard " \
                   "notes, no speech bubbles, no logos, no UI elements, no watermark, no " \
                   "photorealistic rendering, no 3D render look, no camera depth-of-field blur."
        )
      )
    end
  end
end
