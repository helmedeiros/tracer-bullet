function story_files() {
  story_number=$1

  git log --oneline --grep="$story_number" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u
}

function story_commits() {
  git log --pretty=format:"%C(yellow)%ad%Creset %C(green)%an%Creset %s %C(yellow)%h%Creset" --date=short --grep= "$1" | sort -u
}

function story_diff() {
  first_log=$(git log --pretty=format:%h --grep="$1" | tail -n 1)
  last_log=$(git log --pretty=format:%h --reverse --grep="$1" | tail -n 1)
  git difftool -y $first_log $last_log
}
