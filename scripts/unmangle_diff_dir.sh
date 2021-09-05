#!/bin/bash

# This scripts undoes the changes made by mangle_diff_dir.sh

BASE=${PWD}/data/diff_dir

mv /tmp/a.dat ${BASE}/a.dat
mv /tmp/e.dat ${BASE}/c/e.dat
mv ${BASE}/c/g.dat ${BASE}/c/f.dat
rm ${BASE}/c/h.dat
rm ${BASE}/d
mkdir ${BASE}/d

exit 0
