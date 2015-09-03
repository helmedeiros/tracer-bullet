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
  echo "    configure [--autocomplete, --jira, --project, --user]  -- optional configurations";
  echo "    commit -- easy commit for configured projects";
  echo "    pair [--alone, --with, --story]  -- Configure pairing informations";
  echo "    story:commits  -- list all commits for a story #";
  echo "    story:files  -- list all files modified for a story #";
  echo "    story:diary -- add to jira as coments the commit from current day for a story #";
  echo "    story:diff  -- diff all changed files for a story #";
}
