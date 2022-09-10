#
# Adjust the detected Semantic Type in the reference file
#
# Example: addType.sh data/opendata_socrata_com/data.ct.gov/2m3u-43yh.csv,2 TELEPHONE
#
egrep 'new file.*.csv.out' /tmp/q
file=$(echo $1 | cut -f1 -d',')
field=$(echo $1 | cut -f2 -d',')

sed -i .bak "s+^\($1,.*\),\"\",\"\"$+\1,\"$2\",\"\"+" reference.csv
cmp -s reference.csv reference.csv.bak
if [ $? -eq 0 ]
then
	echo "Warning: No update!!" 1>&2
fi
