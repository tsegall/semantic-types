if [ -z "$FTA" ]
then
	FTA=$HOME/src/fta/cli/build/install/fta/bin/cli
fi

files=""

for i in "$@"
do
	if [ ! -f "$i.ignore" ]
	then
		files="$files $i"
	fi
done

if [ -z "$files" ]
then
	exit 0
fi

$FTA $OPTIONS --validate 1 --json --output --xMaxCharsPerColumn 20000 --debug 1 --records 1000  $files

if [ $? -eq 1 ]
then
	echo "Problem with $files" 1>&2
	exit 1

fi
