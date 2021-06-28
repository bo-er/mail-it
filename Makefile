
override OUTPUT = bin

.PHONY:
build:
	go build -o ${OUTPUT}/${NAME} .
.PHONY:
clean:
	cd ${OUTPUT} && rm -rf *