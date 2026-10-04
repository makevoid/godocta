class_name WorldTheme
extends RefCounted
## Sky, ground and lighting for each world. Generated builds pick theirs by title;
## presets and other files use the dark studio.

const STUDIO := "studio"

## Theme name for a structure title: "Cruise Ship" → "cruise_ship", or the studio.
static func name_for(title: String) -> String:
	var key := title.to_snake_case()
	return key if themes().has(key) else STUDIO

## Shared by every world: around 5–6 pm under thin cloud. A low, veiled warm sun with
## soft, light shadows, and a bright warm ambient so faces stay evenly lit.
const LATE_AFTERNOON := {
	"glow": Color("ffb36e"), "glow_strength": 0.55, "sun_size": 0.045, "sun_strength": 1.1,
	"clouds": 0.72, "cloud_color": Color("e9d8c9"), "cloud_scale": 1.0,
	"sun": Color("ffcf9e"), "sun_energy": 0.55, "elevation": 24.0,
	"shadow_opacity": 0.5, "shadow_blur": 2.5,
	"fill": Color("d7cdd0"), "fill_energy": 0.35,
	"ambient": Color("f3dfc8"), "ambient_energy": 0.95,
}

static func _late_afternoon(world: Dictionary) -> Dictionary:
	var theme := LATE_AFTERNOON.duplicate()
	theme.merge(world, true)
	return theme

static func themes() -> Dictionary:
	return {
		STUDIO: {
			"zenith": Color("0a0e11"), "horizon": Color("111a1f"), "below": Color("0a0e11"),
			"ground": Color("0a0e11"), "grid": 1.0, "lit": 0.3,
			"sun": Color("e7f4e9"), "sun_energy": 0.75, "elevation": 48.0, "azimuth": -35.0,
			"shadow_opacity": 0.6, "shadow_blur": 2.0,
			"fill": Color("8cbdcd"), "fill_energy": 0.35, "ambient": Color("b8ced6"), "ambient_energy": 0.62,
		},
		# Meadows and low green hills.
		"castle": _late_afternoon({
			"zenith": Color("6f86a6"), "horizon": Color("e8d4bd"), "below": Color("b3b39a"),
			"silhouette": Color("7c8a74"), "silhouette_height": 0.05, "silhouette_seed": 3.0, "haze": 0.5,
			"ground": Color("44582f"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Blue-grey mountain ranges in the haze.
		"medieval_tower": _late_afternoon({
			"zenith": Color("68789e"), "horizon": Color("ecc9a8"), "below": Color("b8aa9c"),
			"silhouette": Color("6c7390"), "silhouette_height": 0.13, "silhouette_seed": 11.0, "haze": 0.55,
			"ground": Color("475a3a"), "mottle": 0.35, "azimuth": -20.0,
		}),
		# Mediterranean dry earth and olive hills.
		"arena": _late_afternoon({
			"zenith": Color("6c84a8"), "horizon": Color("ecd3b2"), "below": Color("c0b296"),
			"silhouette": Color("8f8a6c"), "silhouette_height": 0.045, "silhouette_seed": 7.0, "haze": 0.5,
			"ground": Color("6e6449"), "mottle": 0.25, "azimuth": -50.0,
		}),
		# Open ocean; the hull sits in the water.
		"cruise_ship": _late_afternoon({
			"zenith": Color("5f80aa"), "horizon": Color("e6d6c6"), "below": Color("a9bccb"),
			"cloud_scale": 1.3,
			"ground": Color("1d4a62"), "water": 1.0, "water_level": 1.6, "azimuth": -40.0,
		}),
		# The harbor with a hazy skyline; a few windows already lit.
		"manhattan": _late_afternoon({
			"zenith": Color("5d6d93"), "horizon": Color("eabf9a"), "below": Color("8c93a3"),
			"cloud_color": Color("dcc6bb"), "city_lights": 0.3,
			"silhouette": Color("59607a"), "silhouette_height": 0.03, "silhouette_city": 1.0, "silhouette_seed": 5.0, "haze": 0.35,
			"ground": Color("1f3d4f"), "water": 1.0, "water_level": 0.25, "azimuth": -60.0,
		}),
		# Paris parkland with a low city skyline in the haze.
		"eiffel_tower": _late_afternoon({
			"zenith": Color("6a84aa"), "horizon": Color("efd2b0"), "below": Color("b9b29c"),
			"silhouette": Color("7d8296"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 9.0, "haze": 0.45,
			"ground": Color("4a5c36"), "mottle": 0.3, "azimuth": -45.0,
		}),
		# Red Square: paved city plaza, a low Moscow skyline in the haze.
		"st_basils_cathedral": _late_afternoon({
			"zenith": Color("6a82aa"), "horizon": Color("ead2b6"), "below": Color("a8a499"),
			"silhouette": Color("6f7488"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 9.0, "haze": 0.45,
			"ground": Color("55524c"), "mottle": 0.2, "azimuth": -40.0,
		}),
		# Hazy Agra afternoon over the dusty Yamuna plain.
		"taj_mahal": _late_afternoon({
			"zenith": Color("7189ab"), "horizon": Color("efd2b0"), "below": Color("c4b294"),
			"silhouette": Color("8d8a70"), "silhouette_height": 0.035, "silhouette_seed": 13.0, "haze": 0.6,
			"ground": Color("6b6646"), "mottle": 0.3, "azimuth": -30.0,
		}),
		# Open Caribbean sea; the hull sits in the water.
		"pirate_galleon": _late_afternoon({
			"zenith": Color("5a7fae"), "horizon": Color("ecd5bd"), "below": Color("a3bfcc"),
			"cloud_scale": 1.2,
			"ground": Color("1b5468"), "water": 1.0, "water_level": 2.4, "azimuth": -30.0,
		}),
		# Grassy plain with low wooded hills; the moat is part of the build.
		"japanese_castle": _late_afternoon({
			"zenith": Color("6a82a8"), "horizon": Color("ebd2bc"), "below": Color("b5b49c"),
			"silhouette": Color("6f7f6c"), "silhouette_height": 0.07, "silhouette_seed": 13.0, "haze": 0.5,
			"ground": Color("4a5e34"), "mottle": 0.3, "azimuth": -30.0,
		}),
		# Atlantic coast; the sea layer is part of the build and runs on into the theme ocean.
		"lighthouse_cove": _late_afternoon({
			"zenith": Color("5c7fac"), "horizon": Color("ead6c2"), "below": Color("a6bccb"),
			"cloud_scale": 1.2, "haze": 0.4,
			"ground": Color("1e5670"), "water": 1.0, "water_level": 0.6, "azimuth": -35.0,
		}),
		# Île de la Cité: the Seine around the island, a low Paris skyline in the haze.
		"gothic_cathedral": _late_afternoon({
			"zenith": Color("6a84aa"), "horizon": Color("efd4b4"), "below": Color("a7b4b8"),
			"silhouette": Color("7d8296"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 11.0, "haze": 0.45,
			"ground": Color("2a5a6e"), "water": 1.0, "water_level": 0.35, "azimuth": -45.0,
		}),
		# Alpine meadow below blue Bavarian mountains.
		"neuschwanstein": _late_afternoon({
			"zenith": Color("6584ad"), "horizon": Color("ecd3b8"), "below": Color("b3b59e"),
			"silhouette": Color("6a7a96"), "silhouette_height": 0.16, "silhouette_seed": 18.0, "haze": 0.5,
			"ground": Color("44592f"), "mottle": 0.35, "azimuth": -35.0,
		}),
		# Hazy London afternoon on the Thames, a low city skyline beyond.
		"tower_bridge": _late_afternoon({
			"zenith": Color("6680a8"), "horizon": Color("e8cfb4"), "below": Color("9aa3ad"),
			"cloud_color": Color("d8c8bf"),
			"silhouette": Color("6c7286"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 7.0, "haze": 0.5,
			"ground": Color("2a4a5c"), "water": 1.0, "water_level": 0.6, "azimuth": -40.0,
		}),
		# Tuscan meadow with low ochre hills and a faint city skyline in the haze.
		"piazza_dei_miracoli": _late_afternoon({
			"zenith": Color("6b86ac"), "horizon": Color("efd6b4"), "below": Color("b8b49a"),
			"silhouette": Color("8a8c78"), "silhouette_height": 0.04, "silhouette_city": 1.0, "silhouette_seed": 13.0, "haze": 0.45,
			"ground": Color("4b6034"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Beijing parkland in a hazy afternoon, a low city skyline beyond the cypresses.
		"temple_of_heaven": _late_afternoon({
			"zenith": Color("6a82a6"), "horizon": Color("e8d0b4"), "below": Color("b0ad98"),
			"silhouette": Color("767c8c"), "silhouette_height": 0.03, "silhouette_city": 1.0, "silhouette_seed": 21.0, "haze": 0.55,
			"ground": Color("475a33"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Desert floor under a high sky, distant red mesas in the haze.
		"wild_west_town": _late_afternoon({
			"zenith": Color("5f86b8"), "horizon": Color("f0d2a8"), "below": Color("c9a882"),
			"silhouette": Color("b0674a"), "silhouette_height": 0.06, "silhouette_seed": 17.0, "haze": 0.45,
			"ground": Color("b98f5e"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Provençal garrigue: dry olive-green hills in a warm haze above the Gardon gorge.
		"pont_du_gard": _late_afternoon({
			"zenith": Color("6886ad"), "horizon": Color("efd5b2"), "below": Color("b9b49a"),
			"silhouette": Color("7f8a6e"), "silhouette_height": 0.06, "silhouette_seed": 17.0, "haze": 0.5,
			"ground": Color("56603c"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Florida coastal scrub at Cape Canaveral; a flat, hazy Atlantic horizon.
		"saturn_v_launch": _late_afternoon({
			"zenith": Color("5f82ae"), "horizon": Color("ecd6bc"), "below": Color("b2b49a"),
			"cloud_scale": 1.2,
			"silhouette": Color("7f8c7a"), "silhouette_height": 0.02, "silhouette_seed": 21.0, "haze": 0.45,
			"ground": Color("4d6136"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Aegean caldera: deep blue sea and a hazy line of island ridges.
		"santorini": _late_afternoon({
			"zenith": Color("5a80b0"), "horizon": Color("efd6bc"), "below": Color("9fbccc"),
			"silhouette": Color("7d7f8c"), "silhouette_height": 0.05, "silhouette_seed": 21.0, "haze": 0.4,
			"ground": Color("1d5873"), "water": 1.0, "water_level": 0.6, "azimuth": -35.0,
		}),
		# Open Pacific swell; the hull sits in the water.
		"battleship": _late_afternoon({
			"zenith": Color("5c7ca8"), "horizon": Color("e4d6c8"), "below": Color("a2b6c6"),
			"cloud_scale": 1.3, "haze": 0.35,
			"ground": Color("1c4a64"), "water": 1.0, "water_level": 4.6, "azimuth": -35.0,
		}),
		# High in the sky: towering clouds, hazy hills and a pale meadow far below.
		"floating_islands": _late_afternoon({
			"zenith": Color("5f84b4"), "horizon": Color("efd9c2"), "below": Color("c9d3dc"),
			"cloud_scale": 1.4, "clouds": 0.85,
			"silhouette": Color("8d9ab8"), "silhouette_height": 0.06, "silhouette_seed": 17.0, "haze": 0.6,
			"ground": Color("6f8a5c"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Thessalian plain under the Pindus foothills; hazy ridges behind the pillars.
		"meteora": _late_afternoon({
			"zenith": Color("6a84aa"), "horizon": Color("ecd2b2"), "below": Color("b6b098"),
			"silhouette": Color("737c8c"), "silhouette_height": 0.1, "silhouette_seed": 17.0, "haze": 0.55,
			"ground": Color("4f5e38"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Humid Cambodian plain: hazy jungle tree line, golden west light on the façade.
		"angkor_wat": _late_afternoon({
			"zenith": Color("6b86ab"), "horizon": Color("f0cfa4"), "below": Color("b2b08f"),
			"silhouette": Color("5f7458"), "silhouette_height": 0.04, "silhouette_seed": 17.0, "haze": 0.6,
			"ground": Color("3f5a2e"), "mottle": 0.35, "azimuth": -25.0,
		}),
		# Paved Old Town square with a low Prague skyline of roofs and spires in the haze.
		"prague_old_town": _late_afternoon({
			"zenith": Color("6a82a8"), "horizon": Color("ecd2b6"), "below": Color("aaa59a"),
			"silhouette": Color("737a8e"), "silhouette_height": 0.04, "silhouette_city": 1.0, "silhouette_seed": 17.0, "haze": 0.45,
			"ground": Color("57534c"), "mottle": 0.2, "azimuth": -40.0,
		}),
		# The Golden Gate on a clear afternoon: Pacific swell, hazy hills of the bay beyond.
		"golden_gate_bridge": _late_afternoon({
			"zenith": Color("5f81ad"), "horizon": Color("ecd4ba"), "below": Color("a2b8c6"),
			"cloud_scale": 1.2,
			"silhouette": Color("6f7d86"), "silhouette_height": 0.06, "silhouette_seed": 21.0, "haze": 0.5,
			"ground": Color("1e5068"), "water": 1.0, "water_level": 1.1, "azimuth": -35.0,
		}),
		# Lunar surface under a deep navy "space" sky; low hard sun, grey regolith.
		"lunar_landing": _late_afternoon({
			"zenith": Color("0d1630"), "horizon": Color("2b3654"), "below": Color("6d6b66"),
			"glow": Color("c8d2f0"), "glow_strength": 0.2, "clouds": 0.0,
			"sun": Color("fff3e2"), "sun_energy": 0.8, "elevation": 20.0,
			"ambient": Color("c9d0e0"), "ambient_energy": 0.8,
			"silhouette": Color("3e4048"), "silhouette_height": 0.03, "silhouette_seed": 7.0, "haze": 0.12,
			"ground": Color("8a877f"), "mottle": 0.35, "azimuth": -35.0,
		}),
		# Green meadow under a warm sky, soft rolling hills in the haze.
		"hillside_burrows": _late_afternoon({
			"zenith": Color("6a88b0"), "horizon": Color("f0d6b4"), "below": Color("b5b896"),
			"silhouette": Color("7d9478"), "silhouette_height": 0.06, "silhouette_seed": 13.0, "haze": 0.5,
			"ground": Color("4f6b34"), "mottle": 0.35, "azimuth": -30.0,
		}),
		# Bay at low tide: the sea creeping over the sands, a low Normandy coast in the haze.
		"mont_saint_michel": _late_afternoon({
			"zenith": Color("6680a8"), "horizon": Color("ecd6bc"), "below": Color("b4b8b0"),
			"silhouette": Color("7f8a80"), "silhouette_height": 0.025, "silhouette_seed": 23.0, "haze": 0.55,
			"ground": Color("3f6f86"), "water": 1.0, "water_level": 1.1, "azimuth": -40.0,
		}),
		# Beijing haze: warm late light over a grey city skyline beyond the moat.
		"forbidden_city": _late_afternoon({
			"zenith": Color("6c84a6"), "horizon": Color("ead2b6"), "below": Color("aaa694"),
			"silhouette": Color("7a7c86"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 9.0, "haze": 0.6,
			"ground": Color("5a6048"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Sydney Harbour on a clear afternoon: blue water around the point, a low city skyline in the haze.
		"sydney_opera_house": _late_afternoon({
			"zenith": Color("5d82b0"), "horizon": Color("ecd5ba"), "below": Color("a3bccb"),
			"cloud_scale": 1.2,
			"silhouette": Color("6f7a8c"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 19.0, "haze": 0.45,
			"ground": Color("1d5873"), "water": 1.0, "water_level": 0.35, "azimuth": -40.0,
		}),
		# Open meadow with low wooded hills; a bright, festive afternoon.
		"funfair": _late_afternoon({
			"zenith": Color("6a86ae"), "horizon": Color("efd6b8"), "below": Color("b6b69c"),
			"silhouette": Color("72846e"), "silhouette_height": 0.05, "silhouette_seed": 17.0, "haze": 0.45,
			"ground": Color("4a6034"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Sultanahmet: paved plaza and park, a hazy Istanbul skyline beyond.
		"hagia_sophia": _late_afternoon({
			"zenith": Color("6b84ab"), "horizon": Color("edd1b0"), "below": Color("b0a896"),
			"silhouette": Color("737a8e"), "silhouette_height": 0.04, "silhouette_city": 1.0, "silhouette_seed": 17.0, "haze": 0.5,
			"ground": Color("5a5b45"), "mottle": 0.25, "azimuth": -40.0,
		}),
		# Andean cloud forest: misty high valley, blue-green ranges behind the ridge.
		"machu_picchu": _late_afternoon({
			"zenith": Color("6787ae"), "horizon": Color("e6d6bf"), "below": Color("a9b29a"),
			"silhouette": Color("5d7068"), "silhouette_height": 0.22, "silhouette_seed": 23.0, "haze": 0.6,
			"ground": Color("3f5a33"), "mottle": 0.35, "azimuth": -30.0,
		}),
		# Venetian lagoon; the built canal runs on into the theme water, a low hazy skyline beyond.
		"venice_grand_canal": _late_afternoon({
			"zenith": Color("6a84aa"), "horizon": Color("efd2b0"), "below": Color("a9b8bc"),
			"silhouette": Color("8a8a96"), "silhouette_height": 0.03, "silhouette_city": 1.0, "silhouette_seed": 21.0, "haze": 0.5,
			"ground": Color("2c5f6e"), "water": 1.0, "water_level": 1.1, "azimuth": -40.0,
		}),
		# Barcelona's Eixample: warm Mediterranean haze over a low city skyline.
		"sagrada_familia": _late_afternoon({
			"zenith": Color("6886ae"), "horizon": Color("efd6b4"), "below": Color("b4ab9a"),
			"silhouette": Color("7a7f90"), "silhouette_height": 0.04, "silhouette_city": 1.0, "silhouette_seed": 17.0, "haze": 0.45,
			"ground": Color("5a564e"), "mottle": 0.2, "azimuth": -40.0,
		}),
		# New York Harbor: open water with a hazy Manhattan skyline beyond.
		"statue_of_liberty": _late_afternoon({
			"zenith": Color("5f7fab"), "horizon": Color("ebd2b8"), "below": Color("9fb2c0"),
			"cloud_color": Color("dcc8bc"),
			"silhouette": Color("5f6680"), "silhouette_height": 0.04, "silhouette_city": 1.0, "silhouette_seed": 5.0, "haze": 0.4,
			"ground": Color("1f4a60"), "water": 1.0, "water_level": 0.8, "azimuth": -40.0,
		}),
		# Rose sandstone desert of Wadi Musa: warm dusty haze over red ranges.
		"petra_treasury": _late_afternoon({
			"zenith": Color("6585b0"), "horizon": Color("f0d0aa"), "below": Color("c9a585"),
			"silhouette": Color("a8664c"), "silhouette_height": 0.08, "silhouette_seed": 19.0, "haze": 0.5,
			"ground": Color("c49a6c"), "mottle": 0.3, "azimuth": -25.0,
		}),
		# Dusky highland loch below blue mountains; lantern-lit castle on its cliff.
		"wizard_academy": _late_afternoon({
			"zenith": Color("4f6790"), "horizon": Color("e3c3a6"), "below": Color("8d98a6"),
			"cloud_color": Color("d6c2bb"),
			"silhouette": Color("56627e"), "silhouette_height": 0.18, "silhouette_seed": 23.0, "haze": 0.55,
			"ground": Color("1d3a4c"), "water": 1.0, "water_level": 0.6, "azimuth": -35.0,
		}),
		# Norwegian fjord under a veiled sky; blue mountain ranges beyond the theme water.
		"viking_harbor": _late_afternoon({
			"zenith": Color("5a6f90"), "horizon": Color("d9cbbd"), "below": Color("9aa8b0"),
			"cloud_color": Color("d4cbc6"), "clouds": 0.85,
			"silhouette": Color("5c6a80"), "silhouette_height": 0.16, "silhouette_seed": 21.0, "haze": 0.55,
			"ground": Color("1d4658"), "water": 1.0, "water_level": 0.8, "azimuth": -30.0,
		}),
		# Westminster on the Thames under a hazy London afternoon, a low city skyline beyond.
		"big_ben": _late_afternoon({
			"zenith": Color("6680a8"), "horizon": Color("e9d0b6"), "below": Color("9aa3ad"),
			"cloud_color": Color("d8c8bf"),
			"silhouette": Color("6c7286"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 9.0, "haze": 0.5,
			"ground": Color("2a4a5c"), "water": 1.0, "water_level": 0.6, "azimuth": -40.0,
		}),
		# Open ocean under a hazy late-afternoon sky.
		"aircraft_carrier": _late_afternoon({
			"zenith": Color("5c7ca8"), "horizon": Color("e4d6c8"), "below": Color("a2b6c6"),
			"cloud_scale": 1.3, "haze": 0.35,
			"ground": Color("1c4a64"), "water": 1.0, "water_level": 3.8, "azimuth": -35.0,
		}),
		# Deep forest glade: warm low sun filtering through a distant wall of old trees.
		"treehouse_village": _late_afternoon({
			"zenith": Color("6886ac"), "horizon": Color("ecd6b0"), "below": Color("aab694"),
			"silhouette": Color("4f6a48"), "silhouette_height": 0.14, "silhouette_seed": 17.0, "haze": 0.55,
			"ground": Color("3f5a30"), "mottle": 0.4, "azimuth": -30.0,
		}),
		# Giza plateau: warm desert haze over the dunes, low sandstone ridges on the horizon.
		"giza_pyramids": _late_afternoon({
			"zenith": Color("6787b2"), "horizon": Color("f2d3a6"), "below": Color("cdb088"),
			"silhouette": Color("b38a62"), "silhouette_height": 0.04, "silhouette_seed": 23.0, "haze": 0.55,
			"ground": Color("c9a774"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Flat Yucatán jungle under a humid, hazy sky; low green tree line on the horizon.
		"chichen_itza": _late_afternoon({
			"zenith": Color("6684a8"), "horizon": Color("e9d3b4"), "below": Color("aab49a"),
			"silhouette": Color("4f6a48"), "silhouette_height": 0.03, "silhouette_seed": 17.0, "haze": 0.55,
			"ground": Color("3f5a2e"), "mottle": 0.35, "azimuth": -35.0,
		}),
		# Lhasa valley: thin clear high-altitude air, bare brown mountain ranges in the haze.
		"potala_palace": _late_afternoon({
			"zenith": Color("4f74ad"), "horizon": Color("e4d6c0"), "below": Color("b4ab95"),
			"silhouette": Color("7d7a86"), "silhouette_height": 0.12, "silhouette_seed": 17.0, "haze": 0.35,
			"ground": Color("7a6d4c"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Breezy hilltop sky-dock above rolling green downs, warm late light.
		"steampunk_airship": _late_afternoon({
			"zenith": Color("6886ae"), "horizon": Color("efd3b0"), "below": Color("b6b59c"),
			"silhouette": Color("72806e"), "silhouette_height": 0.06, "silhouette_seed": 21.0, "haze": 0.5,
			"ground": Color("475c33"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Hazy Attic afternoon: dry olive hills and Hymettus beyond the city.
		"parthenon": _late_afternoon({
			"zenith": Color("6a86ad"), "horizon": Color("efd6b4"), "below": Color("c2b496"),
			"silhouette": Color("8a8a7c"), "silhouette_height": 0.06, "silhouette_seed": 17.0, "haze": 0.5,
			"ground": Color("7a6e52"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Hazy autumn mountain ridges north of Beijing.
		"great_wall": _late_afternoon({
			"zenith": Color("6a80a6"), "horizon": Color("ecd0b0"), "below": Color("b2ab98"),
			"silhouette": Color("6d7a8c"), "silhouette_height": 0.14, "silhouette_seed": 17.0, "haze": 0.6,
			"ground": Color("475a36"), "mottle": 0.35, "azimuth": -30.0,
		}),
		# Adriatic coast: clear blue sea under a bright late sky, karst hills inland.
		"dubrovnik": _late_afternoon({
			"zenith": Color("4f7fb4"), "horizon": Color("ecd8bf"), "below": Color("9cbccb"),
			"silhouette": Color("737a72"), "silhouette_height": 0.07, "silhouette_seed": 43.0, "haze": 0.4,
			"ground": Color("1b5a78"), "water": 1.0, "water_level": 0.6, "azimuth": -35.0,
		}),
		# Kedu plain parkland: humid tropical haze, green volcanic hills on the horizon.
		"borobudur": _late_afternoon({
			"zenith": Color("6886aa"), "horizon": Color("ecd4b2"), "below": Color("a9b393"),
			"silhouette": Color("5d7560"), "silhouette_height": 0.08, "silhouette_seed": 23.0, "haze": 0.6,
			"ground": Color("44602f"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Dusty Martian afternoon: butterscotch sky, pale haze, low rust-red mesas on the horizon.
		"mars_base": _late_afternoon({
			"zenith": Color("8a6a55"), "horizon": Color("e8b98c"), "below": Color("b9744a"),
			"glow": Color("f2d2b0"), "glow_strength": 0.25, "clouds": 0.0,
			"sun": Color("ffe6c8"), "sun_energy": 0.9,
			"silhouette": Color("9a5a3c"), "silhouette_height": 0.05, "silhouette_seed": 19.0, "haze": 0.5,
			"ground": Color("a65a36"), "mottle": 0.35, "azimuth": -35.0,
		}),
		# Highland glen: misty blue-grey mountain ranges beyond the hills.
		"steam_train": _late_afternoon({
			"zenith": Color("6a7fa3"), "horizon": Color("e6d2b8"), "below": Color("a9ac98"),
			"silhouette": Color("66738c"), "silhouette_height": 0.12, "silhouette_seed": 17.0, "haze": 0.55,
			"ground": Color("4a5e36"), "mottle": 0.35, "azimuth": -30.0,
		}),
		# Guanabara Bay haze with green Tijuca ridges; the cove meets the sea.
		"christ_the_redeemer": _late_afternoon({
			"zenith": Color("5f80ad"), "horizon": Color("efd3b4"), "below": Color("9fb8c4"),
			"silhouette": Color("5d7a6a"), "silhouette_height": 0.1, "silhouette_seed": 17.0, "haze": 0.45,
			"ground": Color("1d4f63"), "water": 1.0, "water_level": 2.0, "azimuth": -40.0,
		}),
		# Granada: dry olive hills under a hazy Sierra Nevada.
		"alhambra": _late_afternoon({
			"zenith": Color("6b86ac"), "horizon": Color("efd3b2"), "below": Color("bdb096"),
			"silhouette": Color("7f879c"), "silhouette_height": 0.1, "silhouette_seed": 17.0, "haze": 0.55,
			"ground": Color("5d6340"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Tropical Pacific; the sea plate is part of the build and runs on into the theme ocean.
		"volcano_island": _late_afternoon({
			"zenith": Color("5a86b4"), "horizon": Color("f0d8bc"), "below": Color("9fc4cf"),
			"cloud_scale": 1.1, "haze": 0.35,
			"ground": Color("17607a"), "water": 1.0, "water_level": 0.6, "azimuth": -35.0,
		}),
		# South Pacific coast; the build's sea layer runs on into the theme ocean.
		"easter_island": _late_afternoon({
			"zenith": Color("5a7eab"), "horizon": Color("ead5bd"), "below": Color("a3b8c6"),
			"cloud_scale": 1.25, "haze": 0.4,
			"ground": Color("1d5470"), "water": 1.0, "water_level": 0.8, "azimuth": -35.0,
		}),
		# Roman afternoon: warm haze over the Palatine and a low skyline of domes and pines.
		"colosseum_ruins": _late_afternoon({
			"zenith": Color("6b85aa"), "horizon": Color("efd4b0"), "below": Color("bcb398"),
			"silhouette": Color("857f74"), "silhouette_height": 0.04, "silhouette_city": 1.0, "silhouette_seed": 19.0, "haze": 0.5,
			"ground": Color("5f6040"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Gloomy hilltop at dusk: dark meadow, bruised sky and low hazy hills.
		"haunted_manor": _late_afternoon({
			"zenith": Color("4f5d7e"), "horizon": Color("d9b49a"), "below": Color("8f9488"),
			"clouds": 0.85, "cloud_color": Color("cdb9b4"),
			"silhouette": Color("4c5466"), "silhouette_height": 0.08, "silhouette_seed": 23.0, "haze": 0.6,
			"ground": Color("34472c"), "mottle": 0.35, "azimuth": -35.0,
		}),
		# Kyoto hills in autumn: soft haze over wooded ridges.
		"golden_pavilion": _late_afternoon({
			"zenith": Color("7090b8"), "horizon": Color("efd8bc"), "below": Color("aeb58e"),
			"silhouette": Color("56704e"), "silhouette_height": 0.09, "silhouette_seed": 21.0, "haze": 0.45,
			"ground": Color("4f6a34"), "mottle": 0.32, "azimuth": -35.0,
		}),
		# Open ocean under a late-afternoon sky; the sea plate runs on into the theme sea.
		"clipper_ship": _late_afternoon({
			"zenith": Color("5a80b0"), "horizon": Color("ead7c3"), "below": Color("a4bccb"),
			"cloud_scale": 1.3, "clouds": 0.8, "haze": 0.35,
			"ground": Color("1d5370"), "water": 1.0, "water_level": 0.95, "azimuth": -35.0,
		}),
		# Himalayan valley: forested slopes and blue mountain ranges in the haze.
		"tigers_nest": _late_afternoon({
			"zenith": Color("6580a8"), "horizon": Color("e9d3b8"), "below": Color("aeb4a0"),
			"silhouette": Color("66718e"), "silhouette_height": 0.16, "silhouette_seed": 13.0, "haze": 0.55,
			"ground": Color("3f5a34"), "mottle": 0.35, "azimuth": -30.0,
		}),
		# Moscow: paved city with a low skyline in the haze; the river is part of the build.
		"kremlin": _late_afternoon({
			"zenith": Color("6a82aa"), "horizon": Color("ead2b6"), "below": Color("a8a499"),
			"silhouette": Color("6f7488"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 17.0, "haze": 0.45,
			"ground": Color("55524c"), "mottle": 0.2, "azimuth": -40.0,
		}),
		# Amsterdam's canal ring on a soft Dutch afternoon: hazy gabled skyline beyond.
		"amsterdam_canals": _late_afternoon({
			"zenith": Color("6c86ac"), "horizon": Color("e8d4bc"), "below": Color("a6aaa6"),
			"silhouette": Color("7c8290"), "silhouette_height": 0.035, "silhouette_city": 1.0, "silhouette_seed": 23.0, "haze": 0.5,
			"ground": Color("55524c"), "mottle": 0.2, "azimuth": -40.0,
		}),
		# Theban desert at golden hour: pale sky, faint Theban hills beyond the Nile.
		"karnak_temple": _late_afternoon({
			"zenith": Color("6886b0"), "horizon": Color("f1d2a4"), "below": Color("cbad84"),
			"silhouette": Color("b0855c"), "silhouette_height": 0.05, "silhouette_seed": 29.0, "haze": 0.55,
			"ground": Color("cfae78"), "mottle": 0.3, "azimuth": -30.0,
		}),
		# Red-rock desert under a clear high-plateau sky; distant mesas on the horizon.
		"monument_valley": _late_afternoon({
			"zenith": Color("5a83b9"), "horizon": Color("f1d0a4"), "below": Color("c79a72"),
			"silhouette": Color("a75a3c"), "silhouette_height": 0.07, "silhouette_seed": 29.0, "haze": 0.4,
			"ground": Color("c27a4c"), "mottle": 0.3, "azimuth": -40.0,
		}),
		# Babylon on the Euphrates: warm hazy Mesopotamian plain, low tan horizon.
		"ishtar_gate": _late_afternoon({
			"zenith": Color("6684ae"), "horizon": Color("f0d0a2"), "below": Color("c9a97e"),
			"silhouette": Color("a8825c"), "silhouette_height": 0.035, "silhouette_seed": 31.0, "haze": 0.55,
			"ground": Color("cdab76"), "mottle": 0.3, "azimuth": -30.0,
		}),
		# Mesa Verde canyon country: piñon-juniper mesas in a warm haze.
		"cliff_palace": _late_afternoon({
			"zenith": Color("6d88b0"), "horizon": Color("efd0aa"), "below": Color("c2a98a"),
			"silhouette": Color("9a7a5c"), "silhouette_height": 0.06, "silhouette_seed": 13.0, "haze": 0.45,
			"ground": Color("7a6a4a"), "mottle": 0.3, "azimuth": -30.0,
		}),
		# Norwegian fjord below steep green mountains; the fjord layer is part of the build.
		"stave_church": _late_afternoon({
			"zenith": Color("6282aa"), "horizon": Color("e6d3bf"), "below": Color("a9bcc4"),
			"silhouette": Color("5f7590"), "silhouette_height": 0.18, "silhouette_seed": 21.0, "haze": 0.45,
			"ground": Color("24566a"), "water": 1.0, "water_level": 0.9, "azimuth": -35.0,
		}),
		# Franconian farmland and low wooded hills around the Tauber valley.
		"rothenburg": _late_afternoon({
			"zenith": Color("6a84aa"), "horizon": Color("ecd2b4"), "below": Color("b4b39a"),
			"silhouette": Color("72806a"), "silhouette_height": 0.06, "silhouette_seed": 17.0, "haze": 0.5,
			"ground": Color("4a5e34"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# West Highland sea loch under a soft late sky; blue mountain ridges beyond the theme water.
		"eilean_donan": _late_afternoon({
			"zenith": Color("57709a"), "horizon": Color("e6cdb4"), "below": Color("95a3ad"),
			"cloud_color": Color("d8cac2"), "clouds": 0.8,
			"silhouette": Color("5a6882"), "silhouette_height": 0.16, "silhouette_seed": 27.0, "haze": 0.5,
			"ground": Color("1e4a5e"), "water": 1.0, "water_level": 0.6, "azimuth": -35.0,
		}),
		# Overcast North Sea: grey veiled sky, cold steel-blue water to the horizon.
		"oil_rig": _late_afternoon({
			"zenith": Color("5a6e8c"), "horizon": Color("d6d0c8"), "below": Color("98a8b4"),
			"clouds": 0.9, "cloud_color": Color("cfcac6"), "cloud_scale": 1.3, "haze": 0.5,
			"ground": Color("1c4458"), "water": 1.0, "water_level": 0.6, "azimuth": -35.0,
		}),
		# Istanbul at golden hour: warm Marmara haze, low city skyline beyond the park.
		"blue_mosque": _late_afternoon({
			"zenith": Color("5f7fae"), "horizon": Color("f0cfa6"), "below": Color("aeb3a2"),
			"silhouette": Color("6f7890"), "silhouette_height": 0.05, "silhouette_city": 1.0, "silhouette_seed": 29.0, "haze": 0.55,
			"ground": Color("55603f"), "mottle": 0.3, "azimuth": -35.0,
		}),
		# Ligurian coast: warm late light, deep blue sea, hazy green headlands on the horizon.
		"cinque_terre": _late_afternoon({
			"zenith": Color("4f7fb4"), "horizon": Color("efd8bc"), "below": Color("9cbccb"),
			"silhouette": Color("5f7466"), "silhouette_height": 0.09, "silhouette_seed": 57.0, "haze": 0.45,
			"ground": Color("1b5a78"), "water": 1.0, "water_level": 0.5, "azimuth": -35.0,
		}),
		# Flat Dutch polder under a wide sky; low tree lines on the horizon.
		"kinderdijk": _late_afternoon({
			"zenith": Color("6684ab"), "horizon": Color("e9d6bf"), "below": Color("b1b59c"),
			"silhouette": Color("7f8c78"), "silhouette_height": 0.02, "silhouette_seed": 17.0, "haze": 0.5,
			"ground": Color("4a6234"), "mottle": 0.3, "azimuth": -40.0,
		}),
	}
