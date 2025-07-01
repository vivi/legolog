set datafile separator ','
plot "100-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 title "100 partitions", \
     "400-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 title "400", \
     "800-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 title "800", \
     "1600-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 title "1600"
