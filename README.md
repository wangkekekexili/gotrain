# gotrain

gotrain generates package dependency graph.

# Examples

`gotrain --depth=1 --format=graphviz [package] | dot -Tpng -o dependency.png`

`gotrain [package] | digraph succs [package]`