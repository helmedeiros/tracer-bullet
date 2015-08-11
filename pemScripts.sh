#!/bin/sh
#
source $(dirname $0)/help.sh

function handle_options(){
  # The user didn't specify a command; give them help
  if [ $# = 0 ]; then
    list_commands;
  else
  while [ $# -gt 0 ]; do
    arg=$1;
    case $arg in
      "--help" | *) list_commands;
      break;;
    esac
  done
fi
}

handle_options $@;
