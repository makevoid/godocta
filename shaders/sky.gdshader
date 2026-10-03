shader_type sky;
render_mode use_debanding;
// Procedural sky dome for the world themes: gradient, sun glow and disk (from the key
// light), drifting-free fbm clouds, a distant horizon silhouette and optional stars.
// Static (no TIME), so Godot only re-renders its radiance when parameters change.

uniform vec4 zenith_color : source_color = vec4(0.04, 0.055, 0.067, 1.0);
uniform vec4 horizon_color : source_color = vec4(0.07, 0.1, 0.12, 1.0);
uniform vec4 below_color : source_color = vec4(0.04, 0.055, 0.067, 1.0);
uniform vec4 glow_color : source_color = vec4(1.0, 0.85, 0.6, 1.0);
uniform float glow_strength = 0.0;
uniform float sun_size = 0.03;
uniform float sun_strength = 0.0;
uniform vec4 cloud_color : source_color = vec4(1.0);
uniform float cloud_cover = 0.0;
uniform float cloud_scale = 1.0;
uniform vec4 silhouette_color : source_color = vec4(0.3, 0.35, 0.4, 1.0);
uniform float silhouette_height = 0.0; // Peak elevation, as sin(angle).
uniform float silhouette_city = 0.0; // 0 = rolling hills, 1 = skyline blocks.
uniform float silhouette_seed = 0.0;
uniform float haze = 0.4; // How much the silhouette fades into the horizon.
uniform float stars = 0.0;

float hash(vec2 p) {
	return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

float noise(vec2 p) {
	vec2 i = floor(p);
	vec2 f = fract(p);
	vec2 u = f * f * (3.0 - 2.0 * f);
	return mix(mix(hash(i), hash(i + vec2(1.0, 0.0)), u.x),
		mix(hash(i + vec2(0.0, 1.0)), hash(i + vec2(1.0, 1.0)), u.x), u.y);
}

float fbm(vec2 p) {
	float value = 0.0;
	float amplitude = 0.5;
	for (int i = 0; i < 5; i++) {
		value += amplitude * noise(p);
		p = p * 2.03 + vec2(17.0, 9.0);
		amplitude *= 0.5;
	}
	return value;
}

// Silhouette profile around the horizon, 0..1. Sampled on a circle so it wraps seamlessly.
float ridge(vec3 dir) {
	float azimuth = atan(dir.x, dir.z);
	vec2 around = vec2(cos(azimuth), sin(azimuth));
	if (silhouette_city > 0.5) {
		float lot = floor((azimuth / TAU + 0.5) * 360.0);
		float district = smoothstep(0.35, 0.75, fbm(around * 2.2 + silhouette_seed));
		return (0.15 + 0.85 * pow(hash(vec2(lot, silhouette_seed)), 2.5)) * district;
	}
	return fbm(around * 1.7 + silhouette_seed) * 0.75 + fbm(around * 7.0 + silhouette_seed) * 0.35;
}

void sky() {
	vec3 dir = EYEDIR;
	float y = dir.y;
	vec3 sun = LIGHT0_ENABLED ? LIGHT0_DIRECTION : vec3(0.0, 1.0, 0.0);
	float toward_sun = max(dot(dir, sun), 0.0);
	float height = pow(clamp(y, 0.0, 1.0), 0.45);
	vec3 color = mix(horizon_color.rgb, zenith_color.rgb, height);
	// Warm glow along the horizon on the sun's side.
	color += glow_color.rgb * glow_strength * pow(toward_sun, 5.0) * (1.0 - height);
	if (stars > 0.0) {
		vec3 cell = floor(dir * 220.0);
		float star = step(0.9978, hash(cell.xy + cell.z * 7.31));
		color += vec3(star * stars * smoothstep(0.08, 0.5, y));
	}
	// Sun disk and halo.
	if (sun_strength > 0.0 && LIGHT0_ENABLED) {
		float disk = smoothstep(cos(sun_size * 1.2), cos(sun_size), dot(dir, sun));
		color += LIGHT0_COLOR * (disk * sun_strength + pow(toward_sun, 300.0) * sun_strength * 0.15);
	}
	// Clouds projected onto a layer above: denser toward the horizon, lit toward the sun.
	if (cloud_cover > 0.0 && y > 0.0) {
		vec2 uv = dir.xz / (y + 0.1) * 0.7 * cloud_scale;
		float n = fbm(uv + vec2(3.1, 7.7));
		float density = smoothstep(1.0 - cloud_cover, 1.25 - cloud_cover, n);
		float lit = 0.78 + 0.22 * pow(toward_sun, 2.0) + 0.2 * (n - 0.5);
		color = mix(color, cloud_color.rgb * lit, density * smoothstep(0.0, 0.2, y) * 0.92);
	}
	if (y < 0.0) {
		color = mix(horizon_color.rgb, below_color.rgb, smoothstep(0.0, 0.04, -y));
	}
	// Distant hills or skyline, fading into the haze at their base.
	// Below the horizon the ground plane normally covers the sky; its far edge fades to
	// below_color, so silhouettes stand only above the horizon.
	if (silhouette_height > 0.0 && y >= 0.0) {
		float top = ridge(dir) * silhouette_height;
		if (y < top) {
			float depth = clamp(y / max(top, 0.0001), 0.0, 1.0);
			vec3 land = mix(mix(below_color.rgb, horizon_color.rgb, haze), silhouette_color.rgb, 0.45 + 0.55 * depth);
			if (silhouette_city > 0.5) {
				// A few lit windows.
				vec2 window = floor(vec2(atan(dir.x, dir.z) * 900.0, y * 900.0));
				land += vec3(1.0, 0.78, 0.45) * step(0.975, hash(window)) * 0.6 * stars;
			}
			color = land;
		}
	}
	COLOR = color;
}
