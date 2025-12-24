ifeq ($(OS),Windows_NT)
	OS := windows
else ifeq ($(shell uname -s),Darwin)
	OS := darwin
else
	OS := linux
endif

.PHONY: windows linux darwin clean clean_windows clean_linux clean_darwin clean_all

APP_NAME := dsda-launch
APP_ID := us.hakubi.dsda-launch
VERSION := 1.0.0
BUILD_DIR_PREFIX := build
BUILD_DIR := build/$(OS)


CPU := x86_64
LINUXDEPLOY := NO_STRIP=1 linuxdeploy-$(CPU).AppImage
APP_DIR_PREFIX := $(BUILD_DIR)/$(APP_NAME)
APP_DIR := $(APP_DIR_PREFIX).AppDir

default: $(OS)

windows:
	mkdir "$(BUILD_DIR)" > NUL 2> NUL || echo > NUL
	fyne build --os windows --output "$(BUILD_DIR)/windows/$(APP_NAME).exe" .
	fyne package --os windows --exe "$(BUILD_DIR)/windows/$(APP_NAME).exe" --icon icon.png -app-id "$(APP_ID)" --name "$(APP_NAME)" --source-dir .

linux:
	mkdir -p "$(BUILD_DIR)"
	fyne build --os linux --output "$(APP_DIR_PREFIX)" .
	fyne package --os linux --exe "$(APP_DIR_PREFIX)" --icon icon.png -app-id $(APP_ID) --name "$(APP_NAME)" --source-dir .
	mv "$(APP_NAME).tar.xz" $(BUILD_DIR)
	rm -rf $(APP_DIR)
	mkdir -p "$(APP_DIR)"
	tar xJf "$(BUILD_DIR)/$(APP_NAME).tar.xz" -C "$(APP_DIR)"
	mv "$(APP_DIR)"/usr/local/* "$(APP_DIR)"/usr
	rmdir "$(APP_DIR)"/usr/local
	echo "Categories=Game" >> $(APP_DIR)/usr/share/applications/$(APP_ID).desktop
	$(LINUXDEPLOY) --appdir $(APP_DIR) \
	               --executable $(APP_DIR)/usr/bin/$(APP_NAME) \
	               --desktop-file $(APP_DIR)/usr/share/applications/$(APP_ID).desktop \
	               --icon-file $(APP_DIR)/usr/share/pixmaps/$(APP_ID).png \
	               --output appimage
	mv $(APP_NAME)-$(CPU).AppImage $(BUILD_DIR)

darwin:
	mkdir -p "$(BUILD_DIR)"
	fyne build --os darwin --output "$(BUILD_DIR)/$(APP_NAME).app" .
	fyne package --os darwin --exe "$(BUILD_DIR)/$(APP_NAME).app" --icon icon.png -app-id "$(APP_ID)" --name "$(APP_NAME)" --source-dir .

clean: clean_$(OS) clean_all

clean_windows:
	rd /s /q "$(BUILD_DIR)"

clean_linux:
	rm -rf "$(BUILD_DIR)"

clean_darwin:
	rm -rf "$(BUILD_DIR)"

clean_all:
	go clean
