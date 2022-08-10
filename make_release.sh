#!/bin/bash

set -ex

kool run make-docs
kool run fmt
kool run lint
kool run test

if [ ! -z "$(git status -s)" ]; then
  echo "You have uncommited changes; aborting creating release."
  exit 1
fi

read -p "What version do you want to build (0.0.0 semver format): "
if [[ ! $REPLY =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]
then
  echo "Bad version format; expected semver 0.0.0"
  exit 1
fi

export BUILD_VERSION=$REPLY

bash build_artifacts.sh

if command -v gh &> /dev/null
then
  read -p "You are going to upload all artifacts to release $BUILD_VERSION. Continue? (y/N) "
  if [[ ! $REPLY =~ ^(yes|YES|y|Y)$ ]]
  then
    exit
  fi

  ARTIFACTS=""
  for artifact in dist/*
  do
    ARTIFACTS="$ARTIFACTS $artifact"
  done

  ARTIFACTS=`echo $ARTIFACTS | sed 's/ *$//g'`

  gh release upload $BUILD_VERSION $ARTIFACTS
fi
