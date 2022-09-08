( echo File,FieldOffset,Locale,RecordCount,FieldName,BaseType,SemanticType,Notes ; \
find data -name '*.csv.out' -print | sed 's/.csv.out/.csv/' | xargs -n 8 -P 8 ./bin/makedb.sh | sort -t',' -k 1,1 -k 2,2n ) > current.csv
