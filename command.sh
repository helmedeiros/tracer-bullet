#!/bin/sh
#
# commands gateway.
source $(dirname $0)/config/constants.sh
source $(dirname $0)/builtin/story.sh

function cmd_story_commits(){
  echo "Listing $PROJECT_PREFIX-$2 Commits";
  story_commits $2 $3;
}
