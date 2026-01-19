package yaml

import (
	"bytes"
	"errors"
	"fmt"

	"go.yaml.in/yaml/v3"
)

var (
	ErrYamlUnknown          = errors.New("yaml: unknown error")
	ErrYamlUnmergable       = errors.New("yaml: unmergable error")
	ErrYamlInvalidNodeKinds = errors.New("yaml: invalid node kinds")
	ErrYamlManyDocs         = errors.New("yaml: too many documents")
)

// MergeYAMLNodes merges src into dst consuming src by:
//
// - merging mapping nodes
// - replacing seq nodes
func MergeYAMLNodes(dst, src *yaml.Node) error {
	// ''
	if dst.Kind == 0 {
		*dst = *src
		return nil
	}
	if src.Kind == 0 {
		return nil
	}

	// implicit (foo:) or explicit (foo: null) null scalars.
	if dst.Kind == yaml.ScalarNode && dst.ShortTag() == "!!null" {
		*dst = *src
		return nil
	} else if src.Kind == yaml.ScalarNode && src.ShortTag() == "!!null" {
		*dst = *src
		return nil
	}

	if dst.Kind != src.Kind {
		return ErrYamlInvalidNodeKinds
	}

	if src.HeadComment != "" {
		dst.HeadComment = src.HeadComment
	}
	if src.LineComment != "" {
		dst.LineComment = src.LineComment
	}
	if src.FootComment != "" {
		dst.FootComment = src.FootComment
	}

	switch dst.Kind {
	case yaml.DocumentNode:
		if len(dst.Content) == 0 {
			*dst = *src
			return nil
		}

		if len(src.Content) == 0 {
			return nil
		}

		if len(dst.Content) != 1 || len(src.Content) != 1 {
			return ErrYamlManyDocs
		}

		return MergeYAMLNodes(dst.Content[0], src.Content[0])
	case yaml.MappingNode:
		// We do not change node style.

		// Do not allow to change types in dst for mapping node.
		if dst.ShortTag() != src.ShortTag() {
			if src.Tag != "" {
				if dst.Tag != "" {
					return ErrYamlUnmergable
				}
				dst.Tag = src.Tag
			}
		}

		// Do not allow to break aliases in dst.
		if dst.Anchor != src.Anchor {
			if src.Anchor != "" {
				if dst.Anchor != "" {
					return ErrYamlUnmergable
				}
				dst.Anchor = src.Anchor
			}
		}

		return mergeMappingNodes(dst, src)
	case yaml.AliasNode:
		// Alias contains a pointer to an ANCHOR node in the hierarchy.
		// TODO: if it is a different anchor, we must remap the pointer to the node in dst.
		//       otherwise it may point to the nodes that were left and shadowed in src.
		*dst = *src
	case yaml.SequenceNode:
		// We do not concatenate sequence nodes.
		*dst = *src
	case yaml.ScalarNode:
		*dst = *src
	default:
		return ErrYamlUnknown
	}

	return nil
}

// mergeMappingNodes merges two mapping nodes.
func mergeMappingNodes(dst, src *yaml.Node) error {
	dstMap := mapNodeToMap(dst)

	for i := 0; i+1 < len(src.Content); i += 2 {
		key := src.Content[i]
		val := src.Content[i+1]

		if dstPair, exists := dstMap[key.Value]; exists {
			// key docs
			if err := MergeYAMLNodes(dstPair.key, key); err != nil {
				return err
			}
			if err := MergeYAMLNodes(dstPair.val, val); err != nil {
				return err
			}
		} else {
			dst.Content = append(dst.Content, key, val)
		}
	}

	return nil
}

type yamlNodeKVPair struct {
	key *yaml.Node
	val *yaml.Node
}

// nodeToMap converts a *yaml.Node of kind MappingNode to a Go map
func mapNodeToMap(node *yaml.Node) map[string]yamlNodeKVPair {
	result := make(map[string]yamlNodeKVPair)

	// Keys are at even indices, values at odd indices
	for i := 0; i+1 < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		result[key.Value] = yamlNodeKVPair{key, val}
	}

	return result
}

// FilterNullNodes deletes all null nodes from the hierarchy.
func FilterYAMLNullNodes(n *yaml.Node, explicit bool, implicit bool) {
	_ = filterYAMLNullNodes(n, explicit, implicit)
}

func filterYAMLNullNodes(n *yaml.Node, explicit bool, implicit bool) bool {
	if n.ShortTag() == "!!null" {
		return (explicit && n.Value == "null") || (implicit && n.Value == "")
	}

	switch n.Kind {
	case yaml.DocumentNode:
		fallthrough
	case yaml.SequenceNode:
		for i := len(n.Content) - 1; i >= 0; i-- {
			t := n.Content[i]
			if filterYAMLNullNodes(t, explicit, implicit) {
				n.Content = append(n.Content[:i], n.Content[i+1:]...)
			}
		}
		// instead of conversion to !!null scalar.
		if len(n.Content) == 0 {
			// doc can't be empty
			if n.Kind == yaml.DocumentNode {
				*n = yaml.Node{}
			}
			return true
		}
	case yaml.MappingNode:
		for i := len(n.Content) - 2; i >= 0; i -= 2 {
			t := n.Content[i+1] // val
			if filterYAMLNullNodes(t, explicit, implicit) {
				n.Content = append(n.Content[:i], n.Content[i+2:]...)
			}
		}
		// instead of conversion to !!null scalar.
		if len(n.Content) == 0 {
			return true
		}
	case yaml.AliasNode:
	case yaml.ScalarNode:
	default:
	}

	return false
}

type DebugYamlNode yaml.Node

func (p *DebugYamlNode) String() string {
	if p == nil {
		return ""
	}

	var buf bytes.Buffer

	_, _ = fmt.Fprintf(&buf, "%p %+v\n", p, *p)

	for _, n := range p.Content {
		_, _ = fmt.Fprint(&buf, (*DebugYamlNode)(n).String())
	}

	return buf.String()
}
