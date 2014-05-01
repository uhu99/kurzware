#!/bin/sh

# QR Code Label generieren.
# Beispielaufruf:
# sh iqr.sh http://kurzware.de/q 'http://kurzware.de/q?r=' 1312 10

zeile1=$1
qrcode=$2
start=$3
anzahl=$4

i=0

while [ 1 ]
do
    n=$start$i
    url=$qrcode$n

    echo ./iqr -t1="$zeile1" -t2="?r=$n" -tqr="$url"
    ./iqr -size=11 -t1="http://kurzware" -t2=".de/q?r=$n" -tqr="$url"
    pngcrush -q -res 180 out.png qr$n.png
    pngcrush -q -res 180 out.png etiketten.rtfd/e_1_$i.png
    pngcrush -q -res 150 out.png etiketten.rtfd/e_2_$i.png

    (( i++ ))
    if  [ $i -ge $anzahl ]
    then
        break
    fi
done
