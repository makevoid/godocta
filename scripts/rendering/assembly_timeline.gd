class_name AssemblyTimeline
extends RefCounted
## A single clock drives all GPU instances. The duration grows with the linear size of
## the build (cube root of the cell count): presets take four seconds, a 70,000-cell
## castle about fourteen. Late cells settle exactly at the end.

const BASE_DURATION: float = 4.0
const BASE_CELLS: float = 1729.0
const MAX_DURATION: float = 24.0
const BASE_SETTLE: float = 1.2
var duration: float = BASE_DURATION
var settle_time: float = BASE_SETTLE
var release_window: float = BASE_DURATION - BASE_SETTLE
var elapsed: float = BASE_DURATION

static func duration_for(cell_count: int) -> float:
	var scale := pow(maxf(cell_count, 1.0) / BASE_CELLS, 1.0 / 3.0)
	return clampf(BASE_DURATION * scale, BASE_DURATION, MAX_DURATION)

## Size the clock for a build. Each cell's flight lengthens only gently, so a large build
## shows a crisp rising front rather than a cloud of thousands of cells in the air.
func configure(cell_count: int) -> void:
	duration = duration_for(cell_count)
	settle_time = BASE_SETTLE * pow(duration / BASE_DURATION, 0.25)
	release_window = duration - settle_time
	elapsed = duration

func restart() -> void:
	elapsed = 0.0

func advance(delta: float) -> void:
	elapsed = minf(elapsed + maxf(delta, 0.0), duration)

func is_running() -> bool:
	return elapsed < duration

func fraction() -> float:
	return elapsed / duration

func release_time(rank: float) -> float:
	# Inverse quadratic: the release wave decelerates towards its final cells.
	return release_window * (1.0 - sqrt(1.0 - clampf(rank, 0.0, 1.0)))

func cell_progress(time: float, rank: float) -> float:
	var t := clampf((time - release_time(rank)) / settle_time, 0.0, 1.0)
	return 1.0 - pow(1.0 - t, 3.0)
