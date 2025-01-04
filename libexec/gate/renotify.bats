#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
CMD="${DIR}/renotify"

now=$(date +%s)

@test "when unstable" {
  TASK_STATE=fail TASK_LAST_STATE=unknown TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=$((now-3600)) TASK_LAST_OK=0 \
    run "${CMD}" 300

  [ "${status}" -eq 1 ]
}

@test "when transitioning" {
  TASK_STATE=fail TASK_LAST_STATE=ok TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=$((now-3600)) TASK_LAST_OK=0 \
    run "${CMD}" 300

  [ "${status}" -eq 1 ]
}


@test "when long after previous notification" {
  # it's one hour after the daemon started
  # there are 5 minute renotifications

  TASK_STATE=fail TASK_LAST_STATE=fail TASK_LAST_NOTIFICATION=0 TASK_LAST_FAIL=$((now-3600)) TASK_LAST_OK=0 \
    run "${CMD}" 300

  [ "${status}" -eq 0 ]
}


@test "when shortly after a notification" {
  TASK_STATE=fail TASK_LAST_STATE=fail TASK_LAST_NOTIFICATION=$((now-2)) TASK_LAST_FAIL=$((now-3600)) TASK_LAST_OK=0 \
    run "${CMD}" 300

  [ "${status}" -eq 1 ]
}
