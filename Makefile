SOURCE = ./cmd/piece
BUILD  = build

all: binary

binary: $(BUILD) binary-linux binary-macos binary-windows

$(BUILD):
	-mkdir $(BUILD)

binary-linux: $(BUILD)/linux/piece $(BUILD)/linux/piece-386 $(BUILD)/linux/piece-arm $(BUILD)/linux/piece-arm64

$(BUILD)/linux: $(BUILD)
	-mkdir $@

$(BUILD)/linux/piece $(BUILD)/linux/piece%: $(BUILD)/linux
	GOOS=linux GOARCH=$(patsubst %/piece,amd64,$(patsubst build/linux/piece-%,%,$@)) \
		go build -v \
		-ldflags "-s -w" \
		-o $(patsubst %-amd64,%,$@) \
		$(SOURCE)

binary-macos: $(BUILD)/macos/piece $(BUILD)/macos/piece-386

$(BUILD)/macos: $(BUILD)
	-mkdir $@

$(BUILD)/macos/piece $(BUILD)/macos/piece%: $(BUILD)/macos
	GOOS=darwin GOARCH=$(patsubst %/piece,amd64,$(patsubst build/macos/piece-%,%,$@)) \
		go build -v \
		-ldflags "-s -w" \
		-o $(patsubst %-amd64,%,$@) \
		$(SOURCE)

binary-windows: $(BUILD)/windows/piece.exe $(BUILD)/windows/piece-386.exe

$(BUILD)/windows: $(BUILD)
	-mkdir $@

$(BUILD)/windows/piece.exe $(BUILD)/windows/piece%.exe: $(BUILD)/windows
	GOOS=windows GOARCH=$(patsubst %/piece.exe,amd64,$(patsubst build/windows/piece-%.exe,%,$@)) \
		go build -v \
		-ldflags "-s -w" \
		-o $(patsubst %-amd64.exe,%.exe,$@) \
		$(SOURCE)
