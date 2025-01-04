#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/to_state"

@test "when no argument provided" {
  TASK_STATE=unknown TASK_LAST_STATE=unknown TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}"

  echo ${output}

  [ "${status}" -eq 255 ]
  [ "${output}" = "command did not receive one argument" ]
}

@test "when checking against invalid state" {
  TASK_STATE=unknown TASK_LAST_STATE=unknown TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}" "invalid"

  echo ${output}

  [ "${status}" -eq 255 ]
  [ "${output}" = "received invalid argument" ]
}

@test "when ok from unknown" {
  TASK_STATE=ok TASK_LAST_STATE=unknown TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}" "ok"

  [ "${status}" -eq 1 ]
}

@test "when fail from unknown" {
  TASK_STATE=fail TASK_LAST_STATE=unknown TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}" "fail"

  [ "${status}" -eq 1 ]
}

@test "when ok from fail" {
  TASK_STATE=ok TASK_LAST_STATE=fail TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run bash -x "${CMD}" "ok"

  [ "${status}" -eq 0 ]
}

@test "when fail from ok" {
  TASK_STATE=fail TASK_LAST_STATE=ok TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}" "fail"

  [ "${status}" -eq 0 ]
}

@test "when ok from ok" {
  TASK_STATE=ok TASK_LAST_STATE=ok TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}" "ok"

  [ "${status}" -eq 1 ]
}
