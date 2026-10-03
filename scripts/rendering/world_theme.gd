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
	}
