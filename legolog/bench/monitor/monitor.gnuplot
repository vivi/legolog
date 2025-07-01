set terminal pdf size 1.5,1.3
set terminal pdfcairo font "Times New Roman,10"
set key spacing 1.1
#set key bottom right
#set key box
set key top left
set key reverse Left
set key width -2
set key box
set key font ",9"
set key samplen 1.5
set output "monitor-work.pdf"
set datafile separator ','
set ylabel "Time taken (ms)"
set xlabel "Verification periods offline"
set xrange[0:550]
#set xtics (4,128,256,512,1024)
plot "monitor_agghist_false.csv" using 2:3 skip 1 with lines lw 3 title "No history forest", \
     "monitor_agghist_true.csv" using 2:3 skip 1 with lines lw 3 title "History forest",
