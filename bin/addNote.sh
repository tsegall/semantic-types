#
# Add a Note to the reference file
#
# Example: addNote.sh data/opendata_socrata_com/data.ct.gov/2m3u-43yh.csv,2 "NOTE: This field has issues with formatting"
#
file=$(echo $1 | cut -f1 -d',')
field=$(echo $1 | cut -f2 -d',')

sed -i .bak "s+^\($1,.*\),\"\"$+\1,\"$2\"+" reference.csv
cmp -s reference.csv reference.csv.bak
if [ $? -eq 0 ]
then
	echo "Warning: No update!!" 1>&2
fi
