#!/usr/bin/env bats
DIR=$(dirname "${BATS_TEST_FILENAME}")
source ${DIR}/helpers.inc

@test "fz_check_env: when exists" {
  EXISTS="yes" run fz_check_env EXISTS
  echo ${output}
  [ "${status}" -eq 0 ]
}

@test "fz_check_env: when doesn't exist" {
  run fz_check_env EXISTS
  [ "${output}" = "EXISTS is undefined" ]
  [ "${status}" -eq 1 ]
}
