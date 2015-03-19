scd() {
	d=$(srcsearch "$@")
	st=$?
	if [[ $st -ne 0 ]]; then
		return $st
	fi
	cd "$d"
}
