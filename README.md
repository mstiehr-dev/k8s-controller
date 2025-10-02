# k8s-controller

see: https://www.civo.com/learn/kubernetes-admission-controllers-for-beginners

# building the image
export IMAGE_NAME=mstiehr-k8s-controller
docker build --push -t ttl.sh/${IMAGE_NAME}:1h .

# install dependencies
go get golang.org/x/exp/slog
go get crypto/tls
go get k8s.io/api/admission/v1
go get k8s.io/api/apps/v1
go get k8s.io/apimachinery/pkg/apis/meta/v1
go get k8s.io/apimachinery/pkg/runtime
go get k8s.io/apimachinery/pkg/runtime/serializer
