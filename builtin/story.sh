#!/bin/sh
#
# "story" builtin command.
source $(dirname $0)/config/constants.sh

function stories_after_hash() {
  define_project
  echo "Listing stories played after commit: $2";
  git log --pretty=format:"%s %C(yellow)%h%Creset" --date=short $2.. | awk -F'[: ]' '{print $1}' | grep "^$PROJECT_PREFIX" | sort -u
}

function story_commits() {
  define_project
  echo "Listing $PROJECT_PREFIX-$2 Commits";

  git log --pretty=format:"%C(yellow)%ad%Creset %C(green)%an%Creset %s %C(yellow)%h%Creset" --date=short --grep="$PROJECT_PREFIX"-"$2" | sort -u
}

function story_by() {
  echo "Listing story played by: $2 in the past 10 months";
    git log --since 10.months --no-merges --pretty=format:"%s -- %an" | grep "$2" | awk '{print $1}' | awk -F'-' '{print $2}' | awk '!x[$0]++' | awk -F':' '{print $1}' | sort -n
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
  define_project

  echo "Listing missing tests $PROJECT_PREFIX-$1 Files";

  allfiles=$(git log --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u )
  deletedFiles=$(git log --diff-filter=D --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u )

  while read -r file; do
    if [ `echo $deletedFiles | grep -c "$file" ` -le 0 ]; then
      okLines+=$(printf '\n %s \n' "$file")
    fi
  done <<< "$allfiles"

  find_missing_tests_for "$okLines" $2
}

function find_missing_tests_for() {
  notTestFiles=$(echo "$1" | grep ".java" | grep -v "$2" | xargs -n  1 basename | rev | cut -f 2- -d '.' | rev)
  testFiles=$(echo "$1" | sort -u | grep "$2" | xargs -n 1 basename)

  while read -r notTestfile; do
    if [ `echo $testFiles | grep -c "$notTestfile" ` -le 0 ]; then
      notTestedFiles+=$(printf '\n %s \n' "$notTestfile")
    fi
  done <<< "$notTestFiles"

  echo "$notTestedFiles"
  files_summary "$notTestedFiles"
}

function story_files() {
  define_project

  echo "Listing $PROJECT_PREFIX-$1 Files";

  allfiles=$(git log --diff-filter=ACMRTUXB --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u | grep "$2")
  deletedFiles=$(git log --diff-filter=D --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u )

  while read -r file; do
    if [ `echo $deletedFiles | grep -c "$file" ` -le 0 ]; then
      okLines+=$(printf '\n %s \n' "$file")
    fi
  done <<< "$allfiles"


  echo "$okLines"

  files_summary "$okLines"
}

function story_diary() {
  define_project

  case "$2" in
    -t|--today)
      story_number=$3;
      since="midnight";
    ;;
    -y|--yesterday)
      story_number=$3;
      since="yesterday";
    ;;
    -d|--days)
      story_number=$4;
      since="$3.days";
    ;;
     *)
        story_number=$2;
        since="midnight";
      ;;
  esac

  echo "Logging todays commits into $PROJECT_PREFIX-$story_number"
  logs=$(git log --since "$since" --grep "$PROJECT_PREFIX-$story_number" --pretty=format:'%cd : %s' --date=local | perl -p -e 's/\n/\\n/')

  comment_in_jira $story_number "$logs"
}

function comment_in_jira() {
  define_project
  logs=$2
  jira_key=$(git config --local $PROJECT_PREFIX.jira.key)
  jira_url=$(git config --local $PROJECT_PREFIX.jira.url)

  if [ ! -z "$logs" -a "$logs" != " " ]; then
    json="{\"update\": {\"comment\": [{\"add\": {\"body\": \"${logs}\" }}]}}"
    curl -s -D- -X PUT --data "$json" -H "Authorization: Basic ${jira_key}" -H "Content-Type: application/json" "$jira_url/$PROJECT_PREFIX-$1" > /dev/null
    echo "DONE"

  else
    echo "No COMMITS to be logged $PROJECT_PREFIX-$1"
  fi
}

function story_diff(){
  define_project
  allfiles=$(git log --diff-filter=ACMRTUXB --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u | grep "$2")
  deletedFiles=$(git log --diff-filter=D --oneline --grep="$PROJECT_PREFIX"-"$1" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u )

  while read -r file; do
    if [ `echo $deletedFiles | grep -c "$file" ` -le 0 ]; then
      okLines+=$(printf '\n %s \n' "$file")
    fi
  done <<< "$allfiles"

  for i in $okLines
    do
      story_diff_file $1 $i
    done

  files_summary "$okLines"
}

function story_diff_file(){
  define_project
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
