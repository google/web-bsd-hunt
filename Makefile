# Copyright 2016 The Web BSD Hunt Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
################################################################################
#
# TODO: High-level file comment.
ROOT			:= $(shell pwd)
GOBIN			:= ${ROOT}/bin

GO_LIB_PATH		:= ${ROOT}/lib

GO_TPARTY_PATH		:= ${ROOT}/third_party

FRONTEND_GOPATH		:= ${ROOT}/server-frontend
FRONTEND_DIR		:= server-frontend/src/server-frontend
FRONTEND_VERSION	:= v1

GAME_GOPATH		:= ${ROOT}/server-game
GAME_DIR		:= server-game/src/server-game

PATH			:= ${GOBIN}:${GO_TPARTY_PATH}/bin:${PATH}

ifndef PROJECT
$(error set PROJECT to the name of your GCP project)
endif

export GOBIN
export PATH

server_game_version	:= $(shell date +%Y-%m-%d-%H-%M-%S)
server_game_host	:= gcr.io
server_game_path	:= $(shell echo ${PROJECT} | sed 's/:/\//g')/server-game-${server_game_version}
server_game_tag		:= ${server_game_host}/${server_game_path}

.PHONY: build
build: setup-third_party build-server-frontend build-server-game

.PHONY: clean
clean: clean-server-frontend clean-server-game

.PHONY: deploy
deploy: deploy-server-frontend deploy-server-game

.PHONY: build-server-frontend
build-server-frontend: export GOPATH=${FRONTEND_GOPATH}:${FRONTEND_GOPATH}/vendor
build-server-frontend: install-deps-server-frontend
	cd ${FRONTEND_GOPATH} && goapp build -o ${GOBIN}/server-frontend server-frontend

.PHONY: clean-server-frontend
clean-server-frontend: clean-deps-server-frontend
	rm -f ${GOBIN}/server-frontend

.PHONY: install-deps-server-frontend
install-deps-server-frontend: export GOPATH=${GO_TPARTY_PATH}:${GO_LIB_PATH}:${FRONTEND_GOPATH}
install-deps-server-frontend: clean-deps-server-frontend
	cd ${FRONTEND_DIR} && govendor init
	cd ${FRONTEND_DIR} && govendor add +external
	mkdir -p server-frontend/vendor/src
	mv ${FRONTEND_DIR}/vendor/* server-frontend/vendor/src
	mkdir -p ${FRONTEND_DIR}/assets/fonts/VT220
	cp ${GO_TPARTY_PATH}/assets/christfollower.me/misc/glasstty/* ${FRONTEND_DIR}/assets/fonts/VT220
	mkdir -p ${FRONTEND_DIR}/assets/sounds
	cp ${GO_TPARTY_PATH}/assets/soundbible.com/* ${FRONTEND_DIR}/assets/sounds
	mkdir -p ${FRONTEND_DIR}/assets/js/pixi.js/bin
	cp ${GO_TPARTY_PATH}/src/github.com/pixijs/pixi.js/bin/pixi.min.js ${FRONTEND_DIR}/assets/js/pixi.js/bin
	mkdir -p ${FRONTEND_DIR}/assets/js/Keypress
	cp ${GO_TPARTY_PATH}/src/github.com/dmauro/Keypress/keypress-2.1.4.min.js ${FRONTEND_DIR}/assets/js/Keypress

.PHONY: clean-deps-server-frontend
clean-deps-server-frontend:
	rm -rf server-frontend/src/server-frontend/vendor
	rm -rf server-frontend/vendor
	rm -rf server-frontend/src/server-frontend/assets

.PHONY: play-server-frontend
play-server-frontend: export GOPATH=${FRONTEND_GOPATH}:${FRONTEND_GOPATH}/vendor
play-server-frontend: setup-third_party install-deps-server-frontend
	goapp serve -host 0.0.0.0 server-frontend/src/server-frontend/app-local.yaml

.PHONY: deploy-server-frontend
deploy-server-frontend: export GOPATH=${FRONTEND_GOPATH}:${FRONTEND_GOPATH}/vendor
deploy-server-frontend: build-server-frontend
	appcfg.py -A ${PROJECT} -V ${FRONTEND_VERSION} update ${FRONTEND_DIR}

#
# XXX(tadhunt): This doesn't work even though according to the documentation it should
#
#deploy-server-frontend: export GOPATH=${FRONTEND_GOPATH}:${FRONTEND_GOPATH}/vendor
#deploy-server-frontend: build-server-frontend
#	cd ${FRONTEND_DIR} && gcloud app deploy \
#		app.yaml \
#		--account ${ACCOUNT} \
#		--project ${PROJECT} \
#		--stop-previous-version \
#		--promote \
#		--version ${FRONTEND_VERSION}

.PHONY: build-server-game
build-server-game: export GOPATH=${GAME_GOPATH}:${GAME_GOPATH}/vendor
build-server-game: install-deps-server-game
	rm -f ${GOBIN}/server-game
	cd ${GAME_DIR} && go build -o ${GOBIN}/server-game
	rm -f ${GOBIN}/server-game-run && cp ${GAME_DIR}/run.sh ${GOBIN}/server-game-run

.PHONY: clean-server-game
clean-server-game: clean-deps-server-game
	rm -f ${GOBIN}/server-game ${GOBIN}/server-game-run

.PHONY: install-deps-server-game
install-deps-server-game: export GOPATH=${GO_TPARTY_PATH}:${GO_LIB_PATH}:${GAME_GOPATH}
install-deps-server-game: clean-deps-server-game
	cd ${GAME_DIR} && govendor init
	cd ${GAME_DIR} && govendor add +external
	mkdir -p server-game/vendor/src
	mv ${GAME_DIR}/vendor/* server-game/vendor/src

.PHONY: clean-deps-server-game
clean-deps-server-game:
	rm -rf ${GAME_DIR}/vendor
	rm -rf server-game/vendor

.PHONY: play-server-game
play-server-game: build-server-game
	SERVER_GAME_URL="http://localhost:8080" \
	SERVER_GAME_OPTIONS="LOG_STARTUP,LOG_EVENT,LOG_RPC,LOG_HUNTD_CONNECT,LOG_PLAYER_API,LOG_KEEPALIVE" \
	${GOBIN}/server-game \
		-server-host localhost \
		-server-port 12345 \
		-huntd-well-known-port 4444 \
		-rpc-type jsonrpc

.PHONY: server-game.tag
server-game.tag:
	rm -f server-game.tag
	docker build --tag ${server_game_tag} -f ${GAME_DIR}/Dockerfile .
	echo ${server_game_tag} > server-game.tag

.PHONY: build-server-game-image
build-server-game-image: build-server-game server-game.tag

.PHONY: push-server-game
push-server-game: tag=$(shell cat server-game.tag)
push-server-game: build-server-game-image
	gcloud docker push ${tag}

.PHONY: deploy-server-game
deploy-server-game: tag=$(shell cat server-game.tag)
deploy-server-game: export GOPATH=${GAME_GOPATH}:${GAME_GOPATH}/vendor
deploy-server-game: push-server-game
	cd ${GAME_DIR} && gcloud app deploy \
		app.yaml \
		--image-url ${tag} \
		--promote \
		--stop-previous-version \
		--version ${server_game_version}

.PHONY: setup-third_party
setup-third_party:
	cd third_party && ${MAKE} setup

.PHONY: install-third_party
install-third_party: clean-deps
	cd third_party && ${MAKE} install

.PHONY: clean-third_party
clean-third_party:
	cd third_party && ${MAKE} clean

.PHONY: clean-deps
clean-deps: clean-deps-server-frontend clean-deps-server-game

.PHONY: play-huntd
play-huntd:
	huntd -s -p 4444

.PHONY: play-stop
play-stop:
	pkill server-game || :
	pkill huntd || :
	pkill goapp || :
	sleep 1

.PHONY: autogen-copyright
autogen-copyright:
	echo "~/dist/autogen/autogen.sh -i -l apache -c 'The Web BSD Hunt Authors.' <file>"
