#!/bin/bash 
# Copyright 2019 Crunchy Data Solutions, Inc.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Enforce required environment variables
test="${PGO_CMD:?Need to set PGO_CMD env variable}"
test="${PGOROOT:?Need to set PGOROOT env variable}"
test="${PGO_OPERATOR_NAMESPACE:?Need to set PGO_OPERATOR_NAMESPACE env variable}"
test="${PGO_INSTALLATION_NAME:?Need to set PGO_INSTALLATION_NAME env variable}"

if [[ -z "$1" ]]; then
	echo "usage:  add-targeted-namespace.sh mynewnamespace"
	exit
fi

# create the namespace if necessary
$PGO_CMD get ns $1  > /dev/null
if [ $? -eq 0 ]; then
	echo "namespace" $1 "already exists"
else
	echo "namespace" $1 "is new"
	TARGET_NAMESPACE=$1 expenv -f $DIR/target-namespace.yaml | $PGO_CMD create -f -
fi

# set the labels so that this namespace is owned by this installation
$PGO_CMD label namespace/$1 pgo-created-by=add-script
$PGO_CMD label namespace/$1 vendor=crunchydata
$PGO_CMD label namespace/$1 pgo-installation-name=$PGO_INSTALLATION_NAME

# create RBAC
$PGO_CMD -n $1 delete --ignore-not-found sa pgo-backrest pgo-default pgo-pg pgo-target
$PGO_CMD -n $1 delete --ignore-not-found role pgo-backrest-role pgo-pg-role pgo-target-role
$PGO_CMD -n $1 delete --ignore-not-found rolebinding pgo-backrest-role-binding pgo-pg-role-binding pgo-target-role-binding

cat $PGOROOT/conf/postgres-operator/pgo-default-sa.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-target-sa.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-target-role.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-target-role-binding.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | sed 's/{{.OperatorNamespace}}/'"$PGO_OPERATOR_NAMESPACE"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-backrest-sa.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-backrest-role.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-backrest-role-binding.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-pg-sa.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-pg-role.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -
cat $PGOROOT/conf/postgres-operator/pgo-pg-role-binding.json | sed 's/{{.TargetNamespace}}/'"$1"'/' | $PGO_CMD -n $1 create -f -

if [ -r "$PGO_IMAGE_PULL_SECRET_MANIFEST" ]; then
	$PGO_CMD -n $1 create -f "$PGO_IMAGE_PULL_SECRET_MANIFEST"
fi

if [ -n "$PGO_IMAGE_PULL_SECRET" ]; then
	patch='{"imagePullSecrets": [{ "name": "'"$PGO_IMAGE_PULL_SECRET"'" }]}'

	$PGO_CMD -n $1 patch --type=strategic --patch="$patch" serviceaccount/pgo-backrest
	$PGO_CMD -n $1 patch --type=strategic --patch="$patch" serviceaccount/pgo-default
	$PGO_CMD -n $1 patch --type=strategic --patch="$patch" serviceaccount/pgo-pg
	$PGO_CMD -n $1 patch --type=strategic --patch="$patch" serviceaccount/pgo-target
fi
