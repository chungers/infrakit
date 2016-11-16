package mock

//go:generate mockgen -package mock -destination ec2/mock_ec2iface.go github.com/aws/aws-sdk-go/service/ec2/ec2iface EC2API
//go:generate mockgen -package client -destination docker/docker/client/api.go github.com/docker/docker/client APIClient
