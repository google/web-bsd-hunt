#!/bin/bash -eu
#
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
#!/bin/bash

set -o errexit
set -o nounset

prog=$0
has_error=n

fatal()
{
	echo "${prog}: FATAL: $*" 1>&2
	exit 1
}

error()
{
	echo "${prog}: ERROR: $*" 1>&2
	has_error=y
}

run()
{
	echo "$*"
	$*
}

help()
{
	exec 1>&2
	echo "usage: ${prog} args" 
	echo " Required Arguments:"
	echo "    --huntd-well-known-host hostname-or-ip"
	echo "        The host running huntd (usually localhost)"
	echo "    --huntd-well-known-port port-number"
	echo "        The huntd service discovery port"
	echo "    --server-host hostname-or-ip"
	echo "        The host the game server will listen on (usually 0.0.0.0 for all interfaces)" | fmt
	echo "    --server-port port number"
	echo "        The port the game server will listen on"
	echo "    --rpc-type netrpc | jsonrpc"
	echo "        The type of RPC server to run"
	exit 1
}

version=$(getopt --test > /dev/null 2>&1 || echo $?)
if [ "$version" != "4" ] ; then
	fatal "unsupported getopt version"
fi

long="huntd-well-known-host:,huntd-well-known-port:,server-host:,server-port:,rpc-type:"

parsed=$(getopt -o h --longoptions "${long}" --name "$0" -- "$@")

if [[ $? != 0 ]] ; then
	fatal "bad args"
fi

while [ $# -gt 0 ] ; do
	case $1 in
	-h)
		help
		;;
	--huntd-well-known-host)
		huntd_host="$2"
		shift 2
		;;
	--huntd-well-known-port)
		huntd_port="$2"
		shift 2
		;;
	--server-host)
		server_host="$2"
		shift 2
		;;
	--server-port)
		server_port="$2"
		shift 2
		;;
	--rpc-type)
		rpc_type="$2"
		shift 2
		;;
	--)
		shift
		break
		;;
	*)
		echo "$0: FATAL: unhandled arg: $1" 1>&2
		exit 1
		;;
	esac
done

eval set -- "$parsed"

if [ -z "${huntd_host+x}" ] ; then
	error "missing --huntd-well-known-host"
fi

if [ -z "${huntd_port+x}" ] ; then
	error "missing --huntd-well-known-port"
fi

if [ -z "${server_host+x}" ] ; then
	error "missing --server-host"
fi

if [ -z "${server_port+x}" ] ; then
	error "missing --server-port"
fi

if [ -z "${rpc_type+x}" ] ; then
	error "missing --rpc-type"
fi

if [ "${has_error}" = "y" ] ; then
	fatal "missing args"
fi

#
# huntd daemonizes
#
# NOTE: the Redirect in from /dev/null to address an issue onGAE Flex.
# Huntd misuses poll(2), but setting unused file descriptiors to 0
# in the pollfd array, when it should be setting them to -1.
# It turns out that after the first client connects, something strange
# happens to FD 0, which results in the POLLHUP revent firing
# for each 0 entry.
#
# This should make sure nothing weird ever happsn to FD 0, so
# we can go on using the broken implementation
#
run /usr/sbin/huntd -s -p "${huntd_port}" < /dev/null

#
# the game server does not daemonize
#
run exec /go/bin/server-game \
	-huntd-well-known-host "${huntd_host}" \
	-huntd-well-known-port "${huntd_port}" \
	-server-host "${server_host}" \
	-server-port "${server_port}" \
	-rpc-type "${rpc_type}"
