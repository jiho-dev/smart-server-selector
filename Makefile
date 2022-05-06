BIN=sss

all:
	go build -o ${BIN}
	cp sss ~/bin/

menu:
#	./${BIN}
	./${BIN} dev2

con:
	#./${BIN} con --config "/home/jiho.jung/src/jiho-dev/smart-server-selector/sss.yaml" --host-name lab-eu-1a-0504-h17
	./${BIN} lab-eu-1b-1209-h18


