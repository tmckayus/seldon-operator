/*
Copyright 2019 The Seldon Team.

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

package v1alpha2

import (
	"crypto/md5"
	"encoding/hex"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"strings"
)

const (
	Label_seldon_id  = "seldon-deployment-id"
	Label_seldon_app = "seldon-app"
	Label_svc_orch   = "seldon-deployment-contains-svcorch"

	PODINFO_VOLUME_NAME = "podinfo"
	PODINFO_VOLUME_PATH = "/etc/podinfo"

	ENV_PREDICTIVE_UNIT_SERVICE_PORT = "PREDICTIVE_UNIT_SERVICE_PORT"
	ENV_PREDICTIVE_UNIT_PARAMETERS   = "PREDICTIVE_UNIT_PARAMETERS"
	ENV_PREDICTIVE_UNIT_ID           = "PREDICTIVE_UNIT_ID"
	ENV_PREDICTOR_ID                 = "PREDICTOR_ID"
	ENV_SELDON_DEPLOYMENT_ID         = "SELDON_DEPLOYMENT_ID"

	ANNOTATION_JAVA_OPTS       = "seldon.io/engine-java-opts"
	ANNOTATION_SEPARATE_ENGINE = "seldon.io/engine-separate-pod"
	ANNOTATION_HEADLESS_SVC    = "seldon.io/headless-svc"
)

func hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func containerHash(podSpec *SeldonPodSpec) string {
	s := []string{}
	for i := 0; i < len(podSpec.Spec.Containers); i++ {
		c := podSpec.Spec.Containers[i]
		s = append(s, c.Name)
		s = append(s, c.Image)
	}
	key := strings.Join(s, ":") + ";"
	return hash(key)[:7]
}

func createPredictorHash(p *PredictorSpec) string {
	s := []string{}
	for i := 0; i < len(p.ComponentSpecs); i++ {
		s = append(s, containerHash(p.ComponentSpecs[i]))
	}
	key := strings.Join(s, ",") + ","
	return hash(key)[:7]
}

func GetSeldonDeploymentName(mlDep *SeldonDeployment) string {
	name := mlDep.Spec.Name + "-" + mlDep.ObjectMeta.Name
	if len(name) > 63 {
		return "seldon-" + hash(name)
	} else {
		return name
	}
}

func GetDeploymentName(mlDep *SeldonDeployment, predictorSpec PredictorSpec, podSpec *SeldonPodSpec) string {
	if len(podSpec.Metadata.Name) != 0 {
		return podSpec.Metadata.Name
	} else {
		name := mlDep.Spec.Name + "-" + predictorSpec.Name + "-" + containerHash(podSpec)
		if len(name) > 63 {
			return "seldon-" + hash(name)
		} else {
			return name
		}
	}
}

func GetServiceOrchestratorName(mlDep *SeldonDeployment, p *PredictorSpec) string {
	svcOrchName := mlDep.Spec.Name + "-" + p.Name + "-svc-orch" + "-" + createPredictorHash(p)
	if len(svcOrchName) > 63 {
		return "seldon-" + hash(svcOrchName)
	} else {
		return svcOrchName
	}
}

func GetPredictorKey(mlDep *SeldonDeployment, p *PredictorSpec) string {
	pName := mlDep.Spec.Name + "-" + p.Name
	if len(pName) > 63 {
		return "seldon-" + hash(pName)
	} else {
		return pName
	}
}

func GetPredictorServiceNameKey(c *v1.Container) string {
	return Label_seldon_app + "-" + c.Name
}

func GetPredcitiveUnit(pu *PredictiveUnit, name string) *PredictiveUnit {
	if name == pu.Name {
		return pu
	} else {
		for i := 0; i < len(pu.Children); i++ {
			found := GetPredcitiveUnit(&pu.Children[i], name)
			if found != nil {
				return found
			}
		}
		return nil
	}
}

func cleanContainerName(name string) string {
	var re = regexp.MustCompile("[^-a-z0-9]")
	return re.ReplaceAllString(strings.ToLower(name), "-")
}

func GetContainerServiceName(mlDep *SeldonDeployment, predictorSpec PredictorSpec, c *v1.Container) string {
	containerImageName := cleanContainerName(c.Image)
	svcName := mlDep.Spec.Name + "-" + predictorSpec.Name + "-" + c.Name + "-" + containerImageName
	if len(svcName) > 63 {
		svcName = "seldon" + "-" + containerImageName + "-" + hash(svcName)
		if len(svcName) > 63 {
			return "seldon-" + hash(svcName)
		} else {
			return svcName
		}
	} else {
		return svcName
	}
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SeldonDeploymentSpec defines the desired state of SeldonDeployment
type SeldonDeploymentSpec struct {
	Name        string            `json:"name,omitempty" protobuf:"string,1,opt,name=name"`
	Predictors  []PredictorSpec   `json:"predictors,omitempty" protobuf:"bytes,2,opt,name=name"`
	OauthKey    string            `json:"oauth_key,omitempty" protobuf:"string,3,opt,name=oauth_key"`
	OauthSecret string            `json:"oauth_secret,omitempty" protobuf:"string,4,opt,name=oauth_secret"`
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,5,opt,name=annotations"`
}

type PredictorSpec struct {
	Name            string                  `json:"name,omitempty" protobuf:"string,1,opt,name=name"`
	Graph           *PredictiveUnit         `json:"graph,omitempty" protobuf:"bytes,2,opt,name=predictiveUnit"`
	ComponentSpecs  []*SeldonPodSpec        `json:"componentSpecs,omitempty" protobuf:"bytes,3,opt,name=componentSpecs"`
	Replicas        int32                   `json:"replicas,omitempty" protobuf:"string,4,opt,name=replicas"`
	Annotations     map[string]string       `json:"annotations,omitempty" protobuf:"bytes,5,opt,name=annotations"`
	EngineResources v1.ResourceRequirements `json:"engineResources,omitempty" protobuf:"bytes,6,opt,name=engineResources"`
	Labels          map[string]string       `json:"labels,omitempty" protobuf:"bytes,7,opt,name=labels"`
	SvcOrchSpec     SvcOrchSpec             `json:"svcOrchSpec,omitempty" protobuf:"bytes,8,opt,name=svcOrchSpec"`
	Traffic         int32                   `json:"traffic,omitempty" protobuf:"bytes,9,opt,name=traffic"`
}

type SvcOrchSpec struct {
	Resources *v1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,1,opt,name=resources"`
	Env       []*v1.EnvVar             `json:"env,omitempty" protobuf:"bytes,2,opt,name=env"`
}

type SeldonPodSpec struct {
	Metadata metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec     v1.PodSpec        `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	HpaSpec  *SeldonHpaSpec    `json:"hpaSpec,omitempty" protobuf:"bytes,3,opt,name=hpaSpec"`
}

type SeldonHpaSpec struct {
	MinReplicas *int32                          `json:"minReplicas,omitempty" protobuf:"int,1,opt,name=minReplicas"`
	MaxReplicas int32                           `json:"maxReplicas,omitempty" protobuf:"int,2,opt,name=maxReplicas"`
	Metrics     []autoscalingv2beta2.MetricSpec `json:"metrics,omitempty" protobuf:"bytes,3,opt,name=metrics"`
}

type PredictiveUnitType string

const (
	UNKNOWN_TYPE       PredictiveUnitType = "UNKNOWN_TYPE"
	ROUTER             PredictiveUnitType = "ROUTER"
	COMBINER           PredictiveUnitType = "COMBINER"
	MODEL              PredictiveUnitType = "MODEL"
	TRANSFORMER        PredictiveUnitType = "TRANSFORMER"
	OUTPUT_TRANSFORMER PredictiveUnitType = "OUTPUT_TRANSFORMER"
)

type PredictiveUnitImplementation string

const (
	UNKNOWN_IMPLEMENTATION PredictiveUnitImplementation = "UNKNOWN_IMPLEMENTATION"
	SIMPLE_MODEL           PredictiveUnitImplementation = "SIMPLE_MODEL"
	SIMPLE_ROUTER          PredictiveUnitImplementation = "SIMPLE_ROUTER"
	RANDOM_ABTEST          PredictiveUnitImplementation = "RANDOM_ABTEST"
	AVERAGE_COMBINER       PredictiveUnitImplementation = "AVERAGE_COMBINER"
)

type PredictiveUnitMethod string

const (
	TRANSFORM_INPUT  PredictiveUnitMethod = "TRANSFORM_INPUT"
	TRANSFORM_OUTPUT PredictiveUnitMethod = "TRANSFORM_OUTPUT"
	ROUTE            PredictiveUnitMethod = "ROUTE"
	AGGREGATE        PredictiveUnitMethod = "AGGREGATE"
	SEND_FEEDBACK    PredictiveUnitMethod = "SEND_FEEDBACK"
)

type EndpointType string

const (
	REST EndpointType = "REST"
	GRPC EndpointType = "GRPC"
)

type Endpoint struct {
	ServiceHost string       `json:"service_host,omitempty" protobuf:"string,1,opt,name=service_host"`
	ServicePort int32        `json:"service_port,omitempty" protobuf:"int32,2,opt,name=service_port"`
	Type        EndpointType `json:"type,omitempty" protobuf:"int,3,opt,name=type"`
}

type ParmeterType string

const (
	INT    ParmeterType = "INT"
	FLOAT  ParmeterType = "FLOAT"
	DOUBLE ParmeterType = "DOUBLE"
	STRING ParmeterType = "STRING"
	BOOL   ParmeterType = "BOOL"
)

type Parameter struct {
	Name  string       `json:"name,omitempty" protobuf:"string,1,opt,name=name"`
	Value string       `json:"value,omitempty" protobuf:"string,2,opt,name=value"`
	Type  ParmeterType `json:"type,omitempty" protobuf:"int,3,opt,name=type"`
}

type PredictiveUnit struct {
	Name           string                        `json:"name,omitempty" protobuf:"string,1,opt,name=name"`
	Children       []PredictiveUnit              `json:"children,omitempty" protobuf:"bytes,2,opt,name=children"`
	Type           *PredictiveUnitType           `json:"type,omitempty" protobuf:"int,3,opt,name=type"`
	Implementation *PredictiveUnitImplementation `json:"implementation,omitempty" protobuf:"int,4,opt,name=implementation"`
	Methods        *[]PredictiveUnitMethod       `json:"methods,omitempty" protobuf:"int,5,opt,name=methods"`
	Endpoint       *Endpoint                     `json:"endpoint,omitempty" protobuf:"bytes,6,opt,name=endpoint"`
	Parameters     []Parameter                   `json:"parameters,omitempty" protobuf:"bytes,7,opt,name=parameters"`
}

type DeploymentStatus struct {
	Name              string `json:"name,omitempty" protobuf:"string,1,opt,name=name"`
	Status            string `json:"status,omitempty" protobuf:"string,2,opt,name=status"`
	Description       string `json:"description,omitempty" protobuf:"string,3,opt,name=description"`
	Replicas          int32  `json:"replicas,omitempty" protobuf:"string,4,opt,name=replicas"`
	AvailableReplicas int32  `json:"availableReplicas,omitempty" protobuf:"string,5,opt,name=availableRelicas"`
}

type ServiceStatus struct {
	SvcName      string `json:"svcName,omitempty" protobuf:"string,1,opt,name=svcName"`
	HttpEndpoint string `json:"httpEndpoint,omitempty" protobuf:"string,2,opt,name=httpEndpoint"`
	GrpcEndpoint string `json:"grpcEndpoint,omitempty" protobuf:"string,3,opt,name=grpcEndpoint"`
}

// SeldonDeploymentStatus defines the observed state of SeldonDeployment
type SeldonDeploymentStatus struct {
	State            string                      `json:"state,omitempty" protobuf:"string,1,opt,name=state"`
	Description      string                      `json:"description,omitempty" protobuf:"string,2,opt,name=description"`
	DeploymentStatus map[string]DeploymentStatus `json:"deploymentStatus,omitempty" protobuf:"bytes,3,opt,name=deploymentStatus"`
	ServiceStatus    map[string]ServiceStatus    `json:"serviceStatus,omitempty" protobuf:"bytes,4,opt,name=serviceStatus"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SeldonDeployment is the Schema for the seldondeployments API
// +k8s:openapi-gen=true
// +kubebuilder:resource:shortName=sdep
// +kubebuilder:subresource:status
type SeldonDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SeldonDeploymentSpec   `json:"spec,omitempty"`
	Status SeldonDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SeldonDeploymentList contains a list of SeldonDeployment
type SeldonDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SeldonDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SeldonDeployment{}, &SeldonDeploymentList{})
}
