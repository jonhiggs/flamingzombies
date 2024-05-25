metric_gauge() {
  name=$1
  value=$2

  echo "${STATSD_PREFIX}.${name}:${value}|g|${STATSD_TAGS}" \
    | nc -w0 -u "$(echo "${STATSD_HOST}" | cut -d: -f1)" "$(echo "${STATSD_HOST}" | cut -d: -f2)"
}
