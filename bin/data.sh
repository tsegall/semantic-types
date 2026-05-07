#
# Reference file has the following headers:
#	File,FieldOffset,Locale,RecordCount,FieldName,BaseType,SemanticType,Notes
# Output the data for a given File,FieldOffset
#
# For example, data.sh data/opendata_socrata_com/data.ct.gov/2m3u-43yh.csv,2
#
FTA=$HOME/src/fta/cli/build/install/fta/bin/cli

for i in "$@"
do
	file=$(echo $i | cut -f1 -d',')
	field=$(echo $i | cut -f2 -d',')

        # See if there is an options file and if so add them to the options
        if [ -f "$file.options" ]
        then
                OPTIONS="$OPTIONS $(cat "$file".options)"
        fi

	$FTA $OPTIONS --debug 2 --noAnalysis --verbose --col $field "$file"
done

