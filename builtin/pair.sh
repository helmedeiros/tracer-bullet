#!/bin/sh
#
# "pair" builtin command.
function pairing_options() {
  case "$2" in
    -a|--alone)
      not_pairing
      break;
    ;;

     -s|--story)
       pairing_on_story $3
       break;
     ;;

     -w|--with|*)
        pairing_with $3
        break;
      ;;

  esac
}

function not_pairing() {
  run_cmd "git config --global --unset $PROJECT_PREFIX.current.pair"
}

function pairing_on_story() {
  run_cmd "git config --global $PROJECT_PREFIX.current.story $1"
}

function pairing_with() {
  run_cmd "git config --global $PROJECT_PREFIX.current.pair $1"
}
