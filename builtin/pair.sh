#!/bin/sh
#
# "pair" builtin command.
source $(dirname $0)/config/constants.sh

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
  define_project
  run_cmd "git config --local --unset $PROJECT_PREFIX.current.pair"
}

function pairing_on_story() {
  define_project
  run_cmd "git config --local $PROJECT_PREFIX.current.story $1"
}

function pairing_with() {
  define_project
  run_cmd "git config --local $PROJECT_PREFIX.current.pair $1"
}
