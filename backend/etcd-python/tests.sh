#!/usr/bin/env bash
set -e

if ! [[ "$0" =~ "./tests.sh" ]]; then
  echo "must be run from repository root"
  exit 255
fi

if [[ -z "${ETCD_EXEC}" ]]; then
  echo ETCD_EXEC is not defined!
  exit 255
fi

pushd ..
ETCD_EXEC=${ETCD_EXEC} python -m unittest etcd-python.etcd_test
popd
