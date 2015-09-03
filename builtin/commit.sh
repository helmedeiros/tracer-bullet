#!/bin/sh
#
# "commit" builtin command.
function commit() {
  TEAM=$PROJECT_PREFIX
  STORY=$(git config --global $PROJECT_PREFIX.current.story)
  MY_USER=$(git config --global $PROJECT_PREFIX.user)
  PAIR_USER=$(git config --global $PROJECT_PREFIX.current.pair)

  git commit -m "$TEAM-$STORY: $1 (@$MY_USER, @$PAIR_USER)"
}
