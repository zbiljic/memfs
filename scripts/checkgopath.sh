#!/usr/bin/env bash

_init() {
    PACKAGE="github.com/zbiljic/memfs"

    shopt -s extglob

    # Fetch real paths instead of symlinks before comparing them
    PWD=$(env pwd -P)
    GOPATH=$(cd "$(go env GOPATH)" || exit ; env pwd -P)
}

main() {
    echo "Checking if project is at ${GOPATH}"
    for path in $(echo "${GOPATH}" | tr ':' ' '); do
        if [ ! -d "${path}/src/${PACKAGE}" ]; then
            echo "Project not found in ${path}." \
                && exit 1
        fi
        if [ "x${path}/src/${PACKAGE}" != "x${PWD}" ]; then
            echo "Build outside of ${path}, two source checkouts found. Exiting." && exit 1
        fi
    done
}

_init && main
