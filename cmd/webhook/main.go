package main

import (
	"flag"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/pascal71/namespace-creator-webhook/pkg/webhook"
)

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var port int

	flag.StringVar(
		&metricsAddr,
		"metrics-addr",
		":8080",
		"The address the metric endpoint binds to.",
	)
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager.")
	flag.IntVar(&port, "port", 9443, "The port on which to serve webhook requests.")
	flag.Parse()

	log.SetLogger(zap.New())
	logger := log.Log.WithName("namespace-creator-webhook")

	// Get a config to talk to the apiserver
	logger.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		logger.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}

	// Create a new controller-runtime manager
	logger.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               port,
		CertDir:            "/tmp/k8s-webhook-server/serving-certs",
		LeaderElectionID:   "namespace-creator-webhook-leader",
	})
	if err != nil {
		logger.Error(err, "unable to set up manager")
		os.Exit(1)
	}

	// Setup webhook
	logger.Info("Setting up webhook server")
	hookServer := mgr.GetWebhookServer()

	logger.Info("Registering webhooks to the webhook server")
	hookServer.Register("/mutate-v1-namespace", &admission.Webhook{
		Handler: webhook.NewNamespaceCreatorWebhook(mgr.GetScheme()),
	})

	// Start the controller
	logger.Info("Starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		logger.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
