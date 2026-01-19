github.com/sitano/go-yaml-merge
===

Merge `*yaml.Node` together with `MergeYAMLNodes(dst, src *yaml.Node)`.

```
foo:
    a: 1
    b: 2

+

foo:
    b: 3
    c: 4

===

foo:
    a: 1
    b: 3
    c: 4
```

LICENSE
===

This project is licensed under the MIT License. Please see the LICENSE file for details..
