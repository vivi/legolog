set terminal pdf size 1.5,1.3
set terminal pdfcairo font "Times New Roman,10"
set key spacing 1.1
#set key bottom right
#set key box
set key off
set output "auditor-work.pdf"
set datafile separator ','
set ylabel "Time taken (ms)"
set xlabel "Partitions"
#set xrange[0:1100]
#set xtics (4,128,256,512,1024)
plot "audit.csv" using 1:2 skip 1 with lines lw 3
