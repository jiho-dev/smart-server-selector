all:
	go build

menu:
	./smart-server-selector menu --config "/home/jiho.jung/src/jiho-dev/smart-server-selector/sss.yaml" dev2

con:
	#./smart-server-selector con --config "/home/jiho.jung/src/jiho-dev/smart-server-selector/sss.yaml" --host-name lab-eu-1a-0504-h17
	./smart-server-selector con --config "/home/jiho.jung/src/jiho-dev/smart-server-selector/sss.yaml" --host-name 172.19.6.0


