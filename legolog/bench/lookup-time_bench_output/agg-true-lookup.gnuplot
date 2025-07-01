set terminal pdf size 1.5,1.4
set terminal pdfcairo font "Times New Roman,10"
#set key spacing 1.1
#set key top left
#set key above maxrows 1 width -8 samplen 1.5
#set key box
set key reverse Left
set key top left samplen 1.5 width -2
set key box
set output "agg-true-lookup.pdf"
set datafile separator ','
set ylabel "Time taken (ms)"
set xlabel "Update periods/verification period"
set xrange[0:1000]
set xtics 0,200,1000
set ytics 1
plot "1-partitions_agg-true_lookup_times.csv" using 1:2 skip 1 with lines lw 3 title "1 partition", \
     #"2-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 with lines title "2", \
     #"4-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 with lines title "4", \
     "8-partitions_agg-true_lookup_times.csv" using 1:2 skip 1 with lines lw 3 title "8", \
     #"16-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 with lines title "16", \
     #"32-partitions_agg-false_lookup_times.csv" using 1:2 skip 1 with lines title "32", \
     "64-partitions_agg-true_lookup_times.csv" using 1:2 skip 1 with lines lw 3 title "64", \
