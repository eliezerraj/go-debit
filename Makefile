.PHONY: xray

xray:
	docker run --rm \
		--env AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		--env AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		--env AWS_SESSION_TOKEN=$(AWS_SESSION_TOKEN) \
		--env AWS_REGION=us-east-2 \
		--name xray-daemon \
		--publish 2000:2000/udp \
		amazon/aws-xray-daemon -o