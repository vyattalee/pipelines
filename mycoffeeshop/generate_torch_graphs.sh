foreach mode ( \
  ideal locking finelocking parsteam americano espresso \
  multi-1 multi-2 multi-4 multi-8 \
  linearpipe-0 linearpipe-1 linearpipe-10 \
  splitpipe-0 splitpipe-1 splitpipe-10 \
  multipipe-1 multipipe-2 multipipe-4 multipipe-8 \
)
#taskset -c 0-5 go run *.go --dur=3s --par=0 --trace=./trace-$mode.out -mode=$mode
#'GOMAXPROCS=5'
go run *.go --dur=3s --par=0 --trace=./trace-$mode.out -mode=$mode
go tool trace --pprof=sync ./trace-$mode.out > ./sync-$mode.pprof
go-torch -b ./sync-$mode.pprof -f ./torch-$mode.svg
#go tool pprof -http=":8081" ./bin/coffee ./sync-$mode.pprof
#./sync-$mode.pprof
end