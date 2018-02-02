# Avoid dynamic linking by using the pure Go net package (-tags netgo)
# Avoid dynamic linking by disabling cgo (CGO_ENABLED=0)
# Reduce binary size by omitting dwarf information (-ldflags '-w')

VER=v2
PROJECT=jsgo-192815
IMAGE=back-image
DEPLOYMENT=back-deployment

build:
	mkdir -p out
	CGO_ENABLED=0 GOOS=linux go build -o out/server-bin -a -tags netgo -ldflags '-w' .
	docker build -t gcr.io/${PROJECT}/${IMAGE}:${VER} .
	rm -rf out
	gcloud docker -- push gcr.io/${PROJECT}/${IMAGE}:${VER}

run:
	docker run --rm -p 8080:8080 -v ~/.config/:/root/.config gcr.io/${PROJECT}/${IMAGE}:${VER}
	
undeploy:
	kubectl delete deployment ${DEPLOYMENT}
	
deploy:
	kubectl run ${DEPLOYMENT} --image=gcr.io/${PROJECT}/${IMAGE}:${VER} --port 8080
	
roll:
	kubectl set image deployment/${DEPLOYMENT} ${DEPLOYMENT}=gcr.io/${PROJECT}/${IMAGE}:${VER}

show:
	kubectl get pods
	kubectl get service
	
expose:
	#kubectl expose deployment ${DEPLOYMENT} --type=LoadBalancer --port 80 --target-port 8080
	kubectl expose deployment ${DEPLOYMENT} --port 80 --target-port 8080

delete:
	# Delete the Service: This step will deallocate the Cloud Load Balancer created for your Service:
	kubectl delete service ${DEPLOYMENT}

delete-info:
	gcloud compute forwarding-rules list
	