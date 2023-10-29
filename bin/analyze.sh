diff ncurrent.csv current.csv | egrep '^<' | sed 's/^..//' | cut -f1,2 -d',' | sed 's/$/,/' > /tmp/$$.1
echo "Differences: $(wc -l < /tmp/$$.1)"
for i in `cat /tmp/$$.1`
do
	F=`egrep $i reference.csv | cut -f5 -d','`
	R=`egrep $i reference.csv | cut -f8 -d','`
	COMMENT=`egrep $i reference.csv | cut -f9 -d','`
	if [ ! -z "$COMMENT" ]
	then
		COMMENT=", $COMMENT"
	fi
	N=`egrep $i ncurrent.csv | cut -f8 -d','`
	C=`egrep $i current.csv | cut -f8 -d','`

	if [ "$R" != "$N" ]
	then
		echo "---- $i, field $F, R = $R, N = $N, C = $C$COMMENT"
	else
		if [ "$R" == '""' -a "$N" == '""' ]
		then
			R=`egrep $i reference.csv | cut -f6,7 -d','`
			N=`egrep $i ncurrent.csv | cut -f6,7 -d','`
			C=`egrep $i current.csv | cut -f6,7 -d','`
		fi

		echo "++++ $i, field $F, R = $R, N = $N, C = $C"
	fi
done

bin/performance | tail -1
cp current.csv /tmp
cp ncurrent.csv current.csv
bin/performance | tail -1
cp /tmp/current.csv .


