# Introduction

Modern machines often have tens or even hundreds of CPU cores, and people deploy multiple services on the same machine using Kubernetes or other container orchestration systems.
In such environments, RPC requests can avoid the overhead of serialization, deserialization, memory copying, and network costs.
This project aims to explore this approach by co-designing clients and servers.

Currently, the project encodes Otel spans into shared memory, achieving more than **3 times** performance improvement when transmitting 521 Otel spans.
The next step is writing Otel spans directly to shared memory during their creation, completely avoiding encoding overhead.

<img width="593" alt="Screenshot 2024-12-20 at 09 30 22" src="https://github.com/user-attachments/assets/c4ba6caf-0e81-4eeb-b389-07b187fb88aa" />

# Perofrmance Imporvement

On the Apple M3 Pro chip, sending 512(the default max_export_batch_size) OpenTelemetry spans takes more than 10 ms of wall time and 9 ms of CPU time. The otel-snitch takes less than 3 ms for the same task, providing over **3x** performance improvement.

The next step is to write spans directly into shared memory, completely avoiding serialization costs.

# Local Benchmark
1. install the minikube
3. build the images

   ```
   docker build -t otel-snitch-rb ./otel-snitch-rb
   docker build -t otel-snitch-receiver ./snitch-receiver
   ```

4. start minikub and services
   ```
   minikube start
   eval $(minikube docker-env) # for using local images
   minikube kubectl -- apply -f ./deploy/otel-snitch-rb.yml
   minikube kubectl -- apply -f ./deploy/otel-snitch-receiver.yml
   ```
5. login to otel-snitch-rb pod and run benchmarks
    - run baseline benchmark

       `OTEL_SNITCH_RECEIVER_ENDPOINT=http://otel-snitch-receiver-service:4318/v1/traces bundle exec ruby scripts/benchmark-otel.rb`
    - run the otel-snitch benchmark

       `OTEL_SNITCH_DIR='/dev/shm' OTEL_SNITCH_RECEIVER_ENDPOINT=http://otel-snitch-receiver-service:8081/traces bundle exec ruby scripts/benchmark-otel-snitch.rb`
    
   
