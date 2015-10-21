#!/bin/sh
#
source $(dirname $0)/help.sh
source $(dirname $0)/version.sh
source $(dirname $0)/command.sh

function handle_options(){
  # The user didn't specify a command; give them help
  if [ $# = 0 ]; then
    list_commands;
  else
  while [ $# -gt 0 ]; do
    arg=$1;
    case $arg in
      "configure" ) configure_options $@;
      break;;
      "commit" ) commit "$@";
      break;;
      "pair" ) pairing_options $@;
      break;;
      "story:files"		) story_files_options $@;
      break;;
      "story:commits" ) story_commits $@;
      break;;
      "story:diary"   ) story_diary $@;
      break;;
      "story:diff"   ) story_diff $2 $3;
      break;;
      "--version"     ) version;
      break;;
      "--help" | *    ) list_commands;
      break;;
    esac
  done
fi
}

define_constants;
handle_options $@;
