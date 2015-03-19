scd() {
	d=$(srcsearch "$@")
	st=$?
	if [[ $st -ne 0 ]]; then
		return $st
	fi
	printf "%s\n" "$d"
	cd "$d"
}
