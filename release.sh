#! /usr/bin/env sh

GITHUB_USER=eyedeekay
GITHUB_REPO=onramp
GITHUB_NAME="Removes unnecessary dependency"
GITHUB_DESCRIPTION=$(cat DESC.md)
GITHUB_TAG=0.0.5

github-release release --user "${GITHUB_USER}" \
    --repo "${GITHUB_REPO}" \
    --name "${GITHUB_NAME}" \
    --description "${GITHUB_DESCRIPTION}" \
    --tag "${GITHUB_TAG}"