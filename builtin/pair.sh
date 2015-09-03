#!/bin/sh
#
# "pair" builtin command.
function pairing_options() {
  case "$2" in
     -s|--story)
       pairing_on_story $3
       break;
     ;;
  esac
}

function pairing_on_story() {
  run_cmd "git config --global $PROJECT_PREFIX.current.story $1"
}
