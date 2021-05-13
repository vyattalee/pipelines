# !/bin/sh

for pid in $(ps -ef | grep pprof | grep -v grep | cut -c 13-20); do
    echo $pid
    kill -9 $pid
done



#while [ ! -z $(ps -ef | grep curl | grep -v grep | cut -c 9-15) ]
#do
#    ps -ef | grep curl | grep -v grep | cut -c 15-20 | xargs kill -9
#    ps -ef | grep curl | grep -v grep | cut -c 9-15 | xargs kill -9
#done
