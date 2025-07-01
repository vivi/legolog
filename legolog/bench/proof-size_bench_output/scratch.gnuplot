set datafile separator ','
set ylabel "Time taken (ns)"
set xlabel "Update periods in 1 verification period (#)"
plot "1-partitions_agg-false_proof_sizes.csv" using 1:2 skip 1 with lines title "1 partitions", \
    "8-partitions_agg-false_proof_sizes.csv" using 1:2 skip 1 with lines title "8 partitions", \
