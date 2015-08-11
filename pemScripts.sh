#!/bin/sh
#
source $(dirname $0)/help.sh

function handle_options(){
  # The user didn't specify a command; give them help
  if [ $# = 0 ]; then
    list_commands;
  fi
}

handle_options $@;
