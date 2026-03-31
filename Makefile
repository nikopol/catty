DIST_DIR := dist
BIN := $(DIST_DIR)/catty

.PHONY: build clean

build:
	mkdir -p $(DIST_DIR)
	go build -ldflags "-w"  -o $(BIN) .
	stat -c "built $(BIN) (%s bytes)" $(BIN)
	./$(BIN) -v

clean:
	rm -rf $(DIST_DIR) catty
