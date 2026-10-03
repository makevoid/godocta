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
	}
