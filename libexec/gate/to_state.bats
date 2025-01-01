#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/to_state"

@test "when checking against invalid state" {
  TASK_STATE=unknown TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}" "invalid"
  [ "${status}" -eq 1 ]
  [ "${output}" = "received invalid argument" ]
}
