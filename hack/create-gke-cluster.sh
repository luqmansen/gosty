time gcloud beta container \
  --project "plasma-circle-320708" \
  clusters create "gosty-k8s-cluster" \
   --zone "us-central1-a" \
   --no-enable-basic-auth \
   --cluster-version "1.18.20-gke.900" \
   --release-channel "None" \
   --machine-type "e2-highcpu-2" \
   --image-type "COS_CONTAINERD" \
   --disk-type "pd-standard" \
   --disk-size "32" \
   --metadata disable-legacy-endpoints=true \
   --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" \
   --preemptible \
   --num-nodes "3" \
   --no-enable-stackdriver-kubernetes \
   --enable-ip-alias \
   --network "projects/plasma-circle-320708/global/networks/default" \
   --subnetwork "projects/plasma-circle-320708/regions/us-central1/subnetworks/default" \
   --default-max-pods-per-node "110" \
   --no-enable-master-authorized-networks \
   --addons HttpLoadBalancing,GcePersistentDiskCsiDriver \
   --enable-autoupgrade \
   --enable-autorepair \
   --max-surge-upgrade 1 \
   --max-unavailable-upgrade 0 \
   --enable-shielded-nodes \
   --node-locations "us-central1-a"


#gcloud beta container --project "gosty-311908" clusters create "gosty-k8s-cluster" --zone "asia-southeast2-a" --no-enable-basic-auth --cluster-version "1.19.9-gke.1900" --release-channel "None" --machine-type "e2-custom-8-12288" --image-type "COS_CONTAINERD" --disk-type "pd-standard" --disk-size "32" --metadata disable-legacy-endpoints=true --scopes "https://www.googleapis.com/auth/devstorage.read_only","https://www.googleapis.com/auth/logging.write","https://www.googleapis.com/auth/monitoring","https://www.googleapis.com/auth/servicecontrol","https://www.googleapis.com/auth/service.management.readonly","https://www.googleapis.com/auth/trace.append" --preemptible --num-nodes "1" --no-enable-stackdriver-kubernetes --enable-ip-alias --network "projects/gosty-311908/global/networks/default" --subnetwork "projects/gosty-311908/regions/asia-southeast2/subnetworks/default" --no-enable-intra-node-visibility --default-max-pods-per-node "110" --no-enable-master-authorized-networks --addons HorizontalPodAutoscaling,HttpLoadBalancing,GcePersistentDiskCsiDriver --enable-autoupgrade --enable-autorepair --max-surge-upgrade 1 --max-unavailable-upgrade 0 --node-locations "asia-southeast2-a"