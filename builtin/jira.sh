#!/bin/sh
#
# "jira" builtin command.

function configure_jira() {

  configure_url
  configure_credentials
}

function configure_url() {
  echo "Jira URL (Don't add https://):"
  read url

  define_project
  run_cmd "git config --local $PROJECT_PREFIX.jira.url https://$url/rest/api/2/issue"
}

function configure_credentials() {
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
