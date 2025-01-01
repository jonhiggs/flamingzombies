# check that the enviornment contains variable supplied in $1
fz_check_env() {
  if env | grep -q "^$1=.*"; then
    return 0
  else
    echo "$1 is undefined" >&2
    return 1
  fi
}

# convert bytes to a MB string. A value of 1048576 would become "1 MB".
fz_bytes_to_mb() {
  echo "$1" | awk '{ print int($1 / 1024 / 1024) " MB" }'
}

# convert bps to a more familiar string
fz_bps_to_string() {
  echo "$1" | awk '
    ($1>=1000000000) { print $1/1000000000 " Gb"; exit 0 }
    ($1>=   1000000) { print $1/   1000000 " Mb"; exit 0 }
    ($1>=      1000) { print $1/      1000 " Kb"; exit 0 }
    { print $1 " b" }
  '
}
