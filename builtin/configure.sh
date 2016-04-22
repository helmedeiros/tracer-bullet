#!/bin/sh
#
# "configure" builtin command.
source $(dirname $0)/config/constants.sh
source $(dirname $0)/builtin/jira.sh

function configure_options() {
  case "$2" in

     -a|--autocomplete)
       configure_zsh_autocomplete $@
       break;
     ;;

     -j|--jira)
       configure_jira $@
       break;
     ;;

     -p|--project)
       configure_project $3
       break;
     ;;

     -u|--user)
       configure_user $3
       break;
     ;;

   esac
}

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

function configure_project() {
  run_cmd "git config --local current.project $1"
  define_project
}

function configure_user() {
  define_project
  run_cmd "git config --local $PROJECT_PREFIX.user $1"
}
