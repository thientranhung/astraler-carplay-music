APP     = astra-carplay-music
VERSION = 1.0.0

.PHONY: help build sign release clean run

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build    Build binary vào dist/"
	@echo "  sign     Code sign binary (cần Developer ID)"
	@echo "  release  Build + sign + notarize + tạo tarball"
	@echo "  run      Chạy trực tiếp"
	@echo "  clean    Xóa build artifacts"

build:
	@mkdir -p dist
	go build -ldflags="-s -w" -o dist/$(APP) .
	@echo "✓ dist/$(APP)"

sign: build
	codesign --sign "Developer ID Application: Cuong Nguyen (LSJVHTU65U)" \
		--options runtime \
		--timestamp \
		--force \
		dist/$(APP)
	@echo "✓ Signed"

release: sign
	@# Notarize
	rm -f dist/$(APP).zip
	ditto -c -k --sequesterRsrc dist/$(APP) dist/$(APP).zip
	xcrun notarytool submit dist/$(APP).zip \
		--keychain-profile "$(APP)" \
		--wait
	@# Package cho Homebrew
	rm -f dist/$(APP)-v$(VERSION)-macos.tar.gz
	tar -czf dist/$(APP)-v$(VERSION)-macos.tar.gz -C dist $(APP)
	@echo ""
	@shasum -a 256 dist/$(APP)-v$(VERSION)-macos.tar.gz
	@echo "✓ dist/$(APP)-v$(VERSION)-macos.tar.gz"

run:
	go run . $(ARGS)

clean:
	rm -rf dist/
	@echo "✓ Cleaned"
