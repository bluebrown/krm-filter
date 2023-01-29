package refhash

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/bluebrown/krm-filter/util"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	KindSecret                = "Secret"
	KindConfigMap             = "ConfigMap"
	KindSealedSecret          = "SealedSecret"
	KindPod                   = "Pod"
	AnnotationChecksum        = "internal.refhash/checksum"
	AnnotationPodChecksumFstr = "%s.checksum/%s"
)

var (
	DefaultSecretKinds    = []string{KindSecret, KindSealedSecret}
	DefaultConfigMapKinds = []string{KindConfigMap}
)

type Data struct {
	// kinds that can be secret ref sources
	SecretKinds util.CSV `json:"secrets_kinds,omitempty"`
	// kinds that can be configMap sources
	ConfigMapKinds util.CSV `json:"configmap_kinds,omitempty"`
}

type FunctionConfig struct {
	Data Data `json:"data,omitempty"`
}

func (config *FunctionConfig) Default() error {
	if len(config.Data.ConfigMapKinds) == 0 {
		config.Data.ConfigMapKinds = DefaultConfigMapKinds
	}
	if len(config.Data.SecretKinds) == 0 {
		config.Data.SecretKinds = DefaultSecretKinds
	}
	return nil
}

// This processor finds potential ref sources in the input node list
// and annotates pods with their checksum where applicable.
func Processor() framework.ResourceListProcessor {
	return framework.ResourceListProcessorFunc(func(rl *framework.ResourceList) error {
		config := &FunctionConfig{}
		if err := framework.LoadFunctionConfig(rl.FunctionConfig, config); err != nil {
			return err
		}

		hasher := &RefHasher{
			SecretKinds:    config.Data.SecretKinds,
			ConfigMapKinds: config.Data.ConfigMapKinds,
			annotations:    map[string]string{},
		}

		var err error
		hasher.Refs, err = getRefSources(slices.Insert(hasher.SecretKinds, 0, hasher.ConfigMapKinds...)).Filter(rl.Items)
		if err != nil {
			return err
		}

		_, err = hasher.Each(rl.Items)
		if err != nil {
			return err
		}

		rl.Results = append(rl.Results, hasher.Results...)

		for _, src := range hasher.Refs {
			annos := src.GetAnnotations()
			delete(annos, AnnotationChecksum)
			if err := src.SetAnnotations(annos); err != nil {
				return err
			}
		}

		return err
	})
}

// find potential ref sources in the given node list. Each found source
// is annotated with the checksum of the nodes content. The returned slice
// is a new nodelist containing all the found ref sources
func getRefSources(kinds []string) kio.Filter {
	return kio.FilterFunc(func(objects []*yaml.RNode) ([]*yaml.RNode, error) {
		srcs, err := framework.KindMatcher(kinds...).Filter(objects)
		if err != nil {
			return objects, err
		}

		for _, src := range srcs {
			// strip the checksum annotations
			annos := src.GetAnnotations()
			delete(annos, AnnotationChecksum)
			if err := src.SetAnnotations(annos); err != nil {
				return nil, err
			}

			// then take the checksum, w/o the checksums
			hash, err := contentSum(src)
			if err != nil {
				return objects, err
			}

			// then add back the current checksums
			annos[AnnotationChecksum] = hash
			if err := src.SetAnnotations(annos); err != nil {
				return nil, err
			}
		}

		return srcs, nil
	})
}

// the ref hasher implements a yaml filter
type RefHasher struct {
	SecretKinds    []string
	ConfigMapKinds []string
	Refs           []*yaml.RNode
	Results        framework.Results
	annotations    map[string]string
}

// the filter filters by pod and finds references to secrets and configmap
// in the pod spec. Each found reference is matched against the Refs list
// known to the hasher and if a match is found, an annotation with the ref
// sources checksum is added to the pod spec.
// This method implements the yaml.Filter interface
func (rh *RefHasher) Filter(parent *yaml.RNode) (*yaml.RNode, error) {
	ns := parent.GetNamespace()
	matchSecret := rh.matchRef(ns, rh.SecretKinds)
	matchConfMap := rh.matchRef(ns, rh.ConfigMapKinds)
	return parent.Pipe(
		// only the pod (template) will be passed to the next functions
		yaml.FilterFunc(getPod),
		// use tee to cover all the possibilities for refs on the pod
		yaml.Tee(MatchPath(PathVolumeSecrets...), matchSecret),
		yaml.Tee(MatchPath(PathVolumeConfigMaps...), matchConfMap),
		yaml.Tee(MatchPath(PathContainerEnvSecrets...), matchSecret),
		yaml.Tee(MatchPath(PathContainerEnvConfigMaps...), matchConfMap),
		yaml.Tee(MatchPath(PathContainerEnvFromSecrets...), matchSecret),
		yaml.Tee(MatchPath(PathContainerEnvFromConfigs...), matchConfMap),
		// finally, set the annotations on the pod
		yaml.FilterFunc(rh.annotatePod(parent)),
	)
}

// call the filter func for each element in the node list
// this methods implements the kio.Filter interface
func (rh *RefHasher) Each(objects []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, item := range objects {
		// clear the annotations from the previous item
		rh.annotations = make(map[string]string)
		// run the pipeline
		if err := item.PipeE(rh); err != nil {
			return nil, err
		}
	}
	return objects, nil
}

// annotate a pod based on the annotations in the refHashers annotations map
// and generate a result it its result list, using the parent as context information
func (rh *RefHasher) annotatePod(parent *yaml.RNode) func(pod *yaml.RNode) (*yaml.RNode, error) {
	return func(pod *yaml.RNode) (*yaml.RNode, error) {
		annos := pod.GetAnnotations()
		for k, v := range rh.annotations {
			if annos[k] == v {
				continue
			}
			annos[k] = v
			rh.Results = append(rh.Results, util.ItemToResult(
				parent, framework.Info, fmt.Sprintf("msg=\"set checksum on pod\" annotation=%q", k)))
		}
		return pod, pod.SetAnnotations(annos)
	}
}

// returns a filter that operates on a sequence node whichs elements are expected
// to me mapping nodes with a ref as value. The matcher is used to check if the ref value
// matches any of the ref sources known to the hasher, and if so the corresponding annotation
// is added to the hashers annotations list.
func (rh *RefHasher) matchRef(namespace string, kinds []string) yaml.Filter {
	return util.ForEach(func(node *yaml.RNode) error {
		for _, ref := range rh.Refs {
			if ref.GetNamespace() != namespace {
				continue
			}
			if !slices.Contains(kinds, ref.GetKind()) {
				continue
			}
			if ref.GetName() != node.YNode().Value {
				continue
			}
			rh.annotations[fmt.Sprintf(AnnotationPodChecksumFstr, strings.ToLower(ref.GetKind()), ref.GetName())] =
				ref.GetAnnotations()[AnnotationChecksum]
		}
		return nil
	})
}

// list of conventional pod template path
var podPath = [][]string{
	{"spec", "template"},
	{"spec", "jobTemplate", "spec", "template"},
	{"template"},
}

// get the pod from the input node. If the node is of kind pod,
// the node itself is returned. Otherwise, the pod template from
// a list of conventional path is looked up.
func getPod(item *yaml.RNode) (*yaml.RNode, error) {
	if item.GetKind() == KindPod {
		return item, nil
	}
	return yaml.LookupFirstMatch(podPath).Filter(item)
}

func MatchPath(path ...string) yaml.Filter {
	return &yaml.PathMatcher{Path: path}
}

// the matchers will always return a sequence node, meaning any filter
// after them need to anticipate the sequence and handle it accordingly
var (
	PathVolumeSecrets           = []string{"spec", "volumes", "*", "secret", "secretName"}
	PathVolumeConfigMaps        = []string{"spec", "volumes", "*", "configMap", "name"}
	PathContainerEnvSecrets     = []string{"spec", "containers", "*", "env", "*", "valueFrom", "secretKeyRef", "name"}
	PathContainerEnvConfigMaps  = []string{"spec", "containers", "*", "env", "*", "valueFrom", "configMapKeyRef", "name"}
	PathContainerEnvFromSecrets = []string{"spec", "containers", "*", "envFrom", "*", "secretRef", "name"}
	PathContainerEnvFromConfigs = []string{"spec", "containers", "*", "envFrom", "*", "configMapRef", "name"}
)

// generate a sh256 sum of the given node. The note is marshalled to json first,
// which leads to sorting the keys before taking the sum this allows to generate
// the same hash as long as the actual values of the properties have not changed
func contentSum(node *yaml.RNode) (string, error) {
	b, err := node.MarshalJSON()
	if err != nil {
		return "", err
	}
	h := sha256.New()
	if _, err := h.Write(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
