#!/bin/bash


# ps -ef | grep /usr/lib/golang | grep -v grep
# for i in $(ps -ef  | grep /usr/lib/golang | grep -v grep | awk '{print $2}'); do kill $i; /bin/sleep 0.1; done
#
# /bin/sleep 1
#
# ps -ef | grep go-build | grep -v grep
# for i in $(ps -ef  | grep go-build | grep -v grep | awk '{print $2}'); do kill $i; /bin/sleep 0.1; done
#
# /bin/sleep 1

# ps -ef | grep go | grep -v grep
# for i in $(ps -ef  | grep go | grep -v grep | awk '{print $2}'); do kill $i; /bin/sleep 1; done
kill -9 $(ps -ef | grep go | grep -v grep | awk '{print $2}')

/bin/sleep 0.5

sync; echo 3 > /proc/sys/vm/drop_caches

/bin/sleep 1

cd admin
go run main.go &

/bin/sleep 1

cd ../room-roulette
go run main.go &

/bin/sleep 1

cd ../room-haochehui
go run main.go &

/bin/sleep 1

cd ../room-pinshi
go run main.go &

/bin/sleep 1

cd ../room-fruit
go run main.go &

/bin/sleep 1

cd ../gamespace-lobby
go run main.go &

/bin/sleep 1

ps -ef | grep go | grep -v grep
