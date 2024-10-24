#!/usr/bin/env bash

# check that the files mtime matches the documents .Dd value

for f in $(find . -type f -iname \*\.[0-9]); do
  mtime=$(gdate --date="@$(stat -c %Y $f)" +%Y-%m-%d)
  Dd=$(awk '($1=/^\.Dd /) { print $2 }' $f)
  if [[ "${mtime}" != "${Dd}" ]]; then
    echo "$f: invalid mtime"
    code=1
  fi
done

exit ${code:-0}
