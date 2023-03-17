find data -name '*.csv.out' -print | xargs rm
date; find data -name '*.csv' -print | xargs -P 8 -n 20 bin/doitM.sh; date
bin/current.sh
