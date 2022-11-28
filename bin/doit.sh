FTA=$HOME/src/fta/cli/build/install/fta/bin/cli

for i in "$@"
do
	if [ -f "$1.out" ]
	then
		exit 0
	fi

	# See if there is a Locale file
	if [ -f "$i.locale" ]
	then
		LOCALE="$(cat "$i".locale)"
		OPTIONS="$OPTIONS --locale $LOCALE"
	fi

	# Skip any file tagged to be ignored - multiple reasons, too many columns, columns too wide, ...
	if [ -f "$i.ignore" ]
	then
		exit 0
	fi

        # See if there is an options file and if so add them to the options
        if [ -f "$1.options" ]
        then
                OPTIONS="$OPTIONS $(cat "$1".options)"
        fi

	$FTA $OPTIONS --json --output --xMaxCharsPerColumn 20000 --debug 1 --records 1000 "$i"
done

