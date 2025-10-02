# k8s-controller

see: https://www.civo.com/learn/kubernetes-admission-controllers-for-beginners

# building the image
export IMAGE_NAME=mstiehr-k8s-controller
docker build --push -t ttl.sh/${IMAGE_NAME}:1h .

# install dependencies
go mod tidy

# further reading
[operators](https://www.civo.com/learn/creating-a-kubernetes-operator-with-kubebuilder)
[ttl.sh](https://www.civo.com/learn/ttl-sh-your-anonymous-and-ephemeral-docker-image-registry)
[cert-manager](https://www.civo.com/learn/get-a-wildcard-certificate-with-cert-manager-and-civo-dns)
