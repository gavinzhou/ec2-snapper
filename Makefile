NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

.PHONY: all clean build zip

all: clean zip

daily:
	@echo "$(OK_COLOR)==> Building... $(NO_COLOR)"
	@GOOS=linux GOARCH=amd64 go build -o dailyec2ami dailyec2ami.go

dailyzip: daily
	@echo "$(OK_COLOR)==> Zipping... $(NO_COLOR)"
	@zip dailyec2ami.zip ./dailyec2ami

clean:
	@rm -rf dailyec2ami dailyec2ami.zip