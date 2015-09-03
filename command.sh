#!/bin/sh
#
# commands gateway.
source $(dirname $0)/config/constants.sh
source $(dirname $0)/builtin/configure.sh
source $(dirname $0)/builtin/pair.sh
source $(dirname $0)/builtin/story.sh

function run_cmd(){
  if [ $# -eq 1 ]; then
    echo "Running: $1";
  fi

  $1;

  if [ "$?" -ne "0" ]; then
    echo "command failed: $1";
    exit 1;
  fi
}
