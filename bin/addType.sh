#
# Adjust the detected Semantic Type in the reference file
#
# Example: addType.sh data/opendata_socrata_com/data.ct.gov/2m3u-43yh.csv,2 TELEPHONE
#
file=$(echo $1 | cut -f1 -d',')
field=$(echo $1 | cut -f2 -d',')

if [ "$field" == "$file" ]
then
	echo "Warning: Missing field!!" 1>&2
	exit 1
fi

sed -i .bak "s+^\($1,.*\),\"\",\"\"$+\1,\"$2\",\"\"+" reference.csv
cmp -s reference.csv reference.csv.bak
if [ $? -eq 0 ]
then
	echo "Warning: $1 - No update!!" 1>&2
	exit 1
fi
