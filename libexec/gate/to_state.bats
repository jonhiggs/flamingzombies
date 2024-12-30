#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/to_state"

@test "when checking against invalid state" {
  run "${CMD}" "invalid"
  [ "${status}" -eq 1 ]
  [ "${output}" = "received invalid argument" ]
}
