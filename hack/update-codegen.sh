vendor/k8s.io/code-generator/generate-groups.sh all \
github.com/Shaad7/bookstore-sample-controller/pkg/generated \
github.com/Shaad7/bookstore-sample-controller/pkg/apis \
gopher:v1alpha1 \
--output-base "${GOPATH}/src" \
--go-header-file "hack/boilerplate.go.txt"