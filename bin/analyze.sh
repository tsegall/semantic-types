find data -name '*.csv'  -print | xargs -n 1 -P 8 bin/doit.sh
bin/current.sh
