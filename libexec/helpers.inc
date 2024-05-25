metric_gauge() {
  [ "${STATSD_HOST}" = "" ] && return 1
  [ "${STATSD_PREFIX}" = "" ] && return 1

  name=$1

  if [ $# -eq 2 ]; then
    value=$2
  else
    value=$(timeout 0.2 cat)
  fi

  echo "${STATSD_PREFIX}.${name}:${value}|g|${STATSD_TAGS}" \
    | nc -w0 -u "$(echo "${STATSD_HOST}" | cut -d: -f1)" "$(echo "${STATSD_HOST}" | cut -d: -f2)"
}
