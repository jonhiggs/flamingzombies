# check that the enviornment contains variable supplied in $1
fz_check_env() {
  if env | grep -q "^$1=.*"; then
    return 0
  else
    echo "$1 is undefined" >&2
    return 1
  fi
}
