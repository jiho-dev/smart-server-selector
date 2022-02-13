BIN=sss

all:
	go build -o ${BIN}

menu:
	./${BIN} menu --config "/home/jiho.jung/.ssh/sss.yaml" 

con:
	#./${BIN} con --config "/home/jiho.jung/src/jiho-dev/smart-server-selector/sss.yaml" --host-name lab-eu-1a-0504-h17
	./${BIN} con --config "/home/jiho.jung/.ssh/sss.yaml" --host-name 172.19.6.0


