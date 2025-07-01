set terminal pdf size 1.5,1.4
set terminal pdfcairo font "Times New Roman,10"
set key reverse Left
set key top left samplen 1.5 width -2
set key box
set output "agg-false-validation.pdf"
set datafile separator ','
set ylabel "Time taken (ms)"
set xlabel "Update periods/verification period"
set xrange[0:1000]
set xtics 0,200,1000
plot "1-partitions_agg-false_proof_validations.csv" using 1:2 skip 1 with lines lw 3 title "1 partition", \
     "8-partitions_agg-false_proof_validations.csv" using 1:2 skip 1 with lines lw 3 title "8", \
     "64-partitions_agg-false_proof_validations.csv" using 1:2 skip 1 with lines lw 3 title "64", \
