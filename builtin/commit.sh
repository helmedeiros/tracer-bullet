#!/bin/sh
#
# "commit" builtin command.
source $(dirname $0)/config/constants.sh

function commit() {
  define_project

  TEAM=$PROJECT_PREFIX
  STORY=$(git config --local $PROJECT_PREFIX.current.story)
  MY_USER=$(git config --local $PROJECT_PREFIX.user)
  PAIR_USER=$(git config --local $PROJECT_PREFIX.current.pair)

  PAIR="@$MY_USER"

  if [[ ! -z "$PAIR_USER" ]]; then
    PAIR="@$MY_USER, @$PAIR_USER"
  fi

  echo Title:
  read title

  echo Why?
  read why

  echo How?
  read how

  git commit -m "$TEAM-$STORY: $title ($PAIR)

  Why?
    $why

  How?
    $how
  ";
}
