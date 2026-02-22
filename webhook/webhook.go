package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

// Structura pentru JSON Patch (cum modificăm Pod-ul)
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := io.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// 1. Decodează cererea de la Kubernetes
	admissionReview := admissionv1.AdmissionReview{}
	_, _, err := deserializer.Decode(body, nil, &admissionReview)
	if err != nil {
		http.Error(w, "Nu am putut decoda cererea", http.StatusBadRequest)
		return
	}

	// 2. Extrage Pod-ul din cerere
	raw := admissionReview.Request.Object.Raw
	pod := corev1.Pod{}
	if err := json.Unmarshal(raw, &pod); err != nil {
		http.Error(w, "Eroare unmarshal pod", http.StatusBadRequest)
		return
	}
	
	var patches []patchOperation
	
	// 3.Verificăm etichetele (labels) sau numele imaginii
	jobType := pod.Labels["job-type"]
	isConfidential := pod.Labels["security"] == "confidential"

	nodeSelector := make(map[string]string)

	if jobType == "spark" || isConfidential {
		// Decizie: TRIMITE PE CLOUD (EKS)
		nodeSelector["zone"] = "cloud"
		nodeSelector["security-level"] = "high-confidential"
		fmt.Printf("Orchestrare: Pod %s redirecționat către CLOUD (Confidential Computing)\n", pod.Name)
	} else {
		// Decizie: RĂMÂNE LOCAL (Kind)
		nodeSelector["zone"] = "local"
		fmt.Printf("Orchestrare: Pod %s alocat către resursă LOCALĂ\n", pod.Name)
	}

	// Creăm patch-ul JSON pentru a adăuga NodeSelector-ul la Pod
	patches = append(patches, patchOperation{
		Op:    "add",
		Path:  "/spec/nodeSelector",
		Value: nodeSelector,
	})

	patchBytes, _ := json.Marshal(patches)

	// 4. Pregătește răspunsul (AdmissionResponse)
	response := admissionv1.AdmissionResponse{
		// Permitem întotdeauna crearea pod-ului, dar cu patch-ul nostru pentru a-l redirecționa
		Allowed: true,
		// Raspunsul trebuie să conțină același UID ca cererea pentru a fi corelat
		UID:     admissionReview.Request.UID,
		// Patch-ul JSON pentru a modifica Pod-ul
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}

	admissionReview.Response = &response
	res, _ := json.Marshal(admissionReview)
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func main() {

	// Serverul nostru primeste la fiecare creare de pod 
	// o cerere la endpoint-ul /mutate (vezi in webhook-config.yaml la service) 
	// pe care o procesează in functia handleMutate
	http.HandleFunc("/mutate", handleMutate)
	
	// Fisierele TLS montate din Secret-ul Kubernetes
	// pentru a avea o conexiune securizata HTTPS
	cert := "/etc/webhook/certs/tls.crt"
	key := "/etc/webhook/certs/tls.key"

	
	fmt.Println("Serverul de orchestrare hibrida asculta pe portul 443 (HTTPS)...")
	if err := http.ListenAndServeTLS(":443", cert, key, nil); err != nil {
		fmt.Printf("Eroare start server: %v\n", err)
	}
}