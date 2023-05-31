docker-push:
	docker build . -t yogai
	docker tag yogai registry.ewintr.nl/yogai
	docker push registry.ewintr.nl/yogai
