#!/bin/sh
#
# "story" builtin command.
function story_commits() {
  git log --pretty=format:"%C(yellow)%ad%Creset %C(green)%an%Creset %s %C(yellow)%h%Creset" --date=short --grep="$PROJECT_PREFIX"-"$1" | sort -u
}
