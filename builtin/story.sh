#!/bin/sh
#
# "story" builtin command.
function story_commits() {
  echo "Listing $PROJECT_PREFIX-$2 Commits";

  git log --pretty=format:"%C(yellow)%ad%Creset %C(green)%an%Creset %s %C(yellow)%h%Creset" --date=short --grep="$PROJECT_PREFIX"-"$2" | sort -u
}

function story_files_options() {
  case "$2" in
    -mt|--missing-tests)
      missing_tests $3 $TEST_PATTERN
      break;
    ;;

    -t|--tests)
      story_files $3 $TEST_PATTERN
      break;
    ;;

     -s|--sql)
       story_files $3 $SQL_PATTERN
       break;
     ;;

     *)
        story_files $2 ""
        break;
      ;;

  esac
}

function missing_tests() {
  echo "Listing missing tests $PROJECT_PREFIX-$1 Files";

  allfiles=$(git log --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u )
  deletedFiles=$(git log --diff-filter=D --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u )



  while read -r file; do
    if [ `echo $deletedFiles | grep -c "$file" ` -le 0 ]; then
      okLines+=$(printf '\n %s \n' "$file")
    fi
  done <<< "$allfiles"

  notTestFiles=$(echo "$okLines" | grep ".java" | grep -v "$2" | xargs -n 1 basename | rev | cut -f 2- -d '.' | rev)
  testFiles=$(echo "$okLines" | sort -u | grep "$2" | xargs -n 1 basename)

  while read -r notTestfile; do
    if [ `echo $testFiles | grep -c "$notTestfile" ` -le 0 ]; then
      printf '%s is missing\n' "$notTestfile"
    fi
  done <<< "$notTestFiles"
}

function story_files() {
  echo "Listing $PROJECT_PREFIX-$1 Files";

  files=$(git log --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u | grep "$2")

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
  files=$(git log --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u | grep "$2")
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
  totalTests=`echo "$1" | grep $TEST_PATTERN | wc -l`
  totalSQL=`echo "$1" | grep $SQL_PATTERN | wc -l`

  echo "--------------------------------------------------"
  printf "Total: %s | Tests: %s | SQL: %s\n" "$totalFiles" "$totalTests" "$totalSQL"
}
