package operator

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
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"

	"gopkg.in/yaml.v2"

	crv1 "github.com/crunchydata/postgres-operator/apis/cr/v1"
	"github.com/crunchydata/postgres-operator/config"
	"github.com/crunchydata/postgres-operator/kubeapi"
	"github.com/crunchydata/postgres-operator/util"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// consolidate with cluster.affinityTemplateFields
const AffinityInOperator = "In"
const AFFINITY_NOTINOperator = "NotIn"

const DefaultArchiveTimeout = "60"
const DefaultPgHaConfigMapSuffix = "pgha-default-config"

// affinityType represents the two affinity types provided by Kubernetes, specifically
// either preferredDuringSchedulingIgnoredDuringExecution or
// requiredDuringSchedulingIgnoredDuringExecution
type affinityType string

const (
	requireScheduleIgnoreExec affinityType = "requiredDuringSchedulingIgnoredDuringExecution"
	preferScheduleIgnoreExec  affinityType = "preferredDuringSchedulingIgnoredDuringExecution"
)

type affinityTemplateFields struct {
	NodeLabelKey   string
	NodeLabelValue string
	OperatorValue  string
}

type podAntiAffinityTemplateFields struct {
	AffinityType            affinityType
	ClusterName             string
	PodAntiAffinityLabelKey string
	VendorLabelKey          string
	VendorLabelValue        string
}

// consolidate
type collectTemplateFields struct {
	Name           string
	JobName        string
	CCPImageTag    string
	CCPImagePrefix string
	PgPort         string
	ExporterPort   string
}

//consolidate
type badgerTemplateFields struct {
	CCPImageTag        string
	CCPImagePrefix     string
	BadgerTarget       string
	PGBadgerPort       string
	ContainerResources string
}

type PgbackrestEnvVarsTemplateFields struct {
	PgbackrestStanza            string
	PgbackrestDBPath            string
	PgbackrestRepo1Path         string
	PgbackrestRepo1Host         string
	PgbackrestRepo1Type         string
	PgbackrestLocalAndS3Storage bool
	PgbackrestPGPort            string
}

type PgbackrestS3EnvVarsTemplateFields struct {
	PgbackrestS3Bucket    string
	PgbackrestS3Endpoint  string
	PgbackrestS3Region    string
	PgbackrestS3Key       string
	PgbackrestS3KeySecret string
}

type PgmonitorEnvVarsTemplateFields struct {
	PgmonitorPassword string
}

// needs to be consolidated with cluster.DeploymentTemplateFields
// DeploymentTemplateFields ...
type DeploymentTemplateFields struct {
	Name                string
	ClusterName         string
	Port                string
	CCPImagePrefix      string
	CCPImageTag         string
	CCPImage            string
	Database            string
	DeploymentLabels    string
	PodLabels           string
	DataPathOverride    string
	ArchiveMode         string
	ArchivePVCName      string
	XLOGDir             string
	BackrestPVCName     string
	PVCName             string
	RootSecretName      string
	UserSecretName      string
	PrimarySecretName   string
	SecurityContext     string
	ContainerResources  string
	NodeSelector        string
	ConfVolume          string
	CollectAddon        string
	CollectVolume       string
	BadgerAddon         string
	PgbackrestEnvVars   string
	PgbackrestS3EnvVars string
	PgmonitorEnvVars    string
	ScopeLabel          string
	//next 2 are for the replica deployment only
	Replicas    string
	PrimaryHost string
	// PgBouncer deployment only
	PgbouncerPass            string
	IsInit                   bool
	EnableCrunchyadm         bool
	ReplicaReinitOnStartFail bool
	PodAntiAffinity          string
	SyncReplication          bool
}

type PostgresHaTemplateFields struct {
	LogStatement            string
	LogMinDurationStatement string
	ArchiveTimeout          string
}

//consolidate with cluster.GetPgbackrestEnvVars
func GetPgbackrestEnvVars(backrestEnabled, clusterName, depName, port, storageType string) string {
	if backrestEnabled == "true" {
		fields := PgbackrestEnvVarsTemplateFields{
			PgbackrestStanza:            "db",
			PgbackrestRepo1Host:         clusterName + "-backrest-shared-repo",
			PgbackrestRepo1Path:         "/backrestrepo/" + clusterName + "-backrest-shared-repo",
			PgbackrestDBPath:            "/pgdata/" + depName,
			PgbackrestPGPort:            port,
			PgbackrestRepo1Type:         GetRepoType(storageType),
			PgbackrestLocalAndS3Storage: IsLocalAndS3Storage(storageType),
		}

		var doc bytes.Buffer
		err := config.PgbackrestEnvVarsTemplate.Execute(&doc, fields)
		if err != nil {
			log.Error(err.Error())
			return ""
		}
		return doc.String()
	}
	return ""

}

func GetBadgerAddon(clientset *kubernetes.Clientset, namespace string, cluster *crv1.Pgcluster, pgbadger_target string) string {

	spec := cluster.Spec

	if cluster.Labels[config.LABEL_BADGER] == "true" {
		log.Debug("crunchy_badger was found as a label on cluster create")
		badgerTemplateFields := badgerTemplateFields{}
		badgerTemplateFields.CCPImageTag = spec.CCPImageTag
		badgerTemplateFields.BadgerTarget = pgbadger_target
		badgerTemplateFields.PGBadgerPort = spec.PGBadgerPort
		badgerTemplateFields.CCPImagePrefix = Pgo.Cluster.CCPImagePrefix
		badgerTemplateFields.ContainerResources = ""

		if Pgo.DefaultBadgerResources != "" {
			tmp, err := Pgo.GetContainerResource(Pgo.DefaultBadgerResources)
			if err != nil {
				log.Error(err)
				return ""
			}
			badgerTemplateFields.ContainerResources = GetContainerResourcesJSON(&tmp)

		}

		var badgerDoc bytes.Buffer
		err := config.BadgerTemplate.Execute(&badgerDoc, badgerTemplateFields)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		if CRUNCHY_DEBUG {
			config.BadgerTemplate.Execute(os.Stdout, badgerTemplateFields)
		}
		return badgerDoc.String()
	}
	return ""
}

func GetCollectAddon(clientset *kubernetes.Clientset, namespace string, spec *crv1.PgclusterSpec) string {

	if spec.UserLabels[config.LABEL_COLLECT] == "true" {
		log.Debug("crunchy_collect was found as a label on cluster create")

		log.Debug("creating collect secret for cluster %s", spec.Name)
		err := util.CreateSecret(clientset, spec.Name, spec.CollectSecretName, config.LABEL_COLLECT_PG_USER,
			Pgo.Cluster.PgmonitorPassword, namespace)

		collectTemplateFields := collectTemplateFields{}
		collectTemplateFields.Name = spec.Name
		collectTemplateFields.JobName = spec.Name
		collectTemplateFields.CCPImageTag = spec.CCPImageTag
		collectTemplateFields.ExporterPort = spec.ExporterPort
		collectTemplateFields.CCPImagePrefix = Pgo.Cluster.CCPImagePrefix
		collectTemplateFields.PgPort = spec.Port

		var collectDoc bytes.Buffer
		err = config.CollectTemplate.Execute(&collectDoc, collectTemplateFields)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		if CRUNCHY_DEBUG {
			config.CollectTemplate.Execute(os.Stdout, collectTemplateFields)
		}
		return collectDoc.String()
	}
	return ""
}

//consolidate with cluster.GetConfVolume
func GetConfVolume(clientset *kubernetes.Clientset, cl *crv1.Pgcluster, namespace string) string {
	var found bool
	var configMapStr string

	//check for global custom configmap "pgo-custom-pg-config"
	_, found = kubeapi.GetConfigMap(clientset, config.GLOBAL_CUSTOM_CONFIGMAP, PgoNamespace)
	if found {
		configMapStr = "\"pgo-custom-pg-config\""
	} else {
		log.Debug(config.GLOBAL_CUSTOM_CONFIGMAP + " was not found, skipping global configMap")

		//check for user provided configmap
		if cl.Spec.CustomConfig != "" {
			_, found = kubeapi.GetConfigMap(clientset, cl.Spec.CustomConfig, namespace)
			if !found {
				//you should NOT get this error because of apiserver validation of this value!
				log.Errorf("%s was not found, error, skipping user provided configMap", cl.Spec.CustomConfig)
			} else {
				log.Debugf("user provided configmap %s was used for this cluster", cl.Spec.CustomConfig)
				configMapStr = "\"" + cl.Spec.CustomConfig + "\""
			}
		}
	}

	return configMapStr
}

// Creates a configMap containing 'crunchy-postgres-ha' configuration settings. aA default crunchy-postgres-ha
// configuration file is included if a default config file is not providing using a custom configMap.
func AddDefaultPostgresHaConfigMap(clientset *kubernetes.Clientset, cluster *crv1.Pgcluster, isInit, createDefaultPghaConf bool,
	namespace string) error {

	data := make(map[string]string)

	labels := make(map[string]string)
	labels[config.LABEL_VENDOR] = config.LABEL_CRUNCHY
	labels[config.LABEL_PG_CLUSTER] = cluster.Name
	labels[config.LABEL_PGHA_DEFAULT_CONFIGMAP] = "true"

	if isInit && createDefaultPghaConf {
		var archiveTimeout string

		if _, exists := cluster.Spec.UserLabels[config.LABEL_ARCHIVE_TIMEOUT]; !exists {
			archiveTimeout = cluster.Spec.UserLabels[config.LABEL_ARCHIVE_TIMEOUT]
		} else {
			archiveTimeout = DefaultArchiveTimeout
		}

		postgresHaFields := PostgresHaTemplateFields{
			LogStatement:            Pgo.Cluster.LogStatement,
			LogMinDurationStatement: Pgo.Cluster.LogMinDurationStatement,
			ArchiveTimeout:          archiveTimeout,
		}

		var postgresHaConfig bytes.Buffer

		err := config.PostgresHaTemplate.Execute(&postgresHaConfig, postgresHaFields)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		data[config.PostgresHaTemplatePath] = postgresHaConfig.String()
	}

	if isInit {
		data["init"] = "true"
	} else {
		data["init"] = "false"
	}

	configmap := &v1.ConfigMap{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   cluster.Name + "-" + DefaultPgHaConfigMapSuffix,
			Labels: labels,
		},
		Data: data,
	}

	err := kubeapi.CreateConfigMap(clientset, configmap, namespace)
	if err != nil {
		return err
	}

	return nil
}

// sets the proper collect secret in the deployment spec if collect is enabled
func GetCollectVolume(clientset *kubernetes.Clientset, cl *crv1.Pgcluster, namespace string) string {
	if cl.Spec.UserLabels[config.LABEL_COLLECT] == "true" {
		return "\"secret\": { \"secretName\": \"" + cl.Spec.CollectSecretName + "\" }"
	}

	return "\"emptyDir\": { \"secretName\": \"Memory\" }"
}

// needs to be consolidated with cluster.GetLabelsFromMap
// GetLabelsFromMap ...
func GetLabelsFromMap(labels map[string]string) string {
	var output string

	for key, value := range labels {
		if len(validation.IsQualifiedName(key)) == 0 && len(validation.IsValidLabelValue(value)) == 0 {
			output += fmt.Sprintf("\"%s\": \"%s\",", key, value)
		}
	}
	// removing the trailing comma from the final label
	return strings.TrimSuffix(output, ",")
}

// GetPrimaryLabels ...
/**
func GetPrimaryLabels(serviceName string, ClusterName string, replicaFlag bool, userLabels map[string]string) map[string]string {
	primaryLabels := make(map[string]string)

	primaryLabels["name"] = serviceName
	primaryLabels[config.LABEL_PG_CLUSTER] = ClusterName

	for key, value := range userLabels {
		if key == config.LABEL_PGBOUNCER {
			//these dont apply to a primary or replica
		} else if key == config.LABEL_AUTOFAIL || key == config.LABEL_NODE_LABEL_KEY || key == config.LABEL_NODE_LABEL_VALUE ||
			key == config.LABEL_BACKREST_STORAGE_TYPE {
			//dont add these since they can break label expression checks
			//or autofail toggling
		} else {
			log.Debugf("JEFF label copying XXX key=%s value=%s", key, value)
			primaryLabels[key] = value
		}
	}

	return primaryLabels
}
*/

// GetAffinity ...
func GetAffinity(nodeLabelKey, nodeLabelValue string, affoperator string) string {
	log.Debugf("GetAffinity with nodeLabelKey=[%s] nodeLabelKey=[%s] and operator=[%s]\n", nodeLabelKey, nodeLabelValue, affoperator)
	output := ""
	if nodeLabelKey == "" {
		return output
	}

	affinityTemplateFields := affinityTemplateFields{}
	affinityTemplateFields.NodeLabelKey = nodeLabelKey
	affinityTemplateFields.NodeLabelValue = nodeLabelValue
	affinityTemplateFields.OperatorValue = affoperator

	var affinityDoc bytes.Buffer
	err := config.AffinityTemplate.Execute(&affinityDoc, affinityTemplateFields)
	if err != nil {
		log.Error(err.Error())
		return output
	}

	if CRUNCHY_DEBUG {
		config.AffinityTemplate.Execute(os.Stdout, affinityTemplateFields)
	}

	return affinityDoc.String()
}

// GetReplicaAffinity ...
// use NotIn as an operator when a node-label is not specified on the
// replica, use the node labels from the primary in this case
// use In as an operator when a node-label is specified on the replica
// use the node labels from the replica in this case
func GetReplicaAffinity(clusterLabels, replicaLabels map[string]string) string {
	var operator, key, value string
	log.Debug("GetReplicaAffinity ")
	if replicaLabels[config.LABEL_NODE_LABEL_KEY] != "" {
		//use the replica labels
		operator = "In"
		key = replicaLabels[config.LABEL_NODE_LABEL_KEY]
		value = replicaLabels[config.LABEL_NODE_LABEL_VALUE]
	} else {
		//use the cluster labels
		operator = "NotIn"
		key = clusterLabels[config.LABEL_NODE_LABEL_KEY]
		value = clusterLabels[config.LABEL_NODE_LABEL_VALUE]
	}
	return GetAffinity(key, value, operator)
}

// GetPodAntiAffinity returns the populated pod anti-affinity json that should be attached to
// the various pods comprising the pg cluster
func GetPodAntiAffinity(podAntiAffinityType string, clusterName string) string {

	log.Debugf("GetPodAnitAffinity with clusterName=[%s]", clusterName)

	// get the PodAntiAffinity type from the CR parameter (podAntiAffinityType) or from the
	// pgo.yaml, depending on whether or not either have been set
	var affinityTypeParam crv1.PodAntiAffinityType
	if podAntiAffinityType != "" {
		affinityTypeParam = crv1.PodAntiAffinityType(podAntiAffinityType)
	} else if Pgo.Cluster.PodAntiAffinity != "" {
		affinityTypeParam = crv1.PodAntiAffinityType(Pgo.Cluster.PodAntiAffinity)
	}

	// verify that the affinity type provided is valid (i.e. 'required' or 'preffered'), and
	// log an error and return an empty string if not
	if err := affinityTypeParam.Validate(); affinityTypeParam != "" &&
		err != nil {
		log.Error(fmt.Sprintf("Invalid affinity type '%s' specified when attempting to set "+
			"default pod anti-affinity for cluster %s.  Pod anti-affinity will not be applied.",
			podAntiAffinityType, clusterName))
		return ""
	}

	// set requiredDuringSchedulingIgnoredDuringExecution or
	// prefferedDuringSchedulingIgnoredDuringExecution depending on the pod anti-affinity type
	// specified in the pgcluster CR.  Defaults to preffered if not explicitly specified
	// in the CR or in the pgo.yaml configuration file
	templateAffinityType := preferScheduleIgnoreExec
	switch affinityTypeParam {
	case crv1.PodAntiAffinityDisabled: // if disabled return an empty string
		log.Debugf("Default pod anti-affinity disabled for clusterName=[%s]", clusterName)
		return ""
	case crv1.PodAntiAffinityRequired:
		templateAffinityType = requireScheduleIgnoreExec
	}

	podAntiAffinityTemplateFields := podAntiAffinityTemplateFields{
		AffinityType:            templateAffinityType,
		ClusterName:             clusterName,
		VendorLabelKey:          config.LABEL_VENDOR,
		VendorLabelValue:        config.LABEL_CRUNCHY,
		PodAntiAffinityLabelKey: config.LABEL_POD_ANTI_AFFINITY,
	}

	var podAntiAffinityDoc bytes.Buffer
	err := config.PodAntiAffinityTemplate.Execute(&podAntiAffinityDoc,
		podAntiAffinityTemplateFields)
	if err != nil {
		log.Error(err.Error())
		return ""
	}

	if CRUNCHY_DEBUG {
		config.PodAntiAffinityTemplate.Execute(os.Stdout, podAntiAffinityTemplateFields)
	}

	return podAntiAffinityDoc.String()
}

func GetPgmonitorEnvVars(metricsEnabled string) string {
	if metricsEnabled == "true" {
		fields := PgmonitorEnvVarsTemplateFields{
			PgmonitorPassword: Pgo.Cluster.PgmonitorPassword,
		}

		var doc bytes.Buffer
		err := config.PgmonitorEnvVarsTemplate.Execute(&doc, fields)
		if err != nil {
			log.Error(err.Error())
			return ""
		}
		return doc.String()
	}
	return ""

}

// GetPgbackrestS3EnvVars retrieves the values for the various configuration settings require to
// configure pgBackRest for AWS S3, inlcuding a bucket, endpoint, region, key and key secret.
// The bucket, endpoint & region are obtained from the associated parameters in the pgcluster
// CR, while the key and key secret are obtained from the backrest repository secret.  Once these
// values have been obtained, they are used to populate a template containing the various
// pgBackRest environment variables required to enable S3 support.  After the template has been
// executed with the proper values, the result is then returned a string for inclusion in the PG
// and pgBackRest deployments.
func GetPgbackrestS3EnvVars(cluster crv1.Pgcluster, clientset *kubernetes.Clientset,
	ns string) string {

	if cluster.Labels[config.LABEL_BACKREST] == "true" &&
		strings.Contains(cluster.Spec.UserLabels[config.LABEL_BACKREST_STORAGE_TYPE], "s3") {

		// populate the S3 bucket, endpoint and region using either the values in the pgcluster
		// spec (if present), otherwise populate using the values from the pgo.yaml config file
		s3EnvVars := PgbackrestS3EnvVarsTemplateFields{}
		if cluster.Spec.BackrestS3Bucket != "" {
			s3EnvVars.PgbackrestS3Bucket = cluster.Spec.BackrestS3Bucket
		} else {
			s3EnvVars.PgbackrestS3Bucket = Pgo.Cluster.BackrestS3Bucket
		}

		if cluster.Spec.BackrestS3Endpoint != "" {
			s3EnvVars.PgbackrestS3Endpoint = cluster.Spec.BackrestS3Endpoint
		} else {
			s3EnvVars.PgbackrestS3Endpoint = Pgo.Cluster.BackrestS3Endpoint
		}

		if cluster.Spec.BackrestS3Region != "" {
			s3EnvVars.PgbackrestS3Region = cluster.Spec.BackrestS3Region
		} else {
			s3EnvVars.PgbackrestS3Region = Pgo.Cluster.BackrestS3Region
		}

		secret, secretExists, err := kubeapi.GetSecret(clientset,
			cluster.Name+"-backrest-repo-config", ns)
		if err != nil {
			log.Error(err.Error())
			return ""
		} else if !secretExists {
			log.Errorf("Secret '%s-backrest-repo-config' does not exist. Unable to set S3 env vars "+
				"for pgBackRest", cluster.Name)
			return ""
		}

		type keyData struct {
			Key       string `yaml:"aws-s3-key"`
			KeySecret string `yaml:"aws-s3-key-secret"`
		}
		clusterKeyData := keyData{}
		pgoKeyData := keyData{}

		err = yaml.Unmarshal(secret.Data["aws-s3-credentials.yaml"], &clusterKeyData)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		// if key or key secret no inlcuded in cluster secret, check global secret
		if clusterKeyData.Key == "" || clusterKeyData.KeySecret == "" {
			secret, secretExists, err := kubeapi.GetSecret(clientset, "pgo-backrest-repo-config",
				PgoNamespace)
			if err != nil {
				log.Error(err.Error())
				return ""
			} else if !secretExists {
				log.Errorf("Secret 'pgo-backrest-repo-config' does not exist. Unable to set S3 env vars " +
					"for pgBackRest")
				return ""
			}
			err = yaml.Unmarshal(secret.Data["aws-s3-credentials.yaml"], &pgoKeyData)
			if err != nil {
				log.Error(err.Error())
				return ""
			}
		}

		// set the key and key secret using either the parameters from the cluster spec,
		// or if not present, the parameters from pgo.yaml
		if clusterKeyData.Key != "" {
			s3EnvVars.PgbackrestS3Key = clusterKeyData.Key
		} else if pgoKeyData.Key != "" {
			s3EnvVars.PgbackrestS3Key = pgoKeyData.Key
		}
		if clusterKeyData.KeySecret != "" {
			s3EnvVars.PgbackrestS3KeySecret = clusterKeyData.KeySecret
		} else if pgoKeyData.Key != "" {
			s3EnvVars.PgbackrestS3KeySecret = pgoKeyData.KeySecret
		}

		var b bytes.Buffer
		err = config.PgbackrestS3EnvVarsTemplate.Execute(&b, s3EnvVars)
		if err != nil {
			log.Error(err.Error())
			return ""
		}

		return b.String()
	}
	return ""
}

// UpdatePghaDefaultConfigInitFlag sets the init value for the pgha config file to true or false depending on the vlaue
// provided
func UpdatePghaDefaultConfigInitFlag(clientset *kubernetes.Clientset, initVal bool, clusterName, namespace string) {

	log.Debugf("cluster %s has been initialized, updating init value in default pgha configMap "+
		"to prevent future bootstrap attempts", clusterName)
	selector := config.LABEL_PG_CLUSTER + "=" + clusterName + "," + config.LABEL_PGHA_DEFAULT_CONFIGMAP + "=true"
	configMapList, found := kubeapi.ListConfigMap(clientset, selector, namespace)
	if !found {
		log.Errorf("unable to find the default pgha configMap found for cluster %s using selector %s, unable to set "+
			"init value to false", clusterName, selector)
	} else if len(configMapList.Items) > 1 {
		log.Errorf("more than one default pgha configMap found for cluster %s using selector %s, unable to set "+
			"init value to false", clusterName, selector)
	}
	configMap := &configMapList.Items[0]
	configMap.Data["init"] = strconv.FormatBool(initVal)

	kubeapi.UpdateConfigMap(clientset, configMap, namespace)
}

// GetSyncReplication returns true if synchronous replication has been enabled using either the
// pgcluster CR specification or the pgo.yaml configuration file.  Otherwise, if synchronous
// mode has not been enabled, it returns false.
func GetSyncReplication(specSyncReplication *bool) bool {
	// alawys use the value from the CR if explicitly provided
	if specSyncReplication != nil {
		return *specSyncReplication
	} else if Pgo.Cluster.SyncReplication {
		return true
	}
	return false
}
