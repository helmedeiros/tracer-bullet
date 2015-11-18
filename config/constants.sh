#!/usr/bin/env bash
#
function define_constants(){
  define_project;
  TEST_PATTERN="Test.java"
  SQL_PATTERN=".sql"
  BASEDIR=$(dirname $0)
  JIRA_API_ADDRESS="https://jira.com/rest/api/2/issue"
}

function define_project() {
  PROJECT_PREFIX=$(git config --global current.project)
}
