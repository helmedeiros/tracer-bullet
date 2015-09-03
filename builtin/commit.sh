#!/bin/sh
#
# "commit" builtin command.
function commit() {
  TEAM=$PROJECT_PREFIX
  STORY=$(git config --global $PROJECT_PREFIX.current.story)
  MY_USER=$(git config --global $PROJECT_PREFIX.user)
  PAIR_USER=$(git config --global $PROJECT_PREFIX.current.pair)

  PAIR="@$MY_USER"

  if [[ ! -z "$PAIR_USER" ]]; then
    PAIR="@$MY_USER, @$PAIR_USER"
  fi

  git commit -m "$TEAM-$STORY: $1 ($PAIR)"
}
