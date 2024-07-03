PROJECT_NAME := minigit

brun:
	go build .
	./$(PROJECT_NAME) -dir=.
c2k :
	go build .
	./$(PROJECT_NAME) -mode=c2k
imglib :
	go build .
	./$(PROJECT_NAME) -mode=imglib
vips :
	go build .
	./$(PROJECT_NAME) -mode=vips
