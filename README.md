# Introduction
# Run Benchmarks Locally
1. install minikube
   https://minikube.sigs.k8s.io/docs/start/?arch=%2Fmacos%2Farm64%2Fstable%2Fbinary+download
2. build images
   docker build -t otel-snitch-rb ./otel-snitch-rb
   docker build -t otel-snitch-receiver ./snitch-receiver
3. start minikub and services
   eval $(minikube docker-env) # for using local images
   minikube kubectl -- apply -f ./deploy/otel-snitch-rb.yml
   minikube kubectl -- apply -f ./deploy/otel-snitch-receiver.yml
4. login otel-snitch-rb and run benchmarks
    1. run baseline benchmark
        OTEL_SNITCH_RECEIVER_ENDPOINT=http://otel-snitch-receiver-service-service:4318/v1/traces bundle exec ruby scripts/benchmark-otel.rb

    2. run otel snitch benchmark
    
   
