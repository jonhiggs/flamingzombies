#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/min_priority"

@test "when equal" {
  TASK_PRIORITY="5" run "${CMD}" "5"
  [ "${status}" -eq 0 ]
}

@test "when under" {
  TASK_PRIORITY="5" run "${CMD}" "1"
  [ "${status}" -eq 1 ]
}

@test "when over" {
  TASK_PRIORITY="5" run "${CMD}" "10"
  [ "${status}" -eq 0 ]
}
