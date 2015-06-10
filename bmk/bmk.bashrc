#!/bin/bash

BMK_PATH="$HOME/.config/bmk/bookmarks"
_bmk_all_names=''
_bmk_all_bookmarks=()

_loadbmks()
{
	IFS=$'\n'
	local lines=($(<"$BMK_PATH"))
	unset IFS
	for line in "${lines[@]}"; do
		local name="${line//,*}"
		_bmk_all_names+="$name "
		_bmk_all_bookmarks+=("$name" "${line#*,}")
	done
}
_loadbmks

_bcd()
{
	local cur="${COMP_WORDS[COMP_CWORD]}"
	COMPREPLY=($(compgen -W "$_bmk_all_names" -- "$cur"))
}

bcd()
{
	if [[ $# -lt 1 ]]; then
		printf 'Must provide bookmark to cd to\n'
		return 1
	fi
	for ((i=${#_bmk_all_bookmarks[@]}-2; i >= 0; i-=2)); do
		if [[ $1 == "${_bmk_all_bookmarks[i]}" ]]; then
			cd "${_bmk_all_bookmarks[i+1]}"
			return
		fi
	done
	printf '%s: No such bookmark\n' "$1"
}

complete -F _bcd bcd

_bmk_cwd()
{
	local cwd=$PWD
	local longest=-1
	for ((i=1; i < ${#_bmk_all_bookmarks[@]}; i+=2)); do
		if [[ $cwd == ${_bmk_all_bookmarks[i]} ||
		      $cwd == ${_bmk_all_bookmarks[i]}/* ]] &&
		   [[ $longest == -1 ||
		      ${#_bmk_all_bookmarks[i]} -gt ${#_bmk_all_bookmarks[longest]} ]]
		then
			longest=$i
		fi
	done
	if [[ $longest == -1 ]]; then
		cwd=$(sed "s:^$HOME:~:" <<<"$cwd")
	else
		cwd=$(sed "s|^${_bmk_all_bookmarks[longest]}|#${_bmk_all_bookmarks[longest-1]}|" <<<"$cwd")
	fi
	printf '%s\n' "$cwd"
}
