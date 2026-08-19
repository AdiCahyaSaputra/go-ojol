#!/bin/bash
set -euo pipefail

DATA_DIR="${OSRM_DATA_DIR:-/data}"
REGION="${OSRM_REGION:-indonesia-latest}"
PROFILE="${OSRM_PROFILE:-car}"
PBF_URL="${OSRM_PBF_URL:-https://download.geofabrik.de/asia/indonesia-latest.osm.pbf}"
MODE="${OSRM_MODE:-prepare}"
MAX_TABLE_SIZE="${OSRM_MAX_TABLE_SIZE:-10000}"

PBF="${DATA_DIR}/${REGION}.osm.pbf"
OSRM_BASE="${DATA_DIR}/${REGION}"
PROFILE_LUA="/opt/${PROFILE}.lua"
READY="${OSRM_BASE}.osrm.mldgr"

mkdir -p "$DATA_DIR"

download_pbf() {
	if [[ -f "$PBF" ]]; then
		echo "PBF already present: $PBF"
		return
	fi

	echo "Downloading Indonesia extract from Geofabrik (about 1GB)..."
	curl -fL --retry 3 --retry-delay 5 -o "${PBF}.partial" "$PBF_URL"
	mv "${PBF}.partial" "$PBF"
	echo "Download complete."
}

clear_graph() {
	echo "Removing OSRM graph files for ${REGION}..."
	find "$DATA_DIR" -maxdepth 1 -type f -name "${REGION}.osrm*" -delete
}

prepare_graph() {
	if [[ ! -f "$PROFILE_LUA" ]]; then
		echo "Profile not found: $PROFILE_LUA" >&2
		exit 1
	fi

	if [[ ! -f "$PBF" ]]; then
		echo "Missing PBF: $PBF" >&2
		exit 1
	fi

	if [[ "${OSRM_FORCE_REPREPARE:-0}" == "1" ]]; then
		clear_graph
	fi

	if [[ -f "$READY" ]]; then
		echo "MLD graph already built: $READY"
		return
	fi

	clear_graph

	echo "osrm-extract with ${PROFILE} profile. Indonesia often takes 30-60 min and wants around 16GB RAM."
	osrm-extract -p "$PROFILE_LUA" "$PBF"

	echo "osrm-partition..."
	osrm-partition "${OSRM_BASE}.osrm"

	echo "osrm-customize..."
	osrm-customize "${OSRM_BASE}.osrm"

	if [[ ! -f "$READY" ]]; then
		echo "Prepare finished but ${READY} is missing." >&2
		exit 1
	fi

	echo "Graph ready."
}

serve() {
	if [[ ! -f "$READY" ]]; then
		echo "Graph not ready. Run the prepare service first." >&2
		exit 1
	fi

	echo "osrm-routed mld on :5000 (${REGION}, ${PROFILE})"
	exec osrm-routed \
		--algorithm mld \
		--ip 0.0.0.0 \
		--port 5000 \
		--max-table-size "$MAX_TABLE_SIZE" \
		"${OSRM_BASE}.osrm"
}

case "$MODE" in
prepare)
	download_pbf
	prepare_graph
	;;
routed)
	serve
	;;
*)
	echo "Unknown OSRM_MODE=${MODE} (prepare|routed)" >&2
	exit 1
	;;
esac
