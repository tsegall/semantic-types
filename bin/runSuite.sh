find data -name '*.csv.out' -print | xargs rm
date; find data -name '*.csv' -print | xargs -P 8 -n 1 bin/doit.sh; date
bin/current.sh
