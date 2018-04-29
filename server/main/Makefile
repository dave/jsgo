# Avoid dynamic linking by using the pure Go net package (-tags netgo)
# Avoid dynamic linking by disabling cgo (CGO_ENABLED=0)
# Reduce binary size by omitting dwarf information (-ldflags '-w')

VER=v134
CLUSTER=compile
PROJECT=jsgo-192815
IMAGE=back-image
DEPLOYMENT=back-deployment
BUCKET=pkg.jsgo.io

# Run this to do a full deploy (remember to increment VER first)
all:
	make build
	make undeploy
	make deploy
	make roll
	
scale:
	kubectl delete hpa ${DEPLOYMENT} --ignore-not-found=true
	kubectl autoscale deployment ${DEPLOYMENT} --min=1 --max=24 --cpu-percent=75

# Set the default cluster for the kebectl command
setup:
	gcloud container clusters get-credentials ${CLUSTER} --zone us-central1-a --project ${PROJECT}
	kubectl config set-cluster ${CLUSTER}

# Build dockerfile, push to Google Container Registry:
build:
	mkdir -p out
	CGO_ENABLED=0 GOOS=linux go build -o out/server-bin -a -tags="netgo norwfs" -ldflags '-w' .
	docker build -t gcr.io/${PROJECT}/${IMAGE}:${VER} .
	rm -rf out
	gcloud docker -- push gcr.io/${PROJECT}/${IMAGE}:${VER}

# Run dockerfile on localhost:8080:
run:
	docker run --rm -p 8080:8080 -v ~/.config/:/root/.config gcr.io/${PROJECT}/${IMAGE}:${VER}

# Undeploy (have to do this before deploy if it's already deployed?):
undeploy:
	kubectl delete deployment ${DEPLOYMENT}
	
# Deploy to GKE:
deploy:
	kubectl run ${DEPLOYMENT} --image=gcr.io/${PROJECT}/${IMAGE}:${VER} --port 8080

# Make this the live version:	
roll:
	kubectl set image deployment/${DEPLOYMENT} ${DEPLOYMENT}=gcr.io/${PROJECT}/${IMAGE}:${VER}

# Show stats:
show:
	kubectl get pods
	kubectl get service

# Create a load balancer exposing port 80 and use a public static IP:
expose:
	kubectl expose deployment ${DEPLOYMENT} --type=LoadBalancer --port 80 --target-port 8080

# Delete the load balancer:
delete:
	kubectl delete service ${DEPLOYMENT}

# Check here to be sure the load balancer is deleted:
delete-info:
	gcloud compute forwarding-rules list
	
create-indexes: 
	gcloud datastore create-indexes index.yaml

cleanup-indexes:
	gcloud datastore cleanup-indexes index.yaml
	
bucket:
	gsutil mb -p ${PROJECT} -c multi_regional -l us gs://${BUCKET}/
	gsutil defacl set public-read gs://${BUCKET}
	
cors:
	gsutil cors set cors-config-prod.json gs://${BUCKET}
	
cache:
	gsutil -m setmeta -h "Cache-Control:public,max-age=31536000,immutable" gs://${BUCKET}/**
