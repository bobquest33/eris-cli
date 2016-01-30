#!/usr/bin/env bash

# ---------------------------------------------------------------------------
# PURPOSE

# This script will test the eris stack and the connection between eris cli
# and eris pm. **Generally, it should not be used in isolation.**

# ---------------------------------------------------------------------------
# REQUIREMENTS

# Docker installed locally
# Docker-Machine installed locally (if using remote boxes)
# eris' test_machines image (if testing against eris' test boxes)
# Eris installed locally

# ---------------------------------------------------------------------------
# USAGE

# test_stack.sh

# ----------------------------------------------------------------------------
# Set definitions and defaults

# Where are the Things
start=`pwd`
base=github.com/eris-ltd/eris-cli
repo=$GOPATH/src/$base
if [ "$CIRCLE_BRANCH" ] # TODO add windows/osx
then
  repo=${GOPATH%%:*}/src/github.com/${CIRCLE_PROJECT_USERNAME}/${CIRCLE_PROJECT_REPONAME}
  ci=true
elif [ "$TRAVIS_BRANCH" ]
then
  ci=true
  osx=true
elif [ "$APPVEYOR_REPO_BRANCH" ]
then
  ci=true
  win=true
else
  ci=false
fi

export ERIS_PULL_APPROVE="true"
export ERIS_MIGRATE_APPROVE="true"

ecm=eris-cm
ecm_repo=https://github.com/eris-ltd/$ecm.git
ecm_dir=$repo/../$ecm
ecm_test_dir=$repo/../$ecm/tests
ecm_branch=${ECM_BRANCH:=master}

epm=eris-pm
epm_repo=https://github.com/eris-ltd/$epm.git
epm_dir=$repo/../$epm
epm_test_dir=$repo/../$epm/tests
epm_branch=${EPM_BRANCH:=master}

# ----------------------------------------------------------------------------
# Utility functions

check_and_exit() {
  if [ $test_exit -ne 0 ]
  then
    cd $start
    exit $test_exit
  fi
}

# ----------------------------------------------------------------------------
# Get ECM

if [ -d "$ecm_test_dir" ]; then
  echo "eris-cm present on host; not cloning"
  cd $ecm_test_dir
else
  git clone $ecm_repo $ecm_dir 1>/dev/null
  cd $ecm_test_dir 1>/dev/null
  git checkout origin/$ecm_branch &>/dev/null
fi

# ----------------------------------------------------------------------------
# Run ECM tests

./test.sh
test_exit=$?
check_and_exit
cd $start

# ----------------------------------------------------------------------------
# Get EPM

if [ -d "$epm_test_dir" ]; then
  echo "eris-pm present on host; not cloning"
  cd $epm_test_dir
else
  git clone $epm_repo $epm_dir 1>/dev/null
  cd $epm_test_dir 1>/dev/null
  git checkout origin/$epm_branch &>/dev/null
fi

# ----------------------------------------------------------------------------
# Run EPM tests

./test.sh
test_exit=$?
check_and_exit
cd $start

# ----------------------------------------------------------------------------
# Cleanup
cd $start
exit $test_exit
