#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/is_state"

@test "when matching ok" {
  TASK_STATE="ok" run "${CMD}" "ok"
  [ "${status}" -eq 0 ]
}

@test "when not matching ok" {
  TASK_STATE="unknown" run "${CMD}" "ok"
  [ "${status}" -eq 1 ]
}
