#!/usr/bin/env bash
#
function define_constants(){
  define_project;
  BASEDIR=$(dirname $0)
}

function define_project() {
  PROJECT_PREFIX=$(git config --global current.project)
}
