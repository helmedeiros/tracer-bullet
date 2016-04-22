#!/usr/bin/env bash
#
function define_constants(){
  TEST_PATTERN="Test.java"
  SQL_PATTERN=".sql"
  BASEDIR=$(dirname $0)
}

function define_project() {
  PROJECT_PREFIX=$(git config --local current.project)
}
