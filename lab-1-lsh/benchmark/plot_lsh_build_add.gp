set terminal pngcairo size 1280,720
set output 'lsh_build_add.png'

set title 'LSH: build vs add'
set xlabel 'Number of documents'
set ylabel 'Time per document, us'
set grid
set key left top
set logscale x 10

plot \
  'lsh_build_add.dat' using 1:2 with linespoints lw 2 pt 7 title 'Build', \
  'lsh_build_add.dat' using 1:2:(sprintf("%.3f", $2)) with labels offset 0,1 notitle, \
  'lsh_build_add.dat' using 1:3 with linespoints lw 2 pt 5 title 'Add', \
  'lsh_build_add.dat' using 1:3:(sprintf("%.3f", $3)) with labels offset 0,-1 notitle