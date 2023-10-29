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

var zeebecluster = []byte(`
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.11.3
  creationTimestamp: null
  name: zeebeclusters.cloud.camunda.io
spec:
  group: cloud.camunda.io
  names:
    kind: ZeebeCluster
    listKind: ZeebeClusterList
    plural: zeebeclusters
    shortNames:
    - zb
    singular: zeebecluster
  scope: Cluster
  versions:
  - additionalPrinterColumns:
    - jsonPath: .spec.orgId
      name: Org ID
      type: string
    - jsonPath: .status.ready
      name: Overall Health
      type: string
    - jsonPath: .spec.cloud.salesPlan.name
      name: Sales Plan
      type: string
    - jsonPath: .spec.cloud.clusterPlan.name
      name: Cluster Plan
      type: string
    - jsonPath: .spec.zeebe.broker.clusterSize
      name: Brokers
      priority: 1
      type: integer
    - jsonPath: .status.zeebeStatus
      name: Zeebe
      type: string
    - jsonPath: .status.tasklistStatus
      name: Tasklist
      priority: 1
      type: string
    - jsonPath: .status.operateStatus
      name: Operate
      priority: 1
      type: string
    - jsonPath: .status.optimizeStatus
      name: Optimize
      priority: 1
      type: string
    - jsonPath: .spec.cloud.clusterPlanType.name
      name: Cluster Plan Type
      priority: 1
      type: string
    name: v1alpha1
    schema:
      openAPIV3Schema:
        description: ZeebeCluster is the Schema for the zeebeclusters API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ZeebeClusterSpec defines the desired state of ZeebeCluster
            properties:
              backup:
                description: BackupSpec contains the settings for cold-backups
                properties:
                  enabled:
                    description: If set to true cold-backups will be taken during
                      a generation change
                    type: boolean
                  retainedCount:
                    default: 3
                    description: Hot Backups related configuration The count of backups
                      to retain. Only counted for hot backups
                    type: integer
                  ttl:
                    description: 'Todo: Implement support for this in hot backups
                      Retention time of the backup, after this period it will get
                      deleted defaults to 60 days examples: 60m, supports hours (h)
                      minutes (m) and seconds (s)'
                    type: string
                required:
                - enabled
                - ttl
                type: object
              cloud:
                properties:
                  channel:
                    description: RelationSpec shows DB relation metadata. Most relations
                      are added as labels.
                    properties:
                      name:
                        description: Name of the corresponding relation (like generation
                          name for generations)
                        type: string
                      uuid:
                        description: UUID of the relation (like generation uuid for
                          generations)
                        type: string
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  clusterPlan:
                    description: RelationSpec shows DB relation metadata. Most relations
                      are added as labels.
                    properties:
                      name:
                        description: Name of the corresponding relation (like generation
                          name for generations)
                        type: string
                      uuid:
                        description: UUID of the relation (like generation uuid for
                          generations)
                        type: string
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  clusterPlanType:
                    description: RelationSpec shows DB relation metadata. Most relations
                      are added as labels.
                    properties:
                      name:
                        description: Name of the corresponding relation (like generation
                          name for generations)
                        type: string
                      uuid:
                        description: UUID of the relation (like generation uuid for
                          generations)
                        type: string
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  generation:
                    description: RelationSpec shows DB relation metadata. Most relations
                      are added as labels.
                    properties:
                      name:
                        description: Name of the corresponding relation (like generation
                          name for generations)
                        type: string
                      uuid:
                        description: UUID of the relation (like generation uuid for
                          generations)
                        type: string
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  internal:
                    description: Indicates whether Zeebe Cluster is used internally
                    type: boolean
                  salesPlan:
                    description: SalesPlanSpec shows the relation metadata of a sales
                      plan
                    properties:
                      name:
                        description: Name of the Sales Plan
                        type: string
                      type:
                        description: Type of the Sales Plan
                        type: string
                      uuid:
                        description: UUID of the Sales Plan
                        type: string
                    type: object
                type: object
                x-kubernetes-preserve-unknown-fields: true
              connectorBridge:
                properties:
                  backend:
                    description: BackendSpec contains the typical information for
                      a k8s-deployment, it can be reused when creating additional
                      application specs
                    properties:
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      overrideEnv:
                        description: Any var set here will override those provided
                          to the container. Behaviour if duplicate vars are provided
                          _here_ is undefined.
                        items:
                          description: EnvVar represents an environment variable present
                            in a Container.
                          properties:
                            name:
                              description: Name of the environment variable. Must
                                be a C_IDENTIFIER.
                              type: string
                            value:
                              description: 'Variable references $(VAR_NAME) are expanded
                                using the previously defined environment variables
                                in the container and any service environment variables.
                                If a variable cannot be resolved, the reference in
                                the input string will be unchanged. Double $$ are
                                reduced to a single $, which allows for escaping the
                                $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce
                                the string literal "$(VAR_NAME)". Escaped references
                                will never be expanded, regardless of whether the
                                variable exists or not. Defaults to "".'
                              type: string
                            valueFrom:
                              description: Source for the environment variable's value.
                                Cannot be used if value is not empty.
                              properties:
                                configMapKeyRef:
                                  description: Selects a key of a ConfigMap.
                                  properties:
                                    key:
                                      description: The key to select.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap or
                                        its key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                                fieldRef:
                                  description: 'Selects a field of the pod: supports
                                    metadata.name, metadata.namespace, metadata.labels[''<KEY>''],
                                    metadata.annotations[''<KEY>''], spec.nodeName,
                                    spec.serviceAccountName, status.hostIP, status.podIP,
                                    status.podIPs.'
                                  properties:
                                    apiVersion:
                                      description: Version of the schema the FieldPath
                                        is written in terms of, defaults to "v1".
                                      type: string
                                    fieldPath:
                                      description: Path of the field to select in
                                        the specified API version.
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                  x-kubernetes-map-type: atomic
                                resourceFieldRef:
                                  description: 'Selects a resource of the container:
                                    only resources limits and requests (limits.cpu,
                                    limits.memory, limits.ephemeral-storage, requests.cpu,
                                    requests.memory and requests.ephemeral-storage)
                                    are currently supported.'
                                  properties:
                                    containerName:
                                      description: 'Container name: required for volumes,
                                        optional for env vars'
                                      type: string
                                    divisor:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      description: Specifies the output format of
                                        the exposed resources, defaults to "1"
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    resource:
                                      description: 'Required: resource to select'
                                      type: string
                                  required:
                                  - resource
                                  type: object
                                  x-kubernetes-map-type: atomic
                                secretKeyRef:
                                  description: Selects a key of a secret in the pod's
                                    namespace
                                  properties:
                                    key:
                                      description: The key of the secret to select
                                        from.  Must be a valid secret key.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret or its
                                        key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                              type: object
                          required:
                          - name
                          type: object
                        type: array
                      resources:
                        description: ResourceRequirements describes the compute resource
                          requirements.
                        properties:
                          claims:
                            description: "Claims lists the names of resources, defined
                              in spec.resourceClaims, that are used by this container.
                              \n This is an alpha field and requires enabling the
                              DynamicResourceAllocation feature gate. \n This field
                              is immutable. It can only be set for containers."
                            items:
                              description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                              properties:
                                name:
                                  description: Name must match the name of one entry
                                    in pod.spec.resourceClaims of the Pod where this
                                    field is used. It makes that resource available
                                    inside a container.
                                  type: string
                              required:
                              - name
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - name
                            x-kubernetes-list-type: map
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Limits describes the maximum amount of compute
                              resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Requests describes the minimum amount of
                              compute resources required. If Requests is omitted for
                              a container, it defaults to Limits if that is explicitly
                              specified, otherwise to an implementation-defined value.
                              Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                        type: object
                    type: object
                  replicas:
                    description: How many connector bridge deployments to run
                    format: int32
                    minimum: 0
                    type: integer
                type: object
              domain:
                description: The domain of the endpoint e.g. alpha.camunda.io TODO
                  test cases for validation lowercase letters, numbers, and dashes
                  only, ending in a tld
                pattern: ^[a-z0-9.-]+\.[a-z]+$
                type: string
              generationUUID:
                description: UUID of the generation that gets applied
                type: string
              identity:
                description: IdentitySpec Identity config for the ZeebeCluster
                properties:
                  resourcePermissions:
                    description: Enable/Disable resource base permissions for operate
                      for the env var
                    type: boolean
                type: object
              operate:
                properties:
                  alert:
                    properties:
                      m2mAudience:
                        type: string
                      m2mClientId:
                        type: string
                      webhook:
                        type: string
                    type: object
                  auth0:
                    properties:
                      audience:
                        type: string
                      backendDomain:
                        type: string
                      claimName:
                        type: string
                      clientId:
                        description: 'Deprecated: This needs a related secret and
                          comes from the env now'
                        type: string
                      clusterId:
                        description: 'Deprecated: Keep until Tasklist switches over
                          to use CAMUNDA_TASKLIST_CLOUD_CLUSTERID This is not needed
                          IMHO, as it duplicates the information'
                        type: string
                      domain:
                        type: string
                      organizationId:
                        description: 'Deprecated: Use zb.OrgID'
                        type: string
                      resourceserver:
                        description: 'Deprecated: Calculated from Domain'
                        type: string
                    type: object
                  backend:
                    description: BackendSpec contains the typical information for
                      a k8s-deployment, it can be reused when creating additional
                      application specs
                    properties:
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      overrideEnv:
                        description: Any var set here will override those provided
                          to the container. Behaviour if duplicate vars are provided
                          _here_ is undefined.
                        items:
                          description: EnvVar represents an environment variable present
                            in a Container.
                          properties:
                            name:
                              description: Name of the environment variable. Must
                                be a C_IDENTIFIER.
                              type: string
                            value:
                              description: 'Variable references $(VAR_NAME) are expanded
                                using the previously defined environment variables
                                in the container and any service environment variables.
                                If a variable cannot be resolved, the reference in
                                the input string will be unchanged. Double $$ are
                                reduced to a single $, which allows for escaping the
                                $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce
                                the string literal "$(VAR_NAME)". Escaped references
                                will never be expanded, regardless of whether the
                                variable exists or not. Defaults to "".'
                              type: string
                            valueFrom:
                              description: Source for the environment variable's value.
                                Cannot be used if value is not empty.
                              properties:
                                configMapKeyRef:
                                  description: Selects a key of a ConfigMap.
                                  properties:
                                    key:
                                      description: The key to select.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap or
                                        its key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                                fieldRef:
                                  description: 'Selects a field of the pod: supports
                                    metadata.name, metadata.namespace, metadata.labels[''<KEY>''],
                                    metadata.annotations[''<KEY>''], spec.nodeName,
                                    spec.serviceAccountName, status.hostIP, status.podIP,
                                    status.podIPs.'
                                  properties:
                                    apiVersion:
                                      description: Version of the schema the FieldPath
                                        is written in terms of, defaults to "v1".
                                      type: string
                                    fieldPath:
                                      description: Path of the field to select in
                                        the specified API version.
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                  x-kubernetes-map-type: atomic
                                resourceFieldRef:
                                  description: 'Selects a resource of the container:
                                    only resources limits and requests (limits.cpu,
                                    limits.memory, limits.ephemeral-storage, requests.cpu,
                                    requests.memory and requests.ephemeral-storage)
                                    are currently supported.'
                                  properties:
                                    containerName:
                                      description: 'Container name: required for volumes,
                                        optional for env vars'
                                      type: string
                                    divisor:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      description: Specifies the output format of
                                        the exposed resources, defaults to "1"
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    resource:
                                      description: 'Required: resource to select'
                                      type: string
                                  required:
                                  - resource
                                  type: object
                                  x-kubernetes-map-type: atomic
                                secretKeyRef:
                                  description: Selects a key of a secret in the pod's
                                    namespace
                                  properties:
                                    key:
                                      description: The key of the secret to select
                                        from.  Must be a valid secret key.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret or its
                                        key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                              type: object
                          required:
                          - name
                          type: object
                        type: array
                      resources:
                        description: ResourceRequirements describes the compute resource
                          requirements.
                        properties:
                          claims:
                            description: "Claims lists the names of resources, defined
                              in spec.resourceClaims, that are used by this container.
                              \n This is an alpha field and requires enabling the
                              DynamicResourceAllocation feature gate. \n This field
                              is immutable. It can only be set for containers."
                            items:
                              description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                              properties:
                                name:
                                  description: Name must match the name of one entry
                                    in pod.spec.resourceClaims of the Pod where this
                                    field is used. It makes that resource available
                                    inside a container.
                                  type: string
                              required:
                              - name
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - name
                            x-kubernetes-list-type: map
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Limits describes the maximum amount of compute
                              resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Requests describes the minimum amount of
                              compute resources required. If Requests is omitted for
                              a container, it defaults to Limits if that is explicitly
                              specified, otherwise to an implementation-defined value.
                              Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                        type: object
                    type: object
                  dataRetention:
                    default:
                      days: 30
                    description: DataRetentionSpec used in Operate, Tasklist.
                    properties:
                      days:
                        default: 30
                        maximum: 90
                        minimum: 5
                        type: integer
                    type: object
                  elasticsearch:
                    properties:
                      config:
                        description: Configuration spec only used with the elastic-operator
                        properties:
                          nodesCount:
                            default: 1
                            type: integer
                          storage:
                            description: StorageSpecV2 for persistent storage volumes
                              (PVCs)
                            properties:
                              autoResizing:
                                description: Configure Autoresizing
                                properties:
                                  increase:
                                    type: string
                                  threshold:
                                    type: string
                                required:
                                - increase
                                - threshold
                                type: object
                              resources:
                                description: 'Resources represents the minimum resources
                                  the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources'
                                properties:
                                  claims:
                                    description: "Claims lists the names of resources,
                                      defined in spec.resourceClaims, that are used
                                      by this container. \n This is an alpha field
                                      and requires enabling the DynamicResourceAllocation
                                      feature gate. \n This field is immutable. It
                                      can only be set for containers."
                                    items:
                                      description: ResourceClaim references one entry
                                        in PodSpec.ResourceClaims.
                                      properties:
                                        name:
                                          description: Name must match the name of
                                            one entry in pod.spec.resourceClaims of
                                            the Pod where this field is used. It makes
                                            that resource available inside a container.
                                          type: string
                                      required:
                                      - name
                                      type: object
                                    type: array
                                    x-kubernetes-list-map-keys:
                                    - name
                                    x-kubernetes-list-type: map
                                  limits:
                                    additionalProperties:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    description: 'Limits describes the maximum amount
                                      of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                    type: object
                                  requests:
                                    additionalProperties:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    description: 'Requests describes the minimum amount
                                      of compute resources required. If Requests is
                                      omitted for a container, it defaults to Limits
                                      if that is explicitly specified, otherwise to
                                      an implementation-defined value. Requests cannot
                                      exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                    type: object
                                type: object
                              storageClassName:
                                description: Type of disk to provision
                                type: string
                            required:
                            - resources
                            - storageClassName
                            type: object
                        type: object
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      resources:
                        description: ResourceRequirements describes the compute resource
                          requirements.
                        properties:
                          claims:
                            description: "Claims lists the names of resources, defined
                              in spec.resourceClaims, that are used by this container.
                              \n This is an alpha field and requires enabling the
                              DynamicResourceAllocation feature gate. \n This field
                              is immutable. It can only be set for containers."
                            items:
                              description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                              properties:
                                name:
                                  description: Name must match the name of one entry
                                    in pod.spec.resourceClaims of the Pod where this
                                    field is used. It makes that resource available
                                    inside a container.
                                  type: string
                              required:
                              - name
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - name
                            x-kubernetes-list-type: map
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Limits describes the maximum amount of compute
                              resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Requests describes the minimum amount of
                              compute resources required. If Requests is omitted for
                              a container, it defaults to Limits if that is explicitly
                              specified, otherwise to an implementation-defined value.
                              Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                        type: object
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  elasticsearchCurator:
                    properties:
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                    type: object
                type: object
                x-kubernetes-preserve-unknown-fields: true
              optimize:
                description: OptimizeSpec
                properties:
                  auth0:
                    properties:
                      audience:
                        type: string
                      backendDomain:
                        type: string
                      claimName:
                        type: string
                      clientId:
                        description: 'Deprecated: This needs a related secret and
                          comes from the env now'
                        type: string
                      clusterId:
                        description: 'Deprecated: Keep until Tasklist switches over
                          to use CAMUNDA_TASKLIST_CLOUD_CLUSTERID This is not needed
                          IMHO, as it duplicates the information'
                        type: string
                      domain:
                        type: string
                      organizationId:
                        description: 'Deprecated: Use zb.OrgID'
                        type: string
                      resourceserver:
                        description: 'Deprecated: Calculated from Domain'
                        type: string
                    type: object
                  backend:
                    description: BackendSpec contains the typical information for
                      a k8s-deployment, it can be reused when creating additional
                      application specs
                    properties:
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      overrideEnv:
                        description: Any var set here will override those provided
                          to the container. Behaviour if duplicate vars are provided
                          _here_ is undefined.
                        items:
                          description: EnvVar represents an environment variable present
                            in a Container.
                          properties:
                            name:
                              description: Name of the environment variable. Must
                                be a C_IDENTIFIER.
                              type: string
                            value:
                              description: 'Variable references $(VAR_NAME) are expanded
                                using the previously defined environment variables
                                in the container and any service environment variables.
                                If a variable cannot be resolved, the reference in
                                the input string will be unchanged. Double $$ are
                                reduced to a single $, which allows for escaping the
                                $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce
                                the string literal "$(VAR_NAME)". Escaped references
                                will never be expanded, regardless of whether the
                                variable exists or not. Defaults to "".'
                              type: string
                            valueFrom:
                              description: Source for the environment variable's value.
                                Cannot be used if value is not empty.
                              properties:
                                configMapKeyRef:
                                  description: Selects a key of a ConfigMap.
                                  properties:
                                    key:
                                      description: The key to select.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap or
                                        its key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                                fieldRef:
                                  description: 'Selects a field of the pod: supports
                                    metadata.name, metadata.namespace, metadata.labels[''<KEY>''],
                                    metadata.annotations[''<KEY>''], spec.nodeName,
                                    spec.serviceAccountName, status.hostIP, status.podIP,
                                    status.podIPs.'
                                  properties:
                                    apiVersion:
                                      description: Version of the schema the FieldPath
                                        is written in terms of, defaults to "v1".
                                      type: string
                                    fieldPath:
                                      description: Path of the field to select in
                                        the specified API version.
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                  x-kubernetes-map-type: atomic
                                resourceFieldRef:
                                  description: 'Selects a resource of the container:
                                    only resources limits and requests (limits.cpu,
                                    limits.memory, limits.ephemeral-storage, requests.cpu,
                                    requests.memory and requests.ephemeral-storage)
                                    are currently supported.'
                                  properties:
                                    containerName:
                                      description: 'Container name: required for volumes,
                                        optional for env vars'
                                      type: string
                                    divisor:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      description: Specifies the output format of
                                        the exposed resources, defaults to "1"
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    resource:
                                      description: 'Required: resource to select'
                                      type: string
                                  required:
                                  - resource
                                  type: object
                                  x-kubernetes-map-type: atomic
                                secretKeyRef:
                                  description: Selects a key of a secret in the pod's
                                    namespace
                                  properties:
                                    key:
                                      description: The key of the secret to select
                                        from.  Must be a valid secret key.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret or its
                                        key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                              type: object
                          required:
                          - name
                          type: object
                        type: array
                      resources:
                        description: ResourceRequirements describes the compute resource
                          requirements.
                        properties:
                          claims:
                            description: "Claims lists the names of resources, defined
                              in spec.resourceClaims, that are used by this container.
                              \n This is an alpha field and requires enabling the
                              DynamicResourceAllocation feature gate. \n This field
                              is immutable. It can only be set for containers."
                            items:
                              description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                              properties:
                                name:
                                  description: Name must match the name of one entry
                                    in pod.spec.resourceClaims of the Pod where this
                                    field is used. It makes that resource available
                                    inside a container.
                                  type: string
                              required:
                              - name
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - name
                            x-kubernetes-list-type: map
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Limits describes the maximum amount of compute
                              resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Requests describes the minimum amount of
                              compute resources required. If Requests is omitted for
                              a container, it defaults to Limits if that is explicitly
                              specified, otherwise to an implementation-defined value.
                              Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                        type: object
                    type: object
                  dataRetention:
                    default:
                      days: 180
                    properties:
                      days:
                        default: 180
                        maximum: 180
                        minimum: 30
                        type: integer
                    type: object
                  m2mAccounts:
                    description: M2mAuth0Spec specifies parameters to provide information
                      for a machine-to-machine connection to our accounts service
                    properties:
                      accountsURL:
                        type: string
                      audience:
                        type: string
                      clientId:
                        type: string
                      tokenUrl:
                        type: string
                    type: object
                type: object
                x-kubernetes-preserve-unknown-fields: true
              orgId:
                description: Organization ID of the cluster
                type: string
              suspend:
                description: Suspend means the cluster should not be running applications
                type: boolean
              tasklist:
                properties:
                  auth0:
                    properties:
                      audience:
                        type: string
                      backendDomain:
                        type: string
                      claimName:
                        type: string
                      clientId:
                        description: 'Deprecated: This needs a related secret and
                          comes from the env now'
                        type: string
                      clusterId:
                        description: 'Deprecated: Keep until Tasklist switches over
                          to use CAMUNDA_TASKLIST_CLOUD_CLUSTERID This is not needed
                          IMHO, as it duplicates the information'
                        type: string
                      domain:
                        type: string
                      organizationId:
                        description: 'Deprecated: Use zb.OrgID'
                        type: string
                      resourceserver:
                        description: 'Deprecated: Calculated from Domain'
                        type: string
                    type: object
                  backend:
                    description: BackendSpec contains the typical information for
                      a k8s-deployment, it can be reused when creating additional
                      application specs
                    properties:
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      overrideEnv:
                        description: Any var set here will override those provided
                          to the container. Behaviour if duplicate vars are provided
                          _here_ is undefined.
                        items:
                          description: EnvVar represents an environment variable present
                            in a Container.
                          properties:
                            name:
                              description: Name of the environment variable. Must
                                be a C_IDENTIFIER.
                              type: string
                            value:
                              description: 'Variable references $(VAR_NAME) are expanded
                                using the previously defined environment variables
                                in the container and any service environment variables.
                                If a variable cannot be resolved, the reference in
                                the input string will be unchanged. Double $$ are
                                reduced to a single $, which allows for escaping the
                                $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce
                                the string literal "$(VAR_NAME)". Escaped references
                                will never be expanded, regardless of whether the
                                variable exists or not. Defaults to "".'
                              type: string
                            valueFrom:
                              description: Source for the environment variable's value.
                                Cannot be used if value is not empty.
                              properties:
                                configMapKeyRef:
                                  description: Selects a key of a ConfigMap.
                                  properties:
                                    key:
                                      description: The key to select.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap or
                                        its key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                                fieldRef:
                                  description: 'Selects a field of the pod: supports
                                    metadata.name, metadata.namespace, metadata.labels[''<KEY>''],
                                    metadata.annotations[''<KEY>''], spec.nodeName,
                                    spec.serviceAccountName, status.hostIP, status.podIP,
                                    status.podIPs.'
                                  properties:
                                    apiVersion:
                                      description: Version of the schema the FieldPath
                                        is written in terms of, defaults to "v1".
                                      type: string
                                    fieldPath:
                                      description: Path of the field to select in
                                        the specified API version.
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                  x-kubernetes-map-type: atomic
                                resourceFieldRef:
                                  description: 'Selects a resource of the container:
                                    only resources limits and requests (limits.cpu,
                                    limits.memory, limits.ephemeral-storage, requests.cpu,
                                    requests.memory and requests.ephemeral-storage)
                                    are currently supported.'
                                  properties:
                                    containerName:
                                      description: 'Container name: required for volumes,
                                        optional for env vars'
                                      type: string
                                    divisor:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      description: Specifies the output format of
                                        the exposed resources, defaults to "1"
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    resource:
                                      description: 'Required: resource to select'
                                      type: string
                                  required:
                                  - resource
                                  type: object
                                  x-kubernetes-map-type: atomic
                                secretKeyRef:
                                  description: Selects a key of a secret in the pod's
                                    namespace
                                  properties:
                                    key:
                                      description: The key of the secret to select
                                        from.  Must be a valid secret key.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret or its
                                        key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                              type: object
                          required:
                          - name
                          type: object
                        type: array
                      resources:
                        description: ResourceRequirements describes the compute resource
                          requirements.
                        properties:
                          claims:
                            description: "Claims lists the names of resources, defined
                              in spec.resourceClaims, that are used by this container.
                              \n This is an alpha field and requires enabling the
                              DynamicResourceAllocation feature gate. \n This field
                              is immutable. It can only be set for containers."
                            items:
                              description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                              properties:
                                name:
                                  description: Name must match the name of one entry
                                    in pod.spec.resourceClaims of the Pod where this
                                    field is used. It makes that resource available
                                    inside a container.
                                  type: string
                              required:
                              - name
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - name
                            x-kubernetes-list-type: map
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Limits describes the maximum amount of compute
                              resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Requests describes the minimum amount of
                              compute resources required. If Requests is omitted for
                              a container, it defaults to Limits if that is explicitly
                              specified, otherwise to an implementation-defined value.
                              Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                        type: object
                    type: object
                  dataRetention:
                    default:
                      days: 30
                    description: DataRetentionSpec used in Operate, Tasklist.
                    properties:
                      days:
                        default: 30
                        maximum: 90
                        minimum: 5
                        type: integer
                    type: object
                type: object
                x-kubernetes-preserve-unknown-fields: true
              zeebe:
                properties:
                  broker:
                    properties:
                      clusterSize:
                        description: How many brokers to run
                        format: int32
                        maximum: 42
                        minimum: 1
                        type: integer
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      overrideEnv:
                        description: Any var set here will override those provided
                          to the broker container. Behaviour if duplicate vars are
                          provided _here_ is undefined.
                        items:
                          description: EnvVar represents an environment variable present
                            in a Container.
                          properties:
                            name:
                              description: Name of the environment variable. Must
                                be a C_IDENTIFIER.
                              type: string
                            value:
                              description: 'Variable references $(VAR_NAME) are expanded
                                using the previously defined environment variables
                                in the container and any service environment variables.
                                If a variable cannot be resolved, the reference in
                                the input string will be unchanged. Double $$ are
                                reduced to a single $, which allows for escaping the
                                $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce
                                the string literal "$(VAR_NAME)". Escaped references
                                will never be expanded, regardless of whether the
                                variable exists or not. Defaults to "".'
                              type: string
                            valueFrom:
                              description: Source for the environment variable's value.
                                Cannot be used if value is not empty.
                              properties:
                                configMapKeyRef:
                                  description: Selects a key of a ConfigMap.
                                  properties:
                                    key:
                                      description: The key to select.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap or
                                        its key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                                fieldRef:
                                  description: 'Selects a field of the pod: supports
                                    metadata.name, metadata.namespace, metadata.labels[''<KEY>''],
                                    metadata.annotations[''<KEY>''], spec.nodeName,
                                    spec.serviceAccountName, status.hostIP, status.podIP,
                                    status.podIPs.'
                                  properties:
                                    apiVersion:
                                      description: Version of the schema the FieldPath
                                        is written in terms of, defaults to "v1".
                                      type: string
                                    fieldPath:
                                      description: Path of the field to select in
                                        the specified API version.
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                  x-kubernetes-map-type: atomic
                                resourceFieldRef:
                                  description: 'Selects a resource of the container:
                                    only resources limits and requests (limits.cpu,
                                    limits.memory, limits.ephemeral-storage, requests.cpu,
                                    requests.memory and requests.ephemeral-storage)
                                    are currently supported.'
                                  properties:
                                    containerName:
                                      description: 'Container name: required for volumes,
                                        optional for env vars'
                                      type: string
                                    divisor:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      description: Specifies the output format of
                                        the exposed resources, defaults to "1"
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    resource:
                                      description: 'Required: resource to select'
                                      type: string
                                  required:
                                  - resource
                                  type: object
                                  x-kubernetes-map-type: atomic
                                secretKeyRef:
                                  description: Selects a key of a secret in the pod's
                                    namespace
                                  properties:
                                    key:
                                      description: The key of the secret to select
                                        from.  Must be a valid secret key.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret or its
                                        key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                              type: object
                          required:
                          - name
                          type: object
                        type: array
                      partitionsCount:
                        description: How many partitions to use
                        format: int32
                        maximum: 100
                        minimum: 1
                        type: integer
                      priorityClass:
                        description: PriorityClass placeholder for priority or QoS
                          implementation TODO implement this https://kubernetes.io/docs/concepts/configuration/pod-priority-preemption/#priorityclass
                        type: string
                      replicationFactor:
                        description: How many copies to keep (how many members of
                          each shard)
                        format: int32
                        maximum: 5
                        minimum: 1
                        type: integer
                      resources:
                        description: ResourceRequirements describes the compute resource
                          requirements.
                        properties:
                          claims:
                            description: "Claims lists the names of resources, defined
                              in spec.resourceClaims, that are used by this container.
                              \n This is an alpha field and requires enabling the
                              DynamicResourceAllocation feature gate. \n This field
                              is immutable. It can only be set for containers."
                            items:
                              description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                              properties:
                                name:
                                  description: Name must match the name of one entry
                                    in pod.spec.resourceClaims of the Pod where this
                                    field is used. It makes that resource available
                                    inside a container.
                                  type: string
                              required:
                              - name
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - name
                            x-kubernetes-list-type: map
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Limits describes the maximum amount of compute
                              resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Requests describes the minimum amount of
                              compute resources required. If Requests is omitted for
                              a container, it defaults to Limits if that is explicitly
                              specified, otherwise to an implementation-defined value.
                              Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                        type: object
                      storage:
                        description: StorageSpecV2 for persistent storage volumes
                          (PVCs)
                        properties:
                          autoResizing:
                            description: Configure Autoresizing
                            properties:
                              increase:
                                type: string
                              threshold:
                                type: string
                            required:
                            - increase
                            - threshold
                            type: object
                          resources:
                            description: 'Resources represents the minimum resources
                              the volume should have. More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#resources'
                            properties:
                              claims:
                                description: "Claims lists the names of resources,
                                  defined in spec.resourceClaims, that are used by
                                  this container. \n This is an alpha field and requires
                                  enabling the DynamicResourceAllocation feature gate.
                                  \n This field is immutable. It can only be set for
                                  containers."
                                items:
                                  description: ResourceClaim references one entry
                                    in PodSpec.ResourceClaims.
                                  properties:
                                    name:
                                      description: Name must match the name of one
                                        entry in pod.spec.resourceClaims of the Pod
                                        where this field is used. It makes that resource
                                        available inside a container.
                                      type: string
                                  required:
                                  - name
                                  type: object
                                type: array
                                x-kubernetes-list-map-keys:
                                - name
                                x-kubernetes-list-type: map
                              limits:
                                additionalProperties:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Limits describes the maximum amount
                                  of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                              requests:
                                additionalProperties:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Requests describes the minimum amount
                                  of compute resources required. If Requests is omitted
                                  for a container, it defaults to Limits if that is
                                  explicitly specified, otherwise to an implementation-defined
                                  value. Requests cannot exceed Limits. More info:
                                  https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                            type: object
                          storageClassName:
                            description: Type of disk to provision
                            type: string
                        required:
                        - resources
                        - storageClassName
                        type: object
                    type: object
                  config:
                    description: ZeebeConfig contains any additional configuration
                    properties:
                      whitelistIps:
                        description: 'Enterprise feature: list of ips that get added
                          to the ingress annotation "nginx.ingress.kubernetes.io/whitelist-source-range"
                          of zeebe'
                        items:
                          type: string
                        type: array
                    type: object
                  dataRetention:
                    default:
                      days: 7
                    description: ZeebeDataRetentionSpec used in Zeebe.
                    properties:
                      days:
                        default: 7
                        maximum: 30
                        minimum: 5
                        type: integer
                    type: object
                  gateway:
                    properties:
                      backend:
                        description: BackendSpec contains the typical information
                          for a k8s-deployment, it can be reused when creating additional
                          application specs
                        properties:
                          imageName:
                            description: Repository and name of the container image
                              to use
                            type: string
                          imageTag:
                            description: Tag the container image to use. Tags matching
                              /snapshot/i will use ImagePullPolicy Always
                            type: string
                          overrideEnv:
                            description: Any var set here will override those provided
                              to the container. Behaviour if duplicate vars are provided
                              _here_ is undefined.
                            items:
                              description: EnvVar represents an environment variable
                                present in a Container.
                              properties:
                                name:
                                  description: Name of the environment variable. Must
                                    be a C_IDENTIFIER.
                                  type: string
                                value:
                                  description: 'Variable references $(VAR_NAME) are
                                    expanded using the previously defined environment
                                    variables in the container and any service environment
                                    variables. If a variable cannot be resolved, the
                                    reference in the input string will be unchanged.
                                    Double $$ are reduced to a single $, which allows
                                    for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)"
                                    will produce the string literal "$(VAR_NAME)".
                                    Escaped references will never be expanded, regardless
                                    of whether the variable exists or not. Defaults
                                    to "".'
                                  type: string
                                valueFrom:
                                  description: Source for the environment variable's
                                    value. Cannot be used if value is not empty.
                                  properties:
                                    configMapKeyRef:
                                      description: Selects a key of a ConfigMap.
                                      properties:
                                        key:
                                          description: The key to select.
                                          type: string
                                        name:
                                          description: 'Name of the referent. More
                                            info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                            TODO: Add other useful fields. apiVersion,
                                            kind, uid?'
                                          type: string
                                        optional:
                                          description: Specify whether the ConfigMap
                                            or its key must be defined
                                          type: boolean
                                      required:
                                      - key
                                      type: object
                                      x-kubernetes-map-type: atomic
                                    fieldRef:
                                      description: 'Selects a field of the pod: supports
                                        metadata.name, metadata.namespace, metadata.labels[''<KEY>''],
                                        metadata.annotations[''<KEY>''], spec.nodeName,
                                        spec.serviceAccountName, status.hostIP, status.podIP,
                                        status.podIPs.'
                                      properties:
                                        apiVersion:
                                          description: Version of the schema the FieldPath
                                            is written in terms of, defaults to "v1".
                                          type: string
                                        fieldPath:
                                          description: Path of the field to select
                                            in the specified API version.
                                          type: string
                                      required:
                                      - fieldPath
                                      type: object
                                      x-kubernetes-map-type: atomic
                                    resourceFieldRef:
                                      description: 'Selects a resource of the container:
                                        only resources limits and requests (limits.cpu,
                                        limits.memory, limits.ephemeral-storage, requests.cpu,
                                        requests.memory and requests.ephemeral-storage)
                                        are currently supported.'
                                      properties:
                                        containerName:
                                          description: 'Container name: required for
                                            volumes, optional for env vars'
                                          type: string
                                        divisor:
                                          anyOf:
                                          - type: integer
                                          - type: string
                                          description: Specifies the output format
                                            of the exposed resources, defaults to
                                            "1"
                                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                          x-kubernetes-int-or-string: true
                                        resource:
                                          description: 'Required: resource to select'
                                          type: string
                                      required:
                                      - resource
                                      type: object
                                      x-kubernetes-map-type: atomic
                                    secretKeyRef:
                                      description: Selects a key of a secret in the
                                        pod's namespace
                                      properties:
                                        key:
                                          description: The key of the secret to select
                                            from.  Must be a valid secret key.
                                          type: string
                                        name:
                                          description: 'Name of the referent. More
                                            info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                            TODO: Add other useful fields. apiVersion,
                                            kind, uid?'
                                          type: string
                                        optional:
                                          description: Specify whether the Secret
                                            or its key must be defined
                                          type: boolean
                                      required:
                                      - key
                                      type: object
                                      x-kubernetes-map-type: atomic
                                  type: object
                              required:
                              - name
                              type: object
                            type: array
                          resources:
                            description: ResourceRequirements describes the compute
                              resource requirements.
                            properties:
                              claims:
                                description: "Claims lists the names of resources,
                                  defined in spec.resourceClaims, that are used by
                                  this container. \n This is an alpha field and requires
                                  enabling the DynamicResourceAllocation feature gate.
                                  \n This field is immutable. It can only be set for
                                  containers."
                                items:
                                  description: ResourceClaim references one entry
                                    in PodSpec.ResourceClaims.
                                  properties:
                                    name:
                                      description: Name must match the name of one
                                        entry in pod.spec.resourceClaims of the Pod
                                        where this field is used. It makes that resource
                                        available inside a container.
                                      type: string
                                  required:
                                  - name
                                  type: object
                                type: array
                                x-kubernetes-list-map-keys:
                                - name
                                x-kubernetes-list-type: map
                              limits:
                                additionalProperties:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Limits describes the maximum amount
                                  of compute resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                              requests:
                                additionalProperties:
                                  anyOf:
                                  - type: integer
                                  - type: string
                                  pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                  x-kubernetes-int-or-string: true
                                description: 'Requests describes the minimum amount
                                  of compute resources required. If Requests is omitted
                                  for a container, it defaults to Limits if that is
                                  explicitly specified, otherwise to an implementation-defined
                                  value. Requests cannot exceed Limits. More info:
                                  https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                                type: object
                            type: object
                        type: object
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      replicas:
                        description: How many gateways to run
                        format: int32
                        maximum: 12
                        minimum: 0
                        type: integer
                      standalone:
                        description: Per default false, which means we use an embedded
                          gateway
                        type: boolean
                    type: object
                type: object
              zeebeAnalytics:
                properties:
                  backend:
                    description: BackendSpec contains the typical information for
                      a k8s-deployment, it can be reused when creating additional
                      application specs
                    properties:
                      imageName:
                        description: Repository and name of the container image to
                          use
                        type: string
                      imageTag:
                        description: Tag the container image to use. Tags matching
                          /snapshot/i will use ImagePullPolicy Always
                        type: string
                      overrideEnv:
                        description: Any var set here will override those provided
                          to the container. Behaviour if duplicate vars are provided
                          _here_ is undefined.
                        items:
                          description: EnvVar represents an environment variable present
                            in a Container.
                          properties:
                            name:
                              description: Name of the environment variable. Must
                                be a C_IDENTIFIER.
                              type: string
                            value:
                              description: 'Variable references $(VAR_NAME) are expanded
                                using the previously defined environment variables
                                in the container and any service environment variables.
                                If a variable cannot be resolved, the reference in
                                the input string will be unchanged. Double $$ are
                                reduced to a single $, which allows for escaping the
                                $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will produce
                                the string literal "$(VAR_NAME)". Escaped references
                                will never be expanded, regardless of whether the
                                variable exists or not. Defaults to "".'
                              type: string
                            valueFrom:
                              description: Source for the environment variable's value.
                                Cannot be used if value is not empty.
                              properties:
                                configMapKeyRef:
                                  description: Selects a key of a ConfigMap.
                                  properties:
                                    key:
                                      description: The key to select.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the ConfigMap or
                                        its key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                                fieldRef:
                                  description: 'Selects a field of the pod: supports
                                    metadata.name, metadata.namespace, metadata.labels[''<KEY>''],
                                    metadata.annotations[''<KEY>''], spec.nodeName,
                                    spec.serviceAccountName, status.hostIP, status.podIP,
                                    status.podIPs.'
                                  properties:
                                    apiVersion:
                                      description: Version of the schema the FieldPath
                                        is written in terms of, defaults to "v1".
                                      type: string
                                    fieldPath:
                                      description: Path of the field to select in
                                        the specified API version.
                                      type: string
                                  required:
                                  - fieldPath
                                  type: object
                                  x-kubernetes-map-type: atomic
                                resourceFieldRef:
                                  description: 'Selects a resource of the container:
                                    only resources limits and requests (limits.cpu,
                                    limits.memory, limits.ephemeral-storage, requests.cpu,
                                    requests.memory and requests.ephemeral-storage)
                                    are currently supported.'
                                  properties:
                                    containerName:
                                      description: 'Container name: required for volumes,
                                        optional for env vars'
                                      type: string
                                    divisor:
                                      anyOf:
                                      - type: integer
                                      - type: string
                                      description: Specifies the output format of
                                        the exposed resources, defaults to "1"
                                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                      x-kubernetes-int-or-string: true
                                    resource:
                                      description: 'Required: resource to select'
                                      type: string
                                  required:
                                  - resource
                                  type: object
                                  x-kubernetes-map-type: atomic
                                secretKeyRef:
                                  description: Selects a key of a secret in the pod's
                                    namespace
                                  properties:
                                    key:
                                      description: The key of the secret to select
                                        from.  Must be a valid secret key.
                                      type: string
                                    name:
                                      description: 'Name of the referent. More info:
                                        https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                        TODO: Add other useful fields. apiVersion,
                                        kind, uid?'
                                      type: string
                                    optional:
                                      description: Specify whether the Secret or its
                                        key must be defined
                                      type: boolean
                                  required:
                                  - key
                                  type: object
                                  x-kubernetes-map-type: atomic
                              type: object
                          required:
                          - name
                          type: object
                        type: array
                      resources:
                        description: ResourceRequirements describes the compute resource
                          requirements.
                        properties:
                          claims:
                            description: "Claims lists the names of resources, defined
                              in spec.resourceClaims, that are used by this container.
                              \n This is an alpha field and requires enabling the
                              DynamicResourceAllocation feature gate. \n This field
                              is immutable. It can only be set for containers."
                            items:
                              description: ResourceClaim references one entry in PodSpec.ResourceClaims.
                              properties:
                                name:
                                  description: Name must match the name of one entry
                                    in pod.spec.resourceClaims of the Pod where this
                                    field is used. It makes that resource available
                                    inside a container.
                                  type: string
                              required:
                              - name
                              type: object
                            type: array
                            x-kubernetes-list-map-keys:
                            - name
                            x-kubernetes-list-type: map
                          limits:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Limits describes the maximum amount of compute
                              resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                          requests:
                            additionalProperties:
                              anyOf:
                              - type: integer
                              - type: string
                              pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                              x-kubernetes-int-or-string: true
                            description: 'Requests describes the minimum amount of
                              compute resources required. If Requests is omitted for
                              a container, it defaults to Limits if that is explicitly
                              specified, otherwise to an implementation-defined value.
                              Requests cannot exceed Limits. More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/'
                            type: object
                        type: object
                    type: object
                  bridgeUrl:
                    type: string
                  clusterId:
                    type: string
                  m2mAuth0:
                    properties:
                      audience:
                        type: string
                      clientId:
                        type: string
                      tokenUrl:
                        type: string
                    type: object
                  replicas:
                    description: How many analytics deployments to run
                    format: int32
                    minimum: 0
                    type: integer
                type: object
            required:
            - domain
            type: object
            x-kubernetes-preserve-unknown-fields: true
          status:
            description: ZeebeClusterStatus defines the observed state of ZeebeCluster
            properties:
              backupCount:
                type: integer
              clusterIP:
                type: string
              connectorsUrl:
                type: string
              elasticsearchStatus:
                description: HealthState gives insights whether something is going
                  on or wrong
                enum:
                - Healthy
                - Suspended
                - Unhealthy
                - Updating
                - Unknown
                type: string
              operateStatus:
                description: HealthState gives insights whether something is going
                  on or wrong
                enum:
                - Healthy
                - Suspended
                - Unhealthy
                - Updating
                - Unknown
                type: string
              operateUrl:
                type: string
              optimizeStatus:
                description: HealthState gives insights whether something is going
                  on or wrong
                enum:
                - Healthy
                - Suspended
                - Unhealthy
                - Updating
                - Unknown
                type: string
              optimizeUrl:
                type: string
              ready:
                description: Cluster has endpoint ready for traffic
                enum:
                - Healthy
                - Suspended
                - Unhealthy
                - Updating
                - Unknown
                type: string
              suspendClusterStatus:
                properties:
                  lastSuspendTime:
                    format: date-time
                    type: string
                type: object
              tasklistStatus:
                description: HealthState gives insights whether something is going
                  on or wrong
                enum:
                - Healthy
                - Suspended
                - Unhealthy
                - Updating
                - Unknown
                type: string
              tasklistUrl:
                type: string
              updateClusterStatus:
                description: UpdateClusterStatus contains the information of a cold-backup
                  during an update
                properties:
                  currentGeneration:
                    description: RelationSpec shows DB relation metadata. Most relations
                      are added as labels.
                    properties:
                      name:
                        description: Name of the corresponding relation (like generation
                          name for generations)
                        type: string
                      uuid:
                        description: UUID of the relation (like generation uuid for
                          generations)
                        type: string
                    type: object
                    x-kubernetes-preserve-unknown-fields: true
                  lastUpdatedTime:
                    format: date-time
                    type: string
                  status:
                    description: HealthState gives insights whether something is going
                      on or wrong
                    enum:
                    - Healthy
                    - Suspended
                    - Unhealthy
                    - Updating
                    - Unknown
                    type: string
                  updateState:
                    description: UpdateState Type of step in the updating process
                    enum:
                    - Done
                    - Planned
                    - ErrorInPlanned
                    - InProgress
                    - ErrorInProgress
                    - Error
                    - UpdatingResources
                    type: string
                  updatingResources:
                    items:
                      type: string
                    type: array
                type: object
              zeebeAnalyticsStatus:
                description: HealthState gives insights whether something is going
                  on or wrong
                enum:
                - Healthy
                - Suspended
                - Unhealthy
                - Updating
                - Unknown
                type: string
              zeebeAuthorityHeaderUrl:
                type: string
              zeebeStatus:
                description: HealthState gives insights whether something is going
                  on or wrong
                enum:
                - Healthy
                - Suspended
                - Unhealthy
                - Updating
                - Unknown
                type: string
              zeebeUrl:
                type: string
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}`,
)

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

var zbInstance = []byte(`
apiVersion: cloud.camunda.io/v1alpha1
kind: ZeebeCluster
metadata:
  name: zeebecluster-sample
spec:
  domain: example.com
  operate:
    alert:
      m2mAudience: cloud.example.com
      webhook: https://console.cloud.example.com/api/alert/workflow/cluster/test-1-3-2
    auth0:
      backendDomain: camunda-excitingdev.eu.auth0.com
      claimName: https://camunda.com/orgs
      domain: weblogin.cloud.example.com
    backend:
      imageName: camunda/operate
      imageTag: 8.2.2
      resources:
        limits:
          cpu: 400m
          memory: 300Mi
        requests:
          cpu: 300m
          memory: 200Mi
    elasticsearch:
      imageName: docker.elastic.co/elasticsearch/elasticsearch
      imageTag: 7.16.2
      config:
        nodesCount: 1
        storage:
          autoResizing:
            increase: 1Gi
            threshold: 20%
          resources:
            limits:
              storage: 2Gi
            requests:
              storage: 1Gi
          storageClassName: fast-v3
    elasticsearchCurator:
      imageName: bobrik/curator
      imageTag: 5.8.1
  optimize:
    auth0:
      audience: optimize.example.com
      backendDomain: camunda-excitingdev.eu.auth0.com
      claimName: https://camunda.com/orgs
      domain: weblogin.cloud.example.com
    backend:
      imageName: camunda/optimize
      imageTag: 3.7.1
      resources:
        limits:
          cpu: 500m
          memory: 300Mi
        requests:
          cpu: 500m
          memory: 300Mi
    m2mAccounts:
      accountsURL: https://accounts.cloud.example.com
      audience: cloud.example.com
      tokenUrl: https://login.cloud.example.com/oauth/token
  orgId: f4e522a8-f642-4293-b5cb-1d14e1730534
  tasklist:
    auth0:
      audience: tasklist.dev.ultrawombat.com
      backendDomain: camunda-excitingdev.eu.auth0.com
      claimName: https://camunda.com/orgs
      domain: weblogin.cloud.dev.ultrawombat.com
    backend:
      imageName: camunda/tasklist
      imageTag: 8.2.2
      resources:
        limits:
          cpu: 500m
          memory: 300Mi
        requests:
          cpu: 500m
          memory: 300Mi
  zeebe:
    broker:
      clusterSize: 1
      imageName: camunda/zeebe
      imageTag: 8.2.2
      partitionsCount: 1
      replicationFactor: 1
      resources:
        limits:
          cpu: 500m
          memory: 300Mi
        requests:
          cpu: 500m
          memory: 300Mi
      storage:
        autoResizing:
          increase: 1Gi
          threshold: 20%
        resources:
          requests:
            storage: 1Gi
        storageClassName: fast-v3
    gateway:
      replicas: 0
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
		{
			name:        "stripped zeebecluster",
			crd:         zeebecluster,
			instance:    zbInstance,
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
