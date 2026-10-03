class_name AssemblyTimeline
extends RefCounted
## A single clock drives all GPU instances. Late cells settle exactly at 4 seconds.

const DURATION: float = 4.0
const RELEASE_WINDOW: float = 2.8
const SETTLE_TIME: float = 1.2
var elapsed: float = DURATION

func restart() -> void:
	elapsed = 0.0

func advance(delta: float) -> void:
	elapsed = minf(elapsed + maxf(delta, 0.0), DURATION)

func is_running() -> bool:
	return elapsed < DURATION

func fraction() -> float:
	return elapsed / DURATION

static func release_time(rank: float) -> float:
	# Inverse quadratic: the release wave decelerates towards its final cells.
	return RELEASE_WINDOW * (1.0 - sqrt(1.0 - clampf(rank, 0.0, 1.0)))

static func cell_progress(time: float, rank: float) -> float:
	var t := clampf((time - release_time(rank)) / SETTLE_TIME, 0.0, 1.0)
	return 1.0 - pow(1.0 - t, 3.0)
