#
# Output a single record with the format below for every in field in the input file
#
# Format:
#	 File,FieldOffset,Locale,RecordCount,FieldName,BaseType,SemanticType,Notes
#
for file in "$@"
do
	SAMPLE_COUNT=$(egrep '"sampleCount" : ' "$file".out | head -1 | sed -e 's/.*sampleCount"...//' -e 's/,$//')
	LOCALE=$(egrep '"detectionLocale" : ' "$file".out | head -1 | sed -e 's/.*detectionLocale"...//' -e 's/,$//')
	for i in $(egrep '^Field' "$file".out | sed 's/.*(\([0-9]*\)).*/\1/')
	do
		FIELDNAME="$(egrep "^Field.*\($i\) - {" "$file".out | sed -e "s/Field '\([^']*\).*/\1/" -e 's/"/""/g')"
		TYPE=$(egrep -A10 "^Field.*\($i\) - {" "$file".out | egrep '"type" : "' | sed 's/  "type" : //' | sed 's/,$//')
		QUALIFIER="$(egrep -A12 "^Field.*\($i\) - {" "$file".out | egrep '"typeQualifier"' | sed 's/  "typeQualifier" : //' | sed 's/,$//')"
		if [ -z "$QUALIFIER" ]
		then
			QUALIFIER='""'
		fi
		echo $file,$i,$LOCALE,$SAMPLE_COUNT,'"'"$FIELDNAME"'"',$TYPE,"$QUALIFIER",'""'
	done
done
