/*
Copyright 2020 The CRDS Authors.

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

package crd

import (
	"testing"
)

var _ Modifier = StripLabels()
var _ Modifier = StripAnnotations()
var _ Modifier = StripConversion()

var v1crd = []byte(`
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: crontabs.example.com
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: example.com
  # list of versions supported by this CustomResourceDefinition
  versions:
  - name: v1beta1
    # Each version can be enabled/disabled by Served flag.
    served: true
    # One and only one version must be marked as the storage version.
    storage: true
    # A schema is required
    schema:
      openAPIV3Schema:
        type: object
        properties:
          host:
            type: string
          port:
            type: string
  - name: v1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          host:
            type: string
          port:
            type: string
  # The conversion section is introduced in Kubernetes 1.13+ with a default value of
  # None conversion (strategy sub-field set to None).
  conversion:
    # None conversion assumes the same schema for all versions and only sets the apiVersion
    # field of custom resources to the proper value
    strategy: None
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: crontabs
    # singular name to be used as an alias on the CLI and for display
    singular: crontab
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CronTab
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - ct
`)

var v1beta1crd = []byte(`
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  # name must match the spec fields below, and be in the form: <plural>.<group>
  name: crontabs.example.com
spec:
  # group name to use for REST API: /apis/<group>/<version>
  group: example.com
  # prunes object fields that are not specified in OpenAPI schemas below.
  preserveUnknownFields: false
  # list of versions supported by this CustomResourceDefinition
  versions:
  - name: v1beta1
    # Each version can be enabled/disabled by Served flag.
    served: true
    # One and only one version must be marked as the storage version.
    storage: false
    # Each version can define it's own schema when there is no top-level
    # schema is defined.
    schema:
      openAPIV3Schema:
        type: object
        properties:
          hostPort:
            type: object
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          host:
            type: string
          port:
            type: string
  conversion:
    # a Webhook strategy instruct API server to call an external webhook for any conversion between custom resources.
    strategy: Webhook
    # webhookClientConfig is required when strategy is Webhook and it configures the webhook endpoint to be called by API server.
    webhookClientConfig:
      service:
        namespace: default
        name: example-conversion-webhook-server
        path: /crdconvert
      caBundle: "yblHpwAZDeohdgTqQDAVyg=="
  # either Namespaced or Cluster
  scope: Namespaced
  names:
    # plural name to be used in the URL: /apis/<group>/<version>/<plural>
    plural: crontabs
    # singular name to be used as an alias on the CLI and for display
    singular: crontab
    # kind is normally the CamelCased singular type. Your resource manifests use this.
    kind: CronTab
    # shortNames allow shorter string to match your resource on the CLI
    shortNames:
    - ct
`)

var crossplane = []byte(`
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: cloudmemorystoreinstanceclasses.cache.gcp.crossplane.io
spec:
  additionalPrinterColumns:
  - JSONPath: .specTemplate.providerRef.name
    name: PROVIDER-REF
    type: string
  - JSONPath: .specTemplate.reclaimPolicy
    name: RECLAIM-POLICY
    type: string
  - JSONPath: .metadata.creationTimestamp
    name: AGE
    type: date
  group: cache.gcp.crossplane.io
  names:
    kind: CloudMemorystoreInstanceClass
    listKind: CloudMemorystoreInstanceClassList
    plural: cloudmemorystoreinstanceclasses
    singular: cloudmemorystoreinstanceclass
  scope: ""
  subresources: {}
  validation:
    openAPIV3Schema:
      description: A CloudMemorystoreInstanceClass is a non-portable resource class.
        It defines the desired spec of resource claims that use it to dynamically
        provision a managed resource.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        specTemplate:
          description: SpecTemplate is a template for the spec of a dynamically provisioned
            CloudMemorystoreInstance.
          properties:
            alternativeLocationId:
              description: AlternativeLocationID is only applicable to STANDARD_HA
                tier, which protects the instance against zonal failures by provisioning
                it across two zones. If provided, it must be a different zone from
                the one provided in locationId.
              type: string
            authorizedNetwork:
              description: AuthorizedNetwork specifies the full name of the Google
                Compute Engine network to which the instance is connected. If left
                unspecified, the default network will be used.
              type: string
            locationId:
              description: LocationID specifies the zone where the instance will be
                provisioned. If not provided, the service will choose a zone for the
                instance. For STANDARD_HA tier, instances will be created across two
                zones for protection against zonal failures.
              type: string
            memorySizeGb:
              description: MemorySizeGB specifies the Redis memory size in GiB.
              type: integer
            providerRef:
              description: ProviderReference specifies the provider that will be used
                to create, observe, update, and delete managed resources that are
                dynamically provisioned using this resource class.
              properties:
                apiVersion:
                  description: API version of the referent.
                  type: string
                fieldPath:
                  description: 'If referring to a piece of an object instead of an
                    entire object, this string should contain a valid JSON/Go field
                    access statement, such as desiredState.manifest.containers[2].
                    For example, if the object reference is to a container within
                    a pod, this would take on a value like: "spec.containers{name}"
                    (where "name" refers to the name of the container that triggered
                    the event) or if no container name is specified "spec.containers[2]"
                    (container with index 2 in this pod). This syntax is chosen only
                    to have some well-defined way of referencing a part of an object.
                    TODO: this design is not final and this field is subject to change
                    in the future.'
                  type: string
                kind:
                  description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                  type: string
                name:
                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names'
                  type: string
                namespace:
                  description: 'Namespace of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/'
                  type: string
                resourceVersion:
                  description: 'Specific resourceVersion to which this reference is
                    made, if any. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                  type: string
                uid:
                  description: 'UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids'
                  type: string
              type: object
            reclaimPolicy:
              description: ReclaimPolicy specifies what will happen to external resources
                when managed resources dynamically provisioned using this resource
                class are deleted. "Delete" deletes the external resource, while "Retain"
                (the default) does not. Note this behaviour is subtly different from
                other uses of the ReclaimPolicy concept within the Kubernetes ecosystem
                per https://github.com/crossplaneio/crossplane-runtime/issues/21
              type: string
            redisConfigs:
              additionalProperties:
                type: string
              description: 'RedisConfigs specifies Redis configuration parameters,
                according to http://redis.io/topics/config. Currently, the only supported
                parameters are: * maxmemory-policy * notify-keyspace-events'
              type: object
            redisVersion:
              description: RedisVersion specifies the version of Redis software. If
                not provided, latest supported version will be used. Updating the
                version will perform an upgrade/downgrade to the new version. Currently,
                the supported values are REDIS_3_2 for Redis 3.2, and REDIS_4_0 for
                Redis 4.0 (the default).
              enum:
              - REDIS_3_2
              - REDIS_4_0
              type: string
            region:
              description: Region in which to create this Cloud Memorystore cluster.
              type: string
            reservedIpRange:
              description: ReservedIPRange specifies the CIDR range of internal addresses
                that are reserved for this instance. If not provided, the service
                will choose an unused /29 block, for example, 10.0.0.0/29 or 192.168.0.0/29.
                Ranges must be unique and non-overlapping with existing subnets in
                an authorized network.
              type: string
            tier:
              description: Tier specifies the replication level of the Redis cluster.
                BASIC provides a single Redis instance with no high availability.
                STANDARD_HA provides a cluster of two Redis instances in distinct
                availability zones. https://cloud.google.com/memorystore/docs/redis/redis-tiers
              enum:
              - BASIC
              - STANDARD_HA
              type: string
          required:
          - memorySizeGb
          - providerRef
          - region
          - tier
          type: object
      required:
      - specTemplate
      type: object
  version: v1alpha2
  versions:
  - name: v1alpha2
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)

var a = []byte(`
apiVersion: example.com/v1
kind: CronTab
metadata:
  name: my-new-cron-object
  namespace: hi
hostPort: a
port: a
host: a
`)

var b = []byte(`
---
apiVersion: cache.gcp.crossplane.io/v1alpha2
kind: CloudMemorystoreInstanceClass
metadata:
  name: gcp-redis-standard
  namespace: gcp-infra-dev
specTemplate:
  tier: STANDARD_HA
  region: us-west2
  memorySizeGb: 1
  providerRef:
    name: example
    namespace: gcp-infra-dev
  reclaimPolicy: Delete
`)

func TestValidate(t *testing.T) {
	cases := []struct {
		name        string
		crd         []byte
		instance    []byte
		expectedErr bool
	}{
		{
			name:        "v1 invalid",
			crd:         v1crd,
			instance:    a,
			expectedErr: true,
		},
		{
			name:        "v1beta1 valid",
			crd:         v1beta1crd,
			instance:    a,
			expectedErr: false,
		},
		{
			name:        "crossplane valid",
			crd:         crossplane,
			instance:    b,
			expectedErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewCRDer(tc.crd)
			if err != nil {
				t.Errorf("Failed to create CRDer: %s", err)
			}
			if err := c.Validate(tc.instance); err != nil && !tc.expectedErr {
				t.Errorf("Unexpected validation error: %s", err)
			}

		})
	}
}
