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
	var patches []model.PatchOperation

	// Perform mutations or modifications to the Deployment object
	patch := model.PatchOperation{
		Op:    "replace",
		Path:  "/spec/replicas",
		Value: 3,
	}

	patches = append(patches, patch)

	//marshal the patch into bytes
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		err := errors.New("unable to marshal patch into bytes")
		httpError(w, err)
		return
	}

	// Prepare the AdmissionResponse with the generated patch
	admissionResponse := &admissionv1.AdmissionResponse{}
	patchType := admissionv1.PatchTypeJSONPatch
	admissionResponse.Allowed = true
	admissionResponse.PatchType = &patchType
	admissionResponse.Patch = patchBytes

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
