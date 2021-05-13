foreach mode ( \
#  myLattepipe-1 myLattepipe-2 myLattepipe-4 myLattepipe-8 \
  myLattepipe-1 \
)

go run *.go --dur=2s --par=2,5,8 -mode=$mode

end