DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/defer"

@test "unknown" {
  TASK_STATE=unknown TASK_LAST_FAIL=0 TASK_LAST_OK=0 \
    run "${CMD}" "30"

  [ "${status}" -eq 1 ]
}

@test "fail: 10 seconds ago, defer 30 seconds" {
  now="$(date +%s)"

  TASK_STATE=fail TASK_LAST_FAIL=0 TASK_LAST_OK=$(( now - 10)) \
    run "${CMD}" "30"

  [ "${status}" -eq 1 ]
}

@test "fail: 60 seconds ago, defer 30 seconds" {
  now=$(date +%s)

  TASK_STATE=fail TASK_LAST_FAIL=0 TASK_LAST_OK=$((now-60)) \
    run "${CMD}" "30"

  [ "${status}" -eq 0 ]
}

@test "ok: 10 seconds ago, defer 30 seconds" {
  now="$(date +%s)"

  TASK_STATE=ok TASK_LAST_FAIL=$((now-10)) TASK_LAST_OK=0
    run "${CMD}" "30"

  [ "${status}" -eq 1 ]
}

@test "ok: 60 seconds ago, defer 30 seconds" {
  now=$(date +%s)

  TASK_STATE=ok TASK_LAST_FAIL=$((now-60)) TASK_LAST_OK=0 \
    run "${CMD}" "30"

  [ "${status}" -eq 0 ]
}
