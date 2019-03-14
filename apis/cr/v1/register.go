package v1

/*
Copyright 2019 Crunchy Data Solutions, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeBuilder ...
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme ...
	AddToScheme = SchemeBuilder.AddToScheme
)

// GroupName is the group name used in this package.
//const GroupName = "cr.client-go.k8s.io"
const GroupName = "crunchydata.com"

// SchemeGroupVersion is the group version used to register these objects.
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

// Resource takes an unqualified resource and returns a Group-qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Pgcluster{},
		&PgclusterList{},
		&Pgreplica{},
		&PgreplicaList{},
		&Pgupgrade{},
		&PgupgradeList{},
		&Pgbackup{},
		&PgbackupList{},
		&Pgpolicy{},
		&PgpolicyList{},
		&Pgtask{},
		&PgtaskList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
