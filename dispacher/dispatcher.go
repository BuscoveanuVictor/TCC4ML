package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	fmt.Println("--- STARTING HYBRID DISPATCHER (v1 - STRICT MODE) ---")

	// ---------------------------------------------------------
	// 1. Conectare la Clusterul LOCAL (Kind) - EXCLUSIV INTERN
	// ---------------------------------------------------------
	fmt.Println("1. Încerc conectarea la API-ul intern Kubernetes...")
	
	// Folosim DOAR configurarea internă.
	// Kubernetes injectează automat variabilele de mediu KUBERNETES_SERVICE_HOST și PORT
	localConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Sprintf("EROARE CRITICĂ: InClusterConfig a eșuat! Eroare: %v", err))
	}

	localClient, err := kubernetes.NewForConfig(localConfig)
	if err != nil {
		panic(fmt.Sprintf("Nu pot crea clientul local: %v", err))
	}
	
	// Testăm conexiunea imediat
	fmt.Println("   ... Testez permisiunile (RBAC) ...")
	_, err = localClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{Limit: 1})
	if err != nil {
		panic(fmt.Sprintf("EROARE DE PERMISIUNI (RBAC): M-am conectat, dar nu pot citi pod-uri. Ai aplicat rbac.yaml? Detalii: %v", err))
	}
	
	fmt.Println("✅ SUCCESS: Conectat la Local (Internal).")

	// ---------------------------------------------------------
	// 2. Conectare la Clusterul REMOTE (AWS EKS)
	// ---------------------------------------------------------
	cloudContextName := "cloud" // Asigură-te că e numele corect!

	fmt.Printf("2. Încerc conectarea la CLOUD (%s)...\n", cloudContextName)
	
	// Folosim client-go pentru a încărca configurația kubeconfig și a obține un client pentru EKS
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	cloudConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{CurrentContext: cloudContextName},
	).ClientConfig()

	if err != nil {
		panic(fmt.Sprintf("EROARE CRITICĂ AWS: Nu pot încărca config-ul. Eroare: %v", err))
	}

	// Testăm conexiunea imediat
	cloudClient, err := kubernetes.NewForConfig(cloudConfig)
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ SUCCESS: Conectat la Cloud (AWS).")

	// ---------------------------------------------------------
	// 3. Bucla infinită
	// ---------------------------------------------------------
	fmt.Println("🚀 Dispatcher OPERAȚIONAL. Aștept pod-uri...")
	
	for {
		// Listăm pod-urile din clusterul local pentru a detecta care trebuie migrate
		pods, err := localClient.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			fmt.Printf("⚠️ Eroare citire local (retry in 5s): %v\n", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Verificăm fiecare pod pentru a decide dacă trebuie mutat
		for _, pod := range pods.Items {
			if pod.Status.Phase == corev1.PodPending {
				if val, ok := pod.Spec.NodeSelector["zone"]; ok && val == "cloud" {
					fmt.Printf(">>> DETECTAT: Pod %s trebuie mutat în CLOUD!\n", pod.Name)
					
					if err := movePodToCloud(localClient, cloudClient, &pod); err == nil {
						fmt.Printf("✅ MIGRAT: Pod %s mutat cu succes.\n", pod.Name)
					} else {
						fmt.Printf("❌ FAIL: Eroare migrare %s: %v\n", pod.Name, err)
					}
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func movePodToCloud(local *kubernetes.Clientset, cloud *kubernetes.Clientset, pod *corev1.Pod) error {
	// Pentru a muta un pod, vom crea un nou pod în cloud și apoi vom șterge pe cel local.
	newPod := pod.DeepCopy()
	newPod.ResourceVersion = ""
	newPod.UID = ""
	newPod.Status = corev1.PodStatus{}
	newPod.Spec.NodeName = ""

	// Aici se pot adauga modificări specifice pentru cloud, 
	// dacă e necesar (ex: adăugarea de tolerations, schimbarea imaginii, etc.)
	
	_, err := cloud.CoreV1().Pods("default").Create(context.TODO(), newPod, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("AWS Create: %v", err)
	}

	err = local.CoreV1().Pods("default").Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Local Delete: %v", err)
	}
	return nil
}