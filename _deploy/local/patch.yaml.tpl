spec:
  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        url: "https://host.minikube.internal:9443/convert"
        caBundle: CA_BUNDLE
      conversionReviewVersions:
      - v1
      - v1alpha2
      - v1alpha1