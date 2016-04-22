#!/bin/sh
#
# "jira" builtin command.

function configure_jira() {
  echo User:
  read user

  echo Password:
  read -s password

  write_base64_jira_key `printf "$user:$password" | openssl enc -base64 -A`
}

function write_base64_jira_key() {
  define_project
  run_cmd "git config --local $PROJECT_PREFIX.jira.key $1"
}
