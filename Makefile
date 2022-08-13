build:
	go generate ./... && \
	go build -o prg && \
	mkdir ~/.terraform.d/plugins/github.com/maclermo/ganeti/1.0.0/linux_amd64/ -p && \
	mv prg ~/.terraform.d/plugins/github.com/maclermo/ganeti/1.0.0/linux_amd64/terraform-provider-ganeti_v1.0.0
