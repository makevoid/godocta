class_name WorldTheme
extends RefCounted
## Sky, ground and lighting for each world. Generated builds pick theirs by title;
## presets and other files use the dark studio.

const STUDIO := "studio"

## Theme name for a structure title: "Cruise Ship" → "cruise_ship", or the studio.
static func name_for(title: String) -> String:
	var key := title.to_snake_case()
	return key if themes().has(key) else STUDIO

static func themes() -> Dictionary:
	return {
		STUDIO: {
			"zenith": Color("0a0e11"), "horizon": Color("111a1f"), "below": Color("0a0e11"),
			"ground": Color("0a0e11"), "grid": 1.0, "lit": 0.3,
			"sun": Color("e7f4e9"), "sun_energy": 1.0, "elevation": 48.0, "azimuth": -35.0,
			"fill": Color("8cbdcd"), "fill_energy": 0.45, "ambient": Color("b0ced7"), "ambient_energy": 0.55,
		},
		# A clear midday over meadows and low green hills.
		"castle": {
			"zenith": Color("3a70b4"), "horizon": Color("bed6e4"), "below": Color("a7bfae"),
			"glow": Color("fff1d2"), "glow_strength": 0.3, "sun_size": 0.025, "sun_strength": 5.0,
			"clouds": 0.5, "cloud_color": Color("ffffff"), "cloud_scale": 1.0,
			"silhouette": Color("6d8c7b"), "silhouette_height": 0.05, "silhouette_seed": 3.0, "haze": 0.45,
			"ground": Color("44582f"), "mottle": 0.3,
			"sun": Color("fff3df"), "sun_energy": 1.05, "elevation": 50.0, "azimuth": -35.0,
			"fill": Color("9cc4e0"), "fill_energy": 0.4, "ambient": Color("c4d8e8"), "ambient_energy": 0.5,
		},
		# Misty dawn: a low sun behind blue mountain ranges.
		"medieval_tower": {
			"zenith": Color("48608f"), "horizon": Color("efc39b"), "below": Color("b4a69a"),
			"glow": Color("ffad6b"), "glow_strength": 0.85, "sun_size": 0.035, "sun_strength": 4.0,
			"clouds": 0.32, "cloud_color": Color("f6cfbf"), "cloud_scale": 0.8,
			"silhouette": Color("5b6687"), "silhouette_height": 0.13, "silhouette_seed": 11.0, "haze": 0.55,
			"ground": Color("3f5236"), "mottle": 0.35,
			"sun": Color("ffd2a3"), "sun_energy": 0.95, "elevation": 16.0, "azimuth": -20.0,
			"fill": Color("8196c8"), "fill_energy": 0.45, "ambient": Color("bcc0d8"), "ambient_energy": 0.55,
		},
		# Mediterranean golden hour over dry earth and olive hills.
		"arena": {
			"zenith": Color("2f69a6"), "horizon": Color("ecd6b0"), "below": Color("bfb194"),
			"glow": Color("ffc078"), "glow_strength": 0.7, "sun_size": 0.03, "sun_strength": 4.5,
			"clouds": 0.2, "cloud_color": Color("fff0da"), "cloud_scale": 1.2,
			"silhouette": Color("8c8a68"), "silhouette_height": 0.045, "silhouette_seed": 7.0, "haze": 0.5,
			"ground": Color("6b6248"), "mottle": 0.25,
			"sun": Color("ffdaa8"), "sun_energy": 1.0, "elevation": 26.0, "azimuth": -50.0,
			"fill": Color("8fb2d6"), "fill_energy": 0.4, "ambient": Color("d6c9b4"), "ambient_energy": 0.5,
		},
		# Open ocean under fair-weather clouds; the hull sits in the water.
		"cruise_ship": {
			"zenith": Color("2864ab"), "horizon": Color("cfe5f1"), "below": Color("9fc2d8"),
			"glow": Color("ffffff"), "glow_strength": 0.3, "sun_size": 0.025, "sun_strength": 5.0,
			"clouds": 0.45, "cloud_color": Color("ffffff"), "cloud_scale": 1.3,
			"ground": Color("1a4363"), "water": 1.0, "water_level": 1.6,
			"sun": Color("fff6e6"), "sun_energy": 1.05, "elevation": 42.0, "azimuth": -40.0,
			"fill": Color("9cc8e8"), "fill_energy": 0.45, "ambient": Color("bcd8ea"), "ambient_energy": 0.55,
		},
		# Blue hour over the harbor: a setting sun, first stars and a distant skyline.
		"manhattan": {
			"zenith": Color("111a37"), "horizon": Color("e58a5c"), "below": Color("2b3450"),
			"glow": Color("ff9a5c"), "glow_strength": 1.0, "sun_size": 0.03, "sun_strength": 3.5,
			"clouds": 0.25, "cloud_color": Color("c78b7f"), "cloud_scale": 1.1, "stars": 0.55,
			"silhouette": Color("1b2135"), "silhouette_height": 0.03, "silhouette_city": 1.0, "silhouette_seed": 5.0, "haze": 0.25,
			"ground": Color("17273b"), "water": 1.0, "water_level": 0.25,
			"sun": Color("ffb98f"), "sun_energy": 0.9, "elevation": 18.0, "azimuth": -60.0,
			"fill": Color("6f82c0"), "fill_energy": 0.5, "ambient": Color("8f9cc8"), "ambient_energy": 0.6,
		},
	}
