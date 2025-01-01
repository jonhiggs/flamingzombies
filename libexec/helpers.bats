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

@test "fz_bytes_to_mb: 1048576" {
  run fz_bytes_to_mb 1048576
  [ "${output}" = "1 MB" ]
}

@test "fz_bps_to_string: 300" {
  run fz_bps_to_string 300
  echo "${output}"
  [ "${output}" = "300 b" ]
}

@test "fz_bps_to_string: 1000" {
  run fz_bps_to_string 1000
  echo "${output}"
  [ "${output}" = "1 Kb" ]
}

@test "fz_bps_to_string: 1000000" {
  run fz_bps_to_string 1000000
  echo "${output}"
  [ "${output}" = "1 Mb" ]
}

@test "fz_bps_to_string: 10000000" {
  run fz_bps_to_string 10000000
  echo "${output}"
  [ "${output}" = "10 Mb" ]
}

@test "fz_bps_to_string: 100000000" {
  run fz_bps_to_string 100000000
  echo "${output}"
  [ "${output}" = "100 Mb" ]
}


@test "fz_bps_to_string: 1000000000" {
  run fz_bps_to_string 1000000000
  echo "${output}"
  [ "${output}" = "1 Gb" ]
}

@test "fz_bps_to_string: 2000000000" {
  run fz_bps_to_string 2000000000
  [ "${output}" = "2 Gb" ]
}

@test "fz_bps_to_string: 10000000000" {
  run fz_bps_to_string 10000000000
  [ "${output}" = "10 Gb" ]
}
