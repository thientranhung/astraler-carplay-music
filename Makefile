APP     = astra-mix-sound
PYTHON  = venv/bin/python
PIP     = venv/bin/pip

.PHONY: help setup build clean run

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  setup   Tạo venv và cài dependencies"
	@echo "  build   Build binary release vào dist/"
	@echo "  run     Chạy trực tiếp bằng Python"
	@echo "  clean   Xóa build artifacts (dist/, build/, *.spec)"

setup:
	python3 -m venv venv
	$(PIP) install -q -r requirements-dev.txt
	@echo "✓ Setup xong. Chạy: make run"

build:
	$(PYTHON) -m PyInstaller \
		--onefile \
		--name $(APP) \
		--distpath dist \
		--workpath build \
		--clean \
		--noconfirm \
		mix.py
	@echo ""
	@echo "✓ Binary: dist/$(APP)"
	@echo "  Chạy thử: ./dist/$(APP) --help"

run:
	$(PYTHON) mix.py $(ARGS)

clean:
	rm -rf dist build __pycache__ *.spec
	@echo "✓ Cleaned"
