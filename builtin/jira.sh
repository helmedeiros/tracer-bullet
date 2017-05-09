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

function get_issue_from_jira() {
  story=$1
  jira_key=$(git config --local $PROJECT_PREFIX.jira.key)
  jira_url=$(git config --local $PROJECT_PREFIX.jira.url)
  url=$(echo $jira_url | cut -d'/' -f3)

  issue=$(curl -s -D- -X GET -H "Authorization: Basic ${jira_key}" "Content-Type: application/json" "$jira_url/$story?fields=summary,customfield_10073,status")
  title=$(jsonval "$issue" "summary" | awk -F'|' '{print $3}')
  dev=$(jsonval "$issue" "emailAddress" | awk -F'|' '{print $2}')
  status=$(jsonval "$issue" "name" | awk -F'|' '{if ($2 == "Backlog" || $2 == "In Progress" || $2 == "Read for Validation" || $2 == "Validation" || $2 == "Signoff" || $2 == "done" || $2 == "Done" ) print $2}')

  echo "$story - $title - https://$url/browse/$story ($status) ($dev)"
}


function jsonval {
  echo $1 | sed 's/\\\\\//\//g' | sed 's/[{}]//g' | awk -v k="text" '{n=split($0,a,","); for (i=1; i<=n; i++) print a[i]}' | sed 's/\"\:\"/\|/g' | sed 's/[\,]/ /g' | sed 's/\"//g' | grep -w $2 | sort -u
}
