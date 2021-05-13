foreach mode ( \
  ideal locking finelocking parsteam americano espresso \
  multi-1 multi-2 multi-4 multi-8 \
  linearpipe-0 linearpipe-1 linearpipe-10 \
  splitpipe-0 splitpipe-1 splitpipe-10 \
  multipipe-1 multipipe-2 multipipe-4 multipipe-8 \
  myLattepipe-1 myLattepipe-2 myLattepipe-4 myLattepipe-8 \
)

go run *.go --dur=2s --par=2,5,8 -mode=$mode

end