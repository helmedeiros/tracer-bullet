#!/bin/sh
#
# "commit" builtin command.
source $(dirname $0)/config/constants.sh

function commit() {
  define_project
  commit_message="${@:2}"

  TEAM=$PROJECT_PREFIX
  STORY=$(git config --local $PROJECT_PREFIX.current.story)
  MY_USER=$(git config --local $PROJECT_PREFIX.user)
  PAIR_USER=$(git config --local $PROJECT_PREFIX.current.pair)

  PAIR="@$MY_USER"

  if [[ ! -z "$PAIR_USER" ]]; then
    PAIR="@$MY_USER, @$PAIR_USER"
  fi

  git commit -m "$TEAM-$STORY: $commit_message ($PAIR)"
}
