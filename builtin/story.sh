#!/bin/sh
#
# "story" builtin command.
function story_commits() {
  echo "Listing $PROJECT_PREFIX-$2 Commits";

  git log --pretty=format:"%C(yellow)%ad%Creset %C(green)%an%Creset %s %C(yellow)%h%Creset" --date=short --grep="$PROJECT_PREFIX"-"$2" | sort -u
}

function story_files() {
  echo "Listing $PROJECT_PREFIX-$1 Files";

  files=$(git log --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u)

  echo "$files"

  files_summary "$files"
}

function story_diary() {
  echo "Logging todays commits into $PROJECT_PREFIX-$1"
  logs=$(git log --since yesterday --grep "$PROJECT_PREFIX-$1" --pretty=format:'%cd : %s' --date=local | perl -p -e 's/\n/\\n/')

  comment_in_jira $1 "$logs"
}

function comment_in_jira() {
  logs=$2

  if [ ! -z "$logs" -a "$logs" != " " ]; then
    json="{\"update\": {\"comment\": [{\"add\": {\"body\": \"${logs}\" }}]}}"
    curl -D- -X PUT --data "$json" -H "Authorization: Basic $USER_JIRA_KEY" -H "Content-Type: application/json" "https://jira.com/rest/api/2/issue/$PROJECT_PREFIX-$1"
  else
    echo "No COMMITS to be logged $PROJECT_PREFIX-$1"
  fi
}

function story_diff(){
  files=$(git log --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u)
  for i in $files
    do
      story_diff_file $1 $i
    done
}

function story_diff_file(){
  first_log=$(git log --pretty=format:%h --grep="$PROJECT_PREFIX"-"$1" -- "$2" | tail -n 1)
  last_log=$(git log --pretty=format:%h --reverse --grep="$PROJECT_PREFIX"-"$1" -- "$2" | tail -n 1)
  run_cmd "git difftool -y $last_log $first_log^ -- $2"
}

function files_summary() {

  totalFiles=`echo "$1" | wc -l`
  totalTests=`echo "$1" | grep Test.java | wc -l`
  totalSQL=`echo "$1" | grep .sql | wc -l`

  echo "--------------------------------------------------"
  printf "Total: %s | Tests: %s | SQL: %s\n" "$totalFiles" "$totalTests" "$totalSQL"
}
