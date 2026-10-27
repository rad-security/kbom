#!/usr/bin/env bash
# Runs kbom against the current cluster in every format, full and short,
# and validates the output against the kbom and CycloneDX schemas.
#
# Requirements: jq, yq, go (for the jv JSON schema validator) and cyclonedx-cli.
#
# Environment:
#   KBOM          kbom binary (default: ./kbom)
#   KUBE_CONTEXT  Kubernetes context (default: current context)
#   CYCLONEDX     cyclonedx-cli binary (default: cyclonedx)
#   EXPECT_KIND   optional resource kind that must be present in the KBOM
set -euo pipefail

KBOM=${KBOM:-./kbom}
CYCLONEDX=${CYCLONEDX:-cyclonedx}
KBOM_CMD=("$KBOM")
if [[ -n ${KUBE_CONTEXT:-} ]]; then
  KBOM_CMD+=(--context "$KUBE_CONTEXT")
fi
JV=(go run github.com/santhosh-tekuri/jsonschema/cmd/jv@v0.7.0)

OUT=$(mktemp -d)
trap 'rm -rf "$OUT"' EXIT

step() { echo "--- $*"; }
fail() { echo "FAIL: $*" >&2; exit 1; }

check() {
  local file=$1 filter=$2 msg=$3
  jq -e "$filter" "$file" >/dev/null || fail "$msg ($file)"
}

"$KBOM" schema > "$OUT/schema.json"

for mode in full short; do
  flags=(-q)
  if [[ $mode == short ]]; then
    flags+=(--short)
  fi

  for format in json yaml cyclonedx-json cyclonedx-xml; do
    step "generate $mode $format"
    "${KBOM_CMD[@]}" generate "${flags[@]}" -f "$format" > "$OUT/$mode.$format"
  done

  step "validate $mode json and yaml against kbom schema"
  # convert YAML with yq, jv would decode unquoted timestamps as time values
  yq -o json "$OUT/$mode.yaml" > "$OUT/$mode.yaml.json"
  "${JV[@]}" "$OUT/schema.json" "$OUT/$mode.json" "$OUT/$mode.yaml.json"

  step "check $mode json content"
  json=$OUT/$mode.json
  check "$json" '.cluster.name != ""' "cluster name is empty"
  check "$json" '.cluster.id != null and .cluster.id != ""' "cluster id is empty"
  check "$json" '.cluster.k8s_version != ""' "k8s version is empty"
  check "$json" '.cluster.nodes_count > 0 and (.cluster.nodes | length) == .cluster.nodes_count' "nodes are missing"
  check "$json" '(.cluster.components.images | length) > 0' "images are missing"
  check "$json" '(.cluster.components.resources | length) > 0' "resources are missing"
  check "$json" '[.cluster.components.resources[] | .count > 0] | all' "resource counts are missing"
  if [[ $mode == full ]]; then
    check "$json" '[.cluster.components.resources[] | (.resources | length) == .count] | all' "resource items do not match counts"
  else
    check "$json" '[.cluster.components.resources[] | .resources == null] | all' "short output contains resource items"
  fi
  if [[ -n ${EXPECT_KIND:-} ]]; then
    check "$json" "[.cluster.components.resources[] | select(.kind == \"$EXPECT_KIND\")] | length > 0" "kind $EXPECT_KIND is missing"
  fi

  step "validate $mode cyclonedx against CycloneDX schema"
  "$CYCLONEDX" validate --input-file "$OUT/$mode.cyclonedx-json" --input-format json --input-version v1_7 --fail-on-errors
  "$CYCLONEDX" validate --input-file "$OUT/$mode.cyclonedx-xml" --input-format xml --input-version v1_7 --fail-on-errors
done

# Regression test for #129: KUBECONFIG with multiple files
step "generate with multi-file KUBECONFIG"
kubeconfig=${KUBECONFIG:-$HOME/.kube/config}
printf 'apiVersion: v1\nkind: Config\n' > "$OUT/empty-kubeconfig"
KUBECONFIG="$OUT/empty-kubeconfig:$kubeconfig" "${KBOM_CMD[@]}" generate -q --short > "$OUT/multi.json"
check "$OUT/multi.json" '.cluster.nodes_count > 0' "multi-file KUBECONFIG output has no nodes"

echo "PASS"
