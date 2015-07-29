function story_files() {
  story_number=$1

  git log --oneline --grep="$story_number" --name-only | grep -Eo "\w+/.*\.\w+" | sort -u
}

function story_commits(){
  git log --pretty=format:"%C(yellow)%ad%Creset %C(green)%an%Creset %s %C(yellow)%h%Creset" --date=short --grep= "$1" | sort -u
}
