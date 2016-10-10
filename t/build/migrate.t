#!/bin/sh
echo 1..1
OUT=$(narada-install --check migrate 2>&1) && echo ok || echo not ok
echo "$OUT" | sed 's/^/# /' >&2
