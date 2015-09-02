#!/bin/sh
#
# "configure" builtin command.
function configure_zsh_autocomplete() {
  cd

  if grep -Fxq "fpath=($BASEDIR/completion/zsh" .zshrc; then
    echo "AUTO COMPLETE ALREADY CONFIGURED"
  else
    echo "" >> .zshrc
    echo "fpath=($BASEDIR/completion/zsh \$fpath)" >> .zshrc
    echo "autoload -U compinit" >> .zshrc
    echo "compinit" >> .zshrc
  fi
}

function configure_jira() {
  echo User:
  read user

  echo Password:
  read -s password

  write_base64_jira_key `printf "$user:$password" | openssl enc -base64`

}

function write_base64_jira_key() {
  echo "#!/usr/bin/env bash
#
function load_user_data() {
  USER_JIRA_KEY=\""$1"\";
}
"  > $BASEDIR/config/user_data.sh

}
