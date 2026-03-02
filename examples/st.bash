#!/bin/bash

BOLD=
OFFBOLD=
RESET_COLOR=
RED=
GREEN=
YELLOW=
BLUE=
BLUE_CYAN=

if [ -t 1 ]; then
    BOLD=$(tput bold)
    OFFBOLD=$(tput sgr0)
    RESET_COLOR="$(tput sgr0)"
    RED="$(
        tput bold
        tput setaf 1
    )"
    GREEN="$(
        tput bold
        tput setaf 2
    )"
    YELLOW="$(
        tput bold
        tput setaf 3
    )"
    BLUE="$(
        tput bold
        tput setaf 4
    )"
    BLUE_CYAN="$(
        tput bold
        tput setaf 6
    )"
fi

DOING_MSG=

function st.h1() {
    echo -e "st.h1> ${BOLD}$1${OFFBOLD}"
}

function st.h2() {
    echo -e "st.h2> ${BOLD}$1${OFFBOLD}"
}

function st.h3() {
    echo -e "st.h3> ${BOLD}$1${OFFBOLD}"
}

function st.doing() {
    DOING_MSG=$1

    echo "st.doing> ${BLUE} Doing « ${DOING_MSG:-} »…$RESET_COLOR"
}

function st.done() {
    local DONE="${1:-DONE}"

    echo "st.done> ${DOING_MSG:-} : ${GREEN}$DONE${RESET_COLOR}"
}

function st.nothingTodo() {
    echo "${GREEN}st.nothingtd> Nothing to do…${RESET_COLOR}"
}

function st.skipped() {
    echo "st.skiped> ${DOING_MSG:-} : ${BLUE_CYAN}SKIPPED${RESET_COLOR}"
}

function st.warn() {
    echo "st.warn> ${DOING_MSG:-} : ${BOLD}${YELLOW}$1${RESET_COLOR}${OFFBOLD}"
}

function st.fail() {
    echo -e "st.fail> ${BOLD}${RED}$1\nPROCESS ABORTED${RESET_COLOR}${OFFBOLD}\n"

    exit 1
}

function st.do() {
    # https://unix.stackexchange.com/questions/148109/shifting-command-output-to-the-right
    local -a cmd=("$@")
    echo "st.do> ${BLUE_CYAN}${cmd[@]}${RESET_COLOR}"

    "${cmd[@]}" || st.fail "Command failed: ${cmd[*]}"
}
