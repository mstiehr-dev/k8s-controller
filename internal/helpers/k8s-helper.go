package helpers

import (
	"encoding/json"
	"errors"
	"github.com/mstiehr-dev/k8s-controller/internal/model"
	"golang.org/x/exp/slog"
	"io"
	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
)

func parseAdmissionReview(req *http.Request, deserializer runtime.Decoder) (*admissionv1.AdmissionReview, error) {

	reqData, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("error reading request body", err)
		return nil, err
	}

	admissionReviewRequest := &admissionv1.AdmissionReview{}

	_, _, err = deserializer.Decode(reqData, nil, admissionReviewRequest)
	if err != nil {
		slog.Error("unable to desdeserialize request", err)
		return nil, err
	}
	return admissionReviewRequest, nil
}

func Mutate(w http.ResponseWriter, r *http.Request) {
	slog.Info("recieved new mutate request")

	var minReplicas int32 = 2 // TODO add configuration

	scheme := runtime.NewScheme()
	codecFactory := serializer.NewCodecFactory(scheme)
	deserializer := codecFactory.UniversalDeserializer()

	admissionReviewRequest, err := parseAdmissionReview(r, deserializer)
	if err != nil {
		httpError(w, err)
		return
	}

	// Define the GroupVersionResource for Deployment objects
	deploymentGVR := metav1.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}

	// Check if the admission request is for a Deployment object
	if admissionReviewRequest.Request.Resource != deploymentGVR {
		err := errors.New("admission request is not of kind: Deployment")
		httpError(w, err)
		return
	}

	deployment := appsv1.Deployment{}

	// Extract the Deployment object from the admission request
	_, _, err = deserializer.Decode(admissionReviewRequest.Request.Object.Raw, nil, &deployment)
	if err != nil {
		err := errors.New("unable to unmarshall request to deployment")
		httpError(w, err)
		return
	}

	admissionResponse := &admissionv1.AdmissionResponse{}
	admissionResponse.Allowed = true
	if *deployment.Spec.Replicas >= minReplicas {
		slog.Info("enough replicas - nothing to do") //, deployment.Spec.Replicas)
	} else {
		var patches []model.PatchOperation

		// Perform mutations or modifications to the Deployment object
		patchReplicas := model.PatchOperation{
			Op:    "replace",
			Path:  "/spec/replicas",
			Value: minReplicas,
		}
		patches = append(patches, patchReplicas)

		patchDeployLabels := model.PatchOperation{
			Op:    "add",
			Path:  "/metadata/labels/patched-by",
			Value: "k8s-controller",
		}
		patches = append(patches, patchDeployLabels)

		patchPodLabels := model.PatchOperation{
			Op:    "add",
			Path:  "/spec/template/metadata/labels/patched-by",
			Value: "k8s-controller",
		}
		patches = append(patches, patchPodLabels)

		//marshal the patch into bytes
		patchBytes, err := json.Marshal(patches)
		if err != nil {
			err := errors.New("unable to marshal patch into bytes")
			httpError(w, err)
			return
		}
		slog.Info("patch: ", patchBytes)
		patchType := admissionv1.PatchTypeJSONPatch
		admissionResponse.PatchType = &patchType
		admissionResponse.Patch = patchBytes
	}

	var admissionReviewResponse admissionv1.AdmissionReview
	admissionReviewResponse.Response = admissionResponse

	admissionReviewResponse.SetGroupVersionKind(admissionReviewRequest.GroupVersionKind())
	admissionReviewResponse.Response.UID = admissionReviewRequest.Request.UID

	responseBytes, err := json.Marshal(admissionReviewResponse)
	if err != nil {
		err := errors.New("unable to marshal patch response  into bytes")
		httpError(w, err)
		return
	}
	slog.Info("mutation complete", "deployment mutated", deployment.ObjectMeta.Name)
	w.Write(responseBytes)
}
