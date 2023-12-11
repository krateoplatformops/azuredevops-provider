spec:
  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        service:
          namespace: default
          name: webhook-service
          path: /convert
        caBundle: CA_BUNDLE
      conversionReviewVersions:
      - v1
      - v1alpha2
      - v1alpha1