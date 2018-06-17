NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

.PHONY: all clean build zip

all: clean zip

ec2ami:
	@echo "$(OK_COLOR)==> Building... $(NO_COLOR)"
	@GOOS=linux GOARCH=amd64 go build -o ec2ami ec2ami.go

ec2amizip: ec2ami
	@echo "$(OK_COLOR)==> Zipping... $(NO_COLOR)"
	@zip ec2ami.zip ./ec2ami

clean:
	@rm -rf ec2ami ec2ami.zip