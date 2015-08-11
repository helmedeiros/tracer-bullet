#!/bin/sh
#
# helps messages.
scripts_usage_string="$progname [--version] [--help] <command> [<args>]";

function list_commands(){
  printf "usage: %s\n\n" "$scripts_usage_string";

  list_common_cmds_help
}

function list_common_cmds_help(){
  echo "The most commonly used `basename $0` commands are";
  echo "    story:commits  -- list all commits for a story #";
}
