#!/bin/sh
#
# "story" builtin command.
function story_commits() {
  echo "Listing $PROJECT_PREFIX-$2 Commits";

  git log --pretty=format:"%C(yellow)%ad%Creset %C(green)%an%Creset %s %C(yellow)%h%Creset" --date=short --grep="$PROJECT_PREFIX"-"$2" | sort -u
}

function story_files() {
  echo "Listing $PROJECT_PREFIX-$1 Files";

  git log --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u
}
