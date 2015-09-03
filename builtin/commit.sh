#!/bin/sh
#
# "commit" builtin command.
function commit() {
  TEAM=$PROJECT_PREFIX
  STORY=$(git config --global $PROJECT_PREFIX.current.story)
  PAIR=$(git config --global $PROJECT_PREFIX.current.pair)

  git commit -m "$TEAM-$STORY: $1 ($PAIR)"
}
