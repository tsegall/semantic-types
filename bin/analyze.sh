GREP=rg
diff ncurrent.csv current.csv | egrep '^<' | sed 's/^..//' | cut -f1,2 -d',' | sed 's/$/,/' > /tmp/$$.1
echo "Differences: $(wc -l < /tmp/$$.1)"
for i in `cat /tmp/$$.1`
do
	F=`$GREP $i reference.csv | cut -f5 -d','`
	R=`$GREP $i reference.csv | cut -f8 -d','`
	COMMENT=`$GREP $i reference.csv | cut -f9 -d','`
	if [ ! -z "$COMMENT" ]
	then
		COMMENT=", $COMMENT"
	fi
	N=`$GREP $i ncurrent.csv | cut -f8 -d','`
	C=`$GREP $i current.csv | cut -f8 -d','`

	if [ "$R" != "$N" ]
	then
		echo "---- $i, field $F, R = $R, N = $N, C = $C$COMMENT"
	else
		if [ "$R" == '""' -a "$N" == '""' ]
		then
			R=`$GREP $i reference.csv | cut -f6,7 -d','`
			N=`$GREP $i ncurrent.csv | cut -f6,7 -d','`
			if [ "$C" == '""' ]
			then
				C=`$GREP $i current.csv | cut -f6,7 -d','`
			fi
		fi

		echo "++++ $i, field $F, R = $R, N = $N, C = $C"
	fi
done

bin/performance | tail -1
cp current.csv /tmp
cp ncurrent.csv current.csv
bin/performance | tail -1
cp /tmp/current.csv .


