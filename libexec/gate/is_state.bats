#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/is_state"

@test "when matching" {
  TASK_STATE="ok" run "${CMD}" "ok"
  [ "${status}" -eq 0 ]
}

@test "when not matching" {
  TASK_STATE="unknown" run "${CMD}" "ok"
  [ "${status}" -eq 1 ]
}