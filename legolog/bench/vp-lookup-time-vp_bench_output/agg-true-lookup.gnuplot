set terminal pdf size 3,2
set terminal pdfcairo font "Times New Roman,10"
#set key reverse Left
set key bottom right samplen 1.5 width -2
set key box
set output "agg-true-lookup-vp.pdf"
set datafile separator ','
set ylabel "Time taken (ms)"
set xlabel "Verification periods elapsed"
set xrange[0:1100]
set xtics (4,128,256,512,1024)
set xtics rotate by 45 right
plot "1-partitions_agg-true_lookup_times.csv" using 1:2 skip 1  with lines lw 3 title "1 partition", \
     "8-partitions_agg-true_lookup_times.csv" using 1:2 skip 1  with lines lw 3 title "8", \
     "64-partitions_agg-true_lookup_times.csv" using 1:2 skip 1 with lines lw 3 title "64", \
