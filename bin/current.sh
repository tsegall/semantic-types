( echo File,FieldOffset,Locale,RecordCount,FieldName,BaseType,TypeModifier,SemanticType,Notes ; \
find data -name '*.csv.out' -print | xargs -n 100 ./bin/makedb | LC_COLLATE=C sort -t',' -k 1,1 -k 2,2n ) > ncurrent.csv
